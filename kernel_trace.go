//go:build windows
// +build windows
package etw

/*
	#cgo LDFLAGS: -ltdh

	#include "etw.h"
*/
import "C"

type KernelTrace struct{}

func NewKernelTrace(name string, callback EventCallback) (*Trace, error) {
	return newTrace(name, callback, &KernelTrace{})
}

func (u *KernelTrace) setTraceProperties(trace *Trace) {
	var flags uint64

	for _, provider := range trace.providers {
		flags |= provider.EnableFlags
	}

	trace.properties.LogFileMode |= C.EVENT_TRACE_SYSTEM_LOGGER_MODE
	trace.properties.EnableFlags = C.ulong(flags)
}

// For Kernel Trace, providers were already enabled via Trace Properties
func (u *KernelTrace) enableProviders(trace *Trace) error {
	return nil
}

// Nothing to do for kernel traces
func (u *KernelTrace) disableProviders(trace *Trace) error {
	return nil
}
