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
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

type traceImplementation interface {
	setTraceProperties(trace *Trace)
	enableProviders(trace *Trace) error
	disableProviders(trace *Trace) error
}

// ExistsError is returned by trace.Open() if the session name is already taken.
//
// Having ExistsError you have an option to force kill the session.
type ExistsError struct{}

func (e ExistsError) Error() string {
	return "session already exist"
}

// EventCallback is any function that could handle an ETW event. EventCallback
// is called synchronously and sequentially on every event received by Session
// one by one.
//
// If EventCallback can't handle all ETW events produced, OS will handle a
// tricky file-based cache for you, however, it's recommended not to perform
// long-running tasks inside a callback.
//
// N.B. Event pointer @e is valid ONLY inside a callback. You CAN'T copy a
// whole event, only EventHeader, EventProperties and ExtendedEventInfo
// separately.
type EventCallback func(e *Event)

type Trace struct {
	name      []uint16
	providers map[windows.GUID]*Provider

	registrationHandle C.TRACEHANDLE
	sessionHandle      C.TRACEHANDLE

	properties C.PEVENT_TRACE_PROPERTIES

	callback EventCallback
	cgoKey   uintptr

	impl traceImplementation
}

func newTrace(name string, callback EventCallback, impl traceImplementation) (*Trace, error) {
	// Convert the name to UTF-16
	utf16Name, err := windows.UTF16FromString(name)
	if err != nil {
		return nil, fmt.Errorf("incorrect session name; %w", err) // unlikely
	}

	return &Trace{
		name:               utf16Name,
		providers:          make(map[windows.GUID]*Provider),
		registrationHandle: C.INVALID_PROCESSTRACE_HANDLE,
		sessionHandle:      C.INVALID_PROCESSTRACE_HANDLE,
		properties:         newTraceProperties(utf16Name),
		callback:           callback,
		impl:               impl,
	}, nil
}

func newTraceProperties(name []uint16) C.PEVENT_TRACE_PROPERTIES {
	// We need to allocate a sequential buffer for a structure and a session name
	// which will be placed there by an API call (for the future calls).
	//
	// (Ref: https://docs.microsoft.com/en-us/windows/win32/etw/wnode-header#members)
	//
	// The only way to do it in go -- unsafe cast of the allocated memory.
	sessionNameSize := len(name) * int(unsafe.Sizeof(name[0]))
	bufSize := int(unsafe.Sizeof(C.EVENT_TRACE_PROPERTIES{})) + sessionNameSize
	propertiesBuf := make([]byte, bufSize)

	// We will use Query Performance Counter for timestamp cos it gives us higher
	// time resolution. Event timestamps however would be converted to the common
	// FileTime due to absence of PROCESS_TRACE_MODE_RAW_TIMESTAMP in LogFileMode.
	//
	// Ref: https://docs.microsoft.com/en-us/windows/win32/api/evntrace/ns-evntrace-event_trace_properties
	pProperties := (C.PEVENT_TRACE_PROPERTIES)(unsafe.Pointer(&propertiesBuf[0]))
	pProperties.Wnode.BufferSize = C.ulong(bufSize)
	pProperties.Wnode.ClientContext = 1 // QPC for event Timestamp
	pProperties.Wnode.Flags = C.WNODE_FLAG_TRACED_GUID

	// Mark that we are going to process events in real time using a callback.
	pProperties.LogFileMode = C.EVENT_TRACE_REAL_TIME_MODE | C.EVENT_TRACE_NO_PER_PROCESSOR_BUFFERING

	return pProperties
}

func (trace *Trace) Enable(provider *Provider) {
	if provider != nil {
		trace.providers[provider.ProviderId] = provider
	}
}

func (trace *Trace) Start() error {
	if trace.sessionHandle == C.INVALID_PROCESSTRACE_HANDLE {
		if err := trace.Open(); err != nil {
			return err
		}
	}

	return trace.processTrace()
}

func (trace *Trace) Stop() error {
	trace.stopTrace()
	return trace.closeTrace()
}

func (trace *Trace) OpenTrace() error {
	trace.cgoKey = newCallbackKey(trace)

	// Ref: https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-opentracew
	trace.sessionHandle = C.OpenTraceHelper(
		(C.LPWSTR)(unsafe.Pointer(&trace.name[0])),
		(C.PVOID)(trace.cgoKey),
	)
	if trace.sessionHandle == C.INVALID_PROCESSTRACE_HANDLE {
		return fmt.Errorf("OpenTraceW failed; %w", windows.GetLastError())
	}

	return nil
}

func (trace *Trace) Open() error {
	if err := trace.registerTrace(); err != nil {
		return err
	}

	trace.impl.enableProviders(trace)

	return trace.OpenTrace()
}

func (trace *Trace) Process() error {
	return trace.processTrace()
}

// Kill forces the trace with a given @name to stop. Don't having a
// trace handle we can't shutdown it gracefully unsubscribing from all the
// providers first, so we just stop the trace itself.
//
// Use Kill only to destroy session you've lost control over. If you
// have a session handle always prefer `.Stop`.
func (trace *Trace) Kill() error {
	sessionNameLength := len(trace.name) * int(unsafe.Sizeof(trace.name[0]))

	// We don't know if this session was opened with the log file or not
	// (session could be opened without our library) so allocate memory for LogFile name too.
	const maxLengthLogfileName = 1024
	bufSize := int(unsafe.Sizeof(C.EVENT_TRACE_PROPERTIES{})) + sessionNameLength + maxLengthLogfileName
	propertiesBuf := make([]byte, bufSize)
	pProperties := (C.PEVENT_TRACE_PROPERTIES)(unsafe.Pointer(&propertiesBuf[0]))
	pProperties.Wnode.BufferSize = C.ulong(bufSize)

	// ULONG WMIAPI ControlTraceW(
	//  TRACEHANDLE             TraceHandle,
	//  LPCWSTR                 InstanceName,
	//  PEVENT_TRACE_PROPERTIES Properties,
	//  ULONG                   ControlCode
	// );
	ret := C.ControlTraceW(
		0,
		(*C.ushort)(unsafe.Pointer(&trace.name[0])),
		pProperties,
		C.EVENT_TRACE_CONTROL_STOP)

	// If you receive ERROR_MORE_DATA when stopping the session, ETW will have
	// already stopped the session before generating this error.
	// https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-controltracew
	switch status := windows.Errno(ret); status {
	case windows.ERROR_MORE_DATA, windows.ERROR_SUCCESS:
		return nil
	default:
		return status
	}
}

func (trace *Trace) registerTrace() error {
	// Because common logic is already present in default properties
	// We only need to set properties specific to an implementation here
	trace.impl.setTraceProperties(trace)

	// Try to start the trace
	ret := C.StartTraceW(
		&trace.registrationHandle,
		C.LPWSTR(unsafe.Pointer(&trace.name[0])),
		trace.properties,
	)
	switch err := windows.Errno(ret); err {
	// If it already exists, try to stop it before retrying
	case windows.ERROR_ALREADY_EXISTS:
		// Invalidate the registrationHandle
		trace.registrationHandle = C.INVALID_PROCESSTRACE_HANDLE
		return ExistsError{}
	case windows.ERROR_SUCCESS:
		return nil
	default:
		return fmt.Errorf("StartTraceW failed; %w", err)
	}
}

func (trace *Trace) processTrace() error {
	defer freeCallbackKey(trace.cgoKey)

	// BLOCKS UNTIL CLOSED!
	//
	// Ref: https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-processtrace
	// ETW_APP_DECLSPEC_DEPRECATED ULONG WMIAPI ProcessTrace(
	// 	PTRACEHANDLE HandleArray,
	// 	ULONG        HandleCount,
	// 	LPFILETIME   StartTime,
	// 	LPFILETIME   EndTime
	// );
	ret := C.ProcessTrace(
		C.PTRACEHANDLE(&trace.sessionHandle),
		1,   // ^ Imagine we pass an array with 1 element here.
		nil, // Do not want to limit StartTime (default is from now).
		nil, // Do not want to limit EndTime.
	)
	switch status := windows.Errno(ret); status {
	case windows.ERROR_SUCCESS, windows.ERROR_CANCELLED:
		return nil // Cancelled is obviously ok when we block until closing.
	default:
		return fmt.Errorf("ProcessTrace failed; %w", status)
	}
}

func (trace *Trace) stopTrace() error {
	if trace.registrationHandle != C.INVALID_PROCESSTRACE_HANDLE {
		trace.impl.disableProviders(trace)

		// ULONG WMIAPI ControlTraceW(
		//  TRACEHANDLE             TraceHandle,
		//  LPCWSTR                 InstanceName,
		//  PEVENT_TRACE_PROPERTIES Properties,
		//  ULONG                   ControlCode
		// );
		ret := C.ControlTraceW(
			trace.registrationHandle,
			nil,
			trace.properties,
			C.EVENT_TRACE_CONTROL_STOP)

		// If you receive ERROR_MORE_DATA when stopping the session, ETW will have
		// already stopped the session before generating this error.
		// https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-controltracew
		switch status := windows.Errno(ret); status {
		case windows.ERROR_MORE_DATA, windows.ERROR_SUCCESS:
			return nil
		default:
			return status
		}
	}

	return nil
}

func (trace *Trace) closeTrace() error {
	if trace.sessionHandle != C.INVALID_PROCESSTRACE_HANDLE {
		// ETW_APP_DECLSPEC_DEPRECATED ULONG WMIAPI CloseTrace(
		//	[in] TRACEHANDLE TraceHandle
		// );
		ret := C.CloseTrace(trace.sessionHandle)
		switch status := windows.Errno(ret); status {
		case windows.ERROR_SUCCESS, windows.ERROR_CTX_CLOSE_PENDING:
			return nil
		default:
			return fmt.Errorf("CloseTrace failed: %w", status)
		}
	}

	return nil
}

// We can't pass Go-land pointers to the C-world so we use a classical trick
// storing real pointers inside global map and passing to C "fake pointers"
// which are actually map keys.
//
//nolint:gochecknoglobals
var (
	traces        sync.Map
	tracesCounter uintptr
)

// newCallbackKey stores a @ptr inside a global storage returning its' key.
// After use the key should be freed using `freeCallbackKey`.
func newCallbackKey(trace *Trace) uintptr {
	key := atomic.AddUintptr(&tracesCounter, 1)
	traces.Store(key, trace)

	return key
}

func freeCallbackKey(key uintptr) {
	traces.Delete(key)
}

// handleEvent is exported to guarantee C calling convention (cdecl).
//
// The function should be defined here but would be linked and used inside
// C code in `session.c`.
//
//export handleEvent
func handleEvent(eventRecord C.PEVENT_RECORD) {
	key := uintptr(eventRecord.UserContext)
	targetTrace, ok := traces.Load(key)
	if !ok {
		return
	}

	evt := &Event{
		Header:      eventHeaderToGo(eventRecord.EventHeader),
		eventRecord: eventRecord,
	}
	targetTrace.(*Trace).callback(evt)
	evt.eventRecord = nil
}

func eventHeaderToGo(header C.EVENT_HEADER) EventHeader {
	return EventHeader{
		EventDescriptor: eventDescriptorToGo(header.EventDescriptor),
		ThreadID:        uint32(header.ThreadId),
		ProcessID:       uint32(header.ProcessId),
		TimeStamp:       stampToTime(C.GetTimeStamp(header)),
		ProviderID:      windowsGUIDToGo(header.ProviderId),
		ActivityID:      windowsGUIDToGo(header.ActivityId),

		Flags:         uint16(header.Flags),
		KernelTime:    uint32(C.GetKernelTime(header)),
		UserTime:      uint32(C.GetUserTime(header)),
		ProcessorTime: uint64(C.GetProcessorTime(header)),
	}
}

func eventDescriptorToGo(descriptor C.EVENT_DESCRIPTOR) EventDescriptor {
	return EventDescriptor{
		ID:      uint16(descriptor.Id),
		Version: uint8(descriptor.Version),
		Channel: uint8(descriptor.Channel),
		Level:   uint8(descriptor.Level),
		OpCode:  uint8(descriptor.Opcode),
		Task:    uint16(descriptor.Task),
		Keyword: uint64(descriptor.Keyword),
	}
}
