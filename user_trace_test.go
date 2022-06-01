//go:build windows
// +build windows

package etw

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	msetw "github.com/Microsoft/go-winio/pkg/etw"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sys/windows"
)

func TestUserTrace(t *testing.T) {
	suite.Run(t, new(userTraceSuite))
}

type userTraceSuite struct {
	suite.Suite

	ctx      context.Context
	cancel   context.CancelFunc
	provider *msetw.Provider
	guid     windows.GUID
}

func (s *userTraceSuite) SetupTest() {
	provider, err := msetw.NewProvider("TestProvider", nil)
	s.Require().NoError(err, "Failed to initialize test provider.")

	s.provider = provider
	s.guid = windows.GUID(provider.ID)

	s.ctx, s.cancel = context.WithCancel(context.Background())
}

func (s *userTraceSuite) TearDownTest() {
	s.cancel()
	s.Require().NoError(s.provider.Close(), "Failed to close test provider.")
}

// TestSmoke ensures that etw.UserTrace is working as expected: it could start, process incoming
// events and stop properly.
func (s *userTraceSuite) TestSmoke() {
	const deadline = 10 * time.Second

	// Spam some events to emulate a normal ETW provider behaviour.
	go s.generateEvents(s.ctx, s.provider, []msetw.Level{msetw.LevelInfo})

	// The only thing we are going to do is signal that we've got something.
	gotEvent := make(chan struct{})
	defer close(gotEvent)
	cb := func(_ *Event) {
		trySignal(gotEvent)
	}

	// Ensure we can subscribe to our in-house ETW provider.
	trace, err := NewUserTrace("Test-ETW", cb)
	s.Require().NoError(err, "Failed to create trace")

	trace.Enable(NewProvider(s.guid))

	// Start the processing routine. We expect the routine will stop on `trace.Stop()`.
	done := make(chan struct{})
	go func() {
		s.Require().NoError(trace.Start(), "Error processing events")
		close(done)
	}()

	// Ensure that we are able to receive events from the provider. An ability
	// to get the proper content is tested in TestParsing.
	s.waitForSignal(gotEvent, deadline, "Failed to receive event from provider")

	// Now stop the session and ensure that processing goroutine will also stop.
	s.Require().NoError(trace.Stop(), "Failed to close session properly")
	s.waitForSignal(done, deadline, "Failed to stop event processing")
}

// TestParsing ensures that etw.Trace is able to parse events with all common field types.
func (s *userTraceSuite) TestParsing() {
	const deadline = 20 * time.Second

	go s.generateEvents(
		s.ctx,
		s.provider,
		[]msetw.Level{msetw.LevelInfo},
		msetw.StringField("string", "string value"),
		msetw.StringArray("stringArray", []string{"1", "2", "3"}),
		msetw.Float64Field("float64", 45.7),
		msetw.Struct("struct",
			msetw.StringField("string", "string value"),
			msetw.Float64Field("float64", 46.7),
			msetw.Struct("subStructure",
				msetw.StringField("string", "string value"),
			),
		),
		msetw.StringArray("anotherArray", []string{"3", "4"}),
	)
	expectedMap := map[string]interface{}{
		"string":            "string value",
		"stringArray.Count": "3", // OS artifacts
		"stringArray":       []interface{}{"1", "2", "3"},
		"float64":           "45.700000",
		"struct": map[string]interface{}{
			"string": "string value",

			"float64": "46.700000",
			"subStructure": map[string]interface{}{
				"string": "string value",
			},
		},
		"anotherArray.Count": "2", // OS artifacts
		"anotherArray":       []interface{}{"3", "4"},
	}

	var (
		properties map[string]interface{}
		gotProps   = make(chan struct{}, 1)
		err        error
	)
	cb := func(e *Event) {
		properties, err = e.EventProperties()
		s.Require().NoError(err, "Got error parsing event properties")
		trySignal(gotProps)
	}

	trace, err := NewUserTrace("Test-ETW", cb)
	s.Require().NoError(err, "Failed to create a trace")

	trace.Enable(NewProvider(s.guid))

	done := make(chan struct{})
	go func() {
		s.Require().NoError(trace.Start(), "Error processing events")
		close(done)
	}()

	s.waitForSignal(gotProps, deadline, "Failed to get event")
	s.Equal(expectedMap, properties, "Received unexpected properties")

	s.Require().NoError(trace.Stop(), "Failed to close session properly")
	s.waitForSignal(done, deadline, "Failed to stop event processing")
}

// TestEventOutsideCallback ensures *Event can't be used outside EventCallback.
func (s *userTraceSuite) TestEventOutsideCallback() {
	const deadline = 10 * time.Second
	go s.generateEvents(s.ctx, s.provider, []msetw.Level{msetw.LevelInfo})

	// Grab event pointer from the callback. We expect that outdated pointer
	// will protect user from calling Windows API on freed memory.
	var evt *Event
	gotEvent := make(chan struct{})
	cb := func(e *Event) {
		// Signal on second event only to guarantee that callback with stored event will finish.
		if evt != nil {
			trySignal(gotEvent)
		} else {
			evt = e
		}
	}

	trace, err := NewUserTrace("Test-ETW", cb)
	s.Require().NoError(err, "Failed to create session")

	trace.Enable(NewProvider(s.guid))

	done := make(chan struct{})
	go func() {
		s.Require().NoError(trace.Start(), "Error processing events")
		close(done)
	}()

	// Wait for event arrived and try to access event data.
	s.waitForSignal(gotEvent, deadline, "Failed to receive event from provider")
	s.Assert().Zero(evt.ExtendedInfo(), "Got non-nil ExtendedInfo for freed event")
	_, err = evt.EventProperties()
	s.Assert().Error(err, "Don't get an error using freed event")
	s.Assert().Contains(err.Error(), "EventCallback", "Got unexpected error: %s", err)

	s.Require().NoError(trace.Stop(), "Failed to close session properly")
	s.waitForSignal(done, deadline, "Failed to stop event processing")
}

// TestMultipleProviders ensures we can subscribe to several providers in the same trace session
func (s *userTraceSuite) TestMultipleProviers() {
	const deadline = 2 * time.Second

	// Create a provider that will spam INFO events.
	go s.generateEvents(s.ctx, s.provider, []msetw.Level{msetw.LevelInfo})

	// Create a second provider that will spam CRITICAL events.
	secondProvider, err := msetw.NewProvider("SecondTestProvider", nil)
	s.Require().NoError(err, "Failed to initialize test provider.")
	go s.generateEvents(s.ctx, secondProvider, []msetw.Level{msetw.LevelCritical})

	// Callback will signal about seen event level through corresponding channels.
	var (
		gotCriticalEvent    = make(chan struct{}, 1)
		gotInformationEvent = make(chan struct{}, 1)
	)
	cb := func(e *Event) {
		switch TraceLevel(e.Header.Level) {
		case TRACE_LEVEL_INFORMATION:
			trySignal(gotInformationEvent)
		case TRACE_LEVEL_CRITICAL:
			trySignal(gotCriticalEvent)
		}
	}

	// Then subscribe to the both producers
	trace, err := NewUserTrace("Test-ETW", cb)
	s.Require().NoError(err, "Failed to create session")

	trace.Enable(NewProvider(s.guid))
	trace.Enable(NewProvider(windows.GUID(secondProvider.ID)))

	done := make(chan struct{})
	go func() {
		s.Require().NoError(trace.Start(), "Error processing events")
		close(done)
	}()

	// Ensure that we are getting INFO and CRITICAL events
	s.waitForSignal(gotInformationEvent, deadline, "Failed to get event from the first provider")
	s.waitForSignal(gotCriticalEvent, deadline,
		"Failed to receive event from the second provider")

	// Stop the session and ensure that processing goroutine will also stop.
	s.Require().NoError(trace.Stop(), "Failed to close session properly")
	s.waitForSignal(done, deadline, "Failed to stop event processing")
}

func (s *userTraceSuite) TestExistingTrace() {
	const deadline = 5 * time.Second

	// Callback will signal about seen event
	var gotEvent = make(chan struct{}, 1)
	cb := func(e *Event) {
		trySignal(gotEvent)
	}

	// Try to process Security Events
	done := make(chan struct{})
	trace, err := NewUserTrace("Eventlog-Security", cb)
	s.Require().NoError(err, "Error creating trace object")

	s.Require().NoError(trace.OpenTrace(), "Error opening the trace")

	go func() {
		s.Require().NoError(trace.Process(), "Error processing events")
		close(done)
	}()

	// Execute windows command that will generate events
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				exec.Command("net", "user", os.ExpandEnv("Username")).Run()
			}
		}
	}()

	// Ensure that we are getting events
	s.waitForSignal(gotEvent, deadline, "Failed to get event from Security Auditing")

	// Stop the session and ensure that processing goroutine will also stop
	s.Require().NoError(trace.Stop(), "Failed to close session properly")
	s.waitForSignal(done, deadline, "Failed to stop event processing")
}

// TestStopExisting ensures that we are able to force kill the lost session using only
// its name.
func (s *userTraceSuite) TestStopExisting() {
	sessionName := fmt.Sprintf("go-etw-suicide-%d", time.Now().UnixNano())

	// Ensure we can create a session with a given name.
	trace, _ := NewUserTrace(sessionName, nil)
	s.Require().NoError(trace.Open(), "Failed to create session with name %s", sessionName)

	// Ensure we've got ExistsError creating a session with the same name.
	trace, _ = NewUserTrace(sessionName, nil)
	err := trace.Open()
	s.Require().Error(err)

	var exists ExistsError
	s.Require().True(errors.As(err, &exists), "Got unexpected error starting session with a same name")

	// Try to force-kill the session by name.
	s.Require().NoError(trace.Kill(), "Failed to force stop session")

	// Ensure that fresh session could normally started and stopped.
	s.Require().NoError(trace.Open(), "Failed to create session after a successful kill")
	s.Require().NoError(trace.Stop(), "Failed to close session properly")
}

// trySignal tries to send a signal to @done if it's ready to receive.
// @done expected to be a buffered channel.
func trySignal(done chan<- struct{}) {
	select {
	case done <- struct{}{}:
	default:
	}
}

// waitForSignal waits for anything on @done no longer than @deadline.
// Fails test run if deadline exceeds.
func (s userTraceSuite) waitForSignal(done <-chan struct{}, deadline time.Duration, failMsg string) {
	select {
	case <-done:
		// pass.
	case <-time.After(deadline):
		s.Fail(failMsg, "deadline %s exceeded", deadline)
	}
}

// We have no easy way to ensure that etw session is started and ready to process events,
// so it seems easier to just flood an events and catch some of them than try to catch
// the actual session readiness and sent the only one.
func (s userTraceSuite) generateEvents(ctx context.Context, provider *msetw.Provider, levels []msetw.Level, fields ...msetw.FieldOpt) {
	// If nothing provided, receiver doesn't care about the event content -- send anything.
	if fields == nil {
		fields = msetw.WithFields(msetw.StringField("TestField", "Foo"))
	}
	s.Require().NotEmpty(levels, "Incorrect generateEvents usage")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			for _, l := range levels {
				_ = provider.WriteEvent(
					"TestEvent",
					msetw.WithEventOpts(msetw.WithLevel(l)),
					fields,
				)
			}
		}
	}
}
