//go:build windows
// +build windows

package etw

/*
	#cgo LDFLAGS: -ltdh

	#include "etw.h"
*/
import "C"
import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

type UserTrace struct{}

func NewUserTrace(name string, callback EventCallback) (*Trace, error) {
	return newTrace(name, callback, &UserTrace{})
}

// For User traces, no additional property is needed
func (u *UserTrace) setTraceProperties(trace *Trace) {
	return
}

func (u *UserTrace) enableProviders(trace *Trace) error {
	for _, provider := range trace.providers {
		// https://docs.microsoft.com/en-us/windows/win32/etw/configuring-and-starting-an-event-tracing-session
		params := C.ENABLE_TRACE_PARAMETERS{
			Version: 2, // ENABLE_TRACE_PARAMETERS_VERSION_2
		}
		for _, p := range provider.EnableProperties {
			params.EnableProperty |= C.ULONG(p)
		}

		// ULONG WMIAPI EnableTraceEx2(
		//	TRACEHANDLE              TraceHandle,
		//	LPCGUID                  ProviderId,
		//	ULONG                    ControlCode,
		//	UCHAR                    Level,
		//	ULONGLONG                MatchAnyKeyword,
		//	ULONGLONG                MatchAllKeyword,
		//	ULONG                    Timeout,
		//	PENABLE_TRACE_PARAMETERS EnableParameters
		// );
		//
		// Ref: https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-enabletraceex2
		ret := C.EnableTraceEx2(
			trace.registrationHandle,
			(*C.GUID)(unsafe.Pointer(&provider.ProviderId)),
			C.EVENT_CONTROL_CODE_ENABLE_PROVIDER,
			C.UCHAR(provider.Level),
			C.ULONGLONG(provider.MatchAnyKeyword),
			C.ULONGLONG(provider.MatchAllKeyword),
			0,       // Timeout set to zero to enable the trace asynchronously
			&params, //nolint:gocritic // TODO: dupSubExpr?? gocritic bug?
		)

		if status := windows.Errno(ret); status != windows.ERROR_SUCCESS {
			return fmt.Errorf("EVENT_CONTROL_CODE_ENABLE_PROVIDER failed; %w", status)
		}
	}

	return nil
}

func (u *UserTrace) disableProviders(trace *Trace) error {
	var err error

	for _, provider := range trace.providers {
		// ULONG WMIAPI EnableTraceEx2(
		//	TRACEHANDLE              TraceHandle,
		//	LPCGUID                  ProviderId,
		//	ULONG                    ControlCode,
		//	UCHAR                    Level,
		//	ULONGLONG                MatchAnyKeyword,
		//	ULONGLONG                MatchAllKeyword,
		//	ULONG                    Timeout,
		//	PENABLE_TRACE_PARAMETERS EnableParameters
		// );
		ret := C.EnableTraceEx2(
			trace.registrationHandle,
			(*C.GUID)(unsafe.Pointer(&provider.ProviderId)),
			C.EVENT_CONTROL_CODE_DISABLE_PROVIDER,
			0,
			0,
			0,
			0,
			nil)

		if status := windows.Errno(ret); status != windows.ERROR_SUCCESS && status != windows.ERROR_NOT_FOUND {
			err = status
		}
	}

	return err
}
