//go:build windows
// +build windows

package etw

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestKernelTrace(t *testing.T) {
	suite.Run(t, new(kernelTraceSuite))
}

type kernelTraceSuite struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc
}

func (s *kernelTraceSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())
}

func (s *kernelTraceSuite) TearDownTest() {
	s.cancel()
}

func (s *kernelTraceSuite) TestCollectKernelEvents() {
	const deadline = 5 * time.Second

	// Callback will signal about seen events
	// We'll create a big channel to make sure the callback is never blocked
	var collectedEvents = make(chan map[string]interface{}, 10000)
	defer close(collectedEvents)

	cb := func(e *Event) {
		properties, err := e.EventProperties()
		s.Require().NoError(err, "Got error parsing event properties")
		collectedEvents <- properties
	}

	// Try to process kernel process events
	done := make(chan struct{})
	trace, err := NewKernelTrace("Test-ETW", cb)
	s.Require().NoError(err, "Error creating trace object")

	trace.Enable(KERNEL_PROCESS_PROVIDER)

	go func() {
		s.Require().NoError(trace.Start(), "Error processing events")
		close(done)
	}()

	// Execute windows command that will generate events
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				exec.Command("net", "user", os.ExpandEnv("$Username")).Run()
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Make sure we are getting the event we are looking for
	timeout := time.After(deadline)
waitForEvent:
	for {
		select {
		case event := <-collectedEvents:
			if event["CommandLine"] == os.ExpandEnv("net user $Username") {
				break waitForEvent
			}
		case <-timeout:
			s.Fail("Failed to get expected event from Kernel Trace")
			break waitForEvent
		}
	}

	// Stop the session and ensure that processing goroutine will also stop
	s.Require().NoError(trace.Stop(), "Failed to close trace properly")
	s.waitForSignal(done, deadline, "Failed to stop event processing")
}

// waitForSignal waits for anything on @done no longer than @deadline.
// Fails test run if deadline exceeds.
func (s kernelTraceSuite) waitForSignal(done <-chan struct{}, deadline time.Duration, failMsg string) {
	select {
	case <-done:
		// pass.
	case <-time.After(deadline):
		s.Fail(failMsg, "deadline %s exceeded", deadline)
	}
}
