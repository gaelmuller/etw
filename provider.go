//go:build windows
// +build windows

package etw

/*
	#include "windows.h"
*/
import "C"
import "golang.org/x/sys/windows"

// Provider represents the trace configuration associated with a provider
type Provider struct {
	// The provider ID (control GUID) of the event provider that you want to configure.
	ProviderId windows.GUID

	// KernelTrace Providers
	EnableFlags uint64

	// UserTrace Providers

	// Level represents provider-defined value that specifies the level of
	// detail included in the event. Higher levels imply that you get lower
	// levels as well. For example, with TRACE_LEVEL_ERROR you'll get all
	// events except ones with level critical. Check `EventDescriptor.Level`
	// values for current event verbosity level.
	Level TraceLevel

	// MatchAnyKeyword is a bitmask of keywords that determine the category of
	// events that you want the provider to write. The provider writes the
	// event if any of the event's keyword bits match any of the bits set in
	// this mask.
	//
	// If MatchAnyKeyword is not set the session will receive ALL possible
	// events (which is equivalent setting all 64 bits to 1).
	//
	// Passed as is to EnableTraceEx2. Refer to its remarks for more info:
	// https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-enabletraceex2#remarks
	MatchAnyKeyword uint64

	// MatchAllKeyword is an optional bitmask that further restricts the
	// category of events that you want the provider to write. If the event's
	// keyword meets the MatchAnyKeyword condition, the provider will write the
	// event only if all of the bits in this mask exist in the event's keyword.
	//
	// This mask is not used if MatchAnyKeyword is zero.
	//
	// Passed as is to EnableTraceEx2. Refer to its remarks for more info:
	// https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-enabletraceex2#remarks
	MatchAllKeyword uint64

	// EnableProperties defines a set of provider properties consumer wants to
	// enable. Properties adds fields to ExtendedEventInfo or asks provider to
	// sent more events.
	//
	// For more info about available properties check EnableProperty doc and
	// original API reference:
	// https://docs.microsoft.com/en-us/windows/win32/api/evntrace/ns-evntrace-enable_trace_parameters
	EnableProperties []EnableProperty
}

// TraceLevel represents provider-defined value that specifies the level of
// detail included in the event. Higher levels imply that you get lower
// levels as well.
type TraceLevel C.UCHAR

//nolint:golint,stylecheck // We keep original names to underline that it's an external constants.
const (
	TRACE_LEVEL_CRITICAL    = TraceLevel(1)
	TRACE_LEVEL_ERROR       = TraceLevel(2)
	TRACE_LEVEL_WARNING     = TraceLevel(3)
	TRACE_LEVEL_INFORMATION = TraceLevel(4)
	TRACE_LEVEL_VERBOSE     = TraceLevel(5)
)

// EnableProperty enables a property of a provider session is subscribing for.
//
// For more info about available properties check original API reference:
// https://docs.microsoft.com/en-us/windows/win32/api/evntrace/ns-evntrace-enable_trace_parameters
type EnableProperty C.ULONG

//nolint:golint,stylecheck // We keep original names to underline that it's an external constants.
const (
	// Include in the ExtendedEventInfo the security identifier (SID) of the user.
	EVENT_ENABLE_PROPERTY_SID = EnableProperty(0x001)

	// Include in the ExtendedEventInfo the terminal session identifier.
	EVENT_ENABLE_PROPERTY_TS_ID = EnableProperty(0x002)

	// Include in the ExtendedEventInfo a call stack trace for events written
	// using EventWrite.
	EVENT_ENABLE_PROPERTY_STACK_TRACE = EnableProperty(0x004)

	// Filters out all events that do not have a non-zero keyword specified.
	// By default events with 0 keywords are accepted.
	EVENT_ENABLE_PROPERTY_IGNORE_KEYWORD_0 = EnableProperty(0x010)

	// Filters out all events that are either marked as an InPrivate event or
	// come from a process that is marked as InPrivate. InPrivate implies that
	// the event or process contains some data that would be considered private
	// or personal. It is up to the process or event to designate itself as
	// InPrivate for this to work.
	EVENT_ENABLE_PROPERTY_EXCLUDE_INPRIVATE = EnableProperty(0x200)
)

func NewProvider(id windows.GUID) *Provider {
	return &Provider{ProviderId: id, Level: TRACE_LEVEL_VERBOSE}
}

func NewKernelProvider(id windows.GUID, flags uint64) *Provider {
	return &Provider{ProviderId: id, EnableFlags: flags}
}
