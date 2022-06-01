# etw
[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8&style=flat-square)](https://pkg.go.dev/github.com/gaelmuller/etw)
[![Go Report Card](https://goreportcard.com/badge/github.com/bi-zone/etw)](https://goreportcard.com/report/github.com/gaelmuller/etw)
[![Lint & Test Go code](https://img.shields.io/github/workflow/status/bi-zone/etw/Lint%20&%20Test%20Go%20code?style=flat-square)](https://github.com/gaelmuller/etw/actions)

`etw` is a Go-package that allows you to receive [Event Tracing for Windows (ETW)](https://docs.microsoft.com/en-us/windows/win32/etw/about-event-tracing)
events in go code.

`etw` allows you to process events from new 
[TraceLogging](https://docs.microsoft.com/en-us/windows/win32/tracelogging/trace-logging-about) providers
as well as from classic (aka EventLog) providers, so you could actually listen to anything you can
see in Event Viewer window.

ETW API expects you to pass `stdcall` callback to process events, so `etw` **requires CGO** to be used. 
To use `etw` you need to have [mingw-w64](http://mingw-w64.org/) installed and pass some environment to the
Go compiler (take a look at [build/vars.sh](./build/vars.sh) and [examples/tracer/Makefile](./examples/tracer/Makefile)).

## Docs

Package reference is available at https://pkg.go.dev/github.com/gaelmuller/etw

You can look at `user_trace_test.go` and `kernel_trace_test.go` to see examples.