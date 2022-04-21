//go:build windows
// +build windows

package etw

// SessionOptions describes Session subscription options.
//
// Most of options will be passed to EnableTraceEx2 and could be refined in
// its docs: https://docs.microsoft.com/en-us/windows/win32/api/evntrace/nf-evntrace-enabletraceex2
type SessionOptions struct {
	// Name specifies a name of ETW session being created. Further a session
	// could be controlled from other processed by it's name, so it should be
	// unique.
	Name string
}

// Option is any function that modifies SessionOptions. Options will be called
// on default config in NewSession. Subsequent options that modifies same
// fields will override each other.
type Option func(cfg *SessionOptions)

// WithName specifies a provided @name for the creating session. Further that
// session could be controlled from other processed by it's name, so it should be
// unique.
func WithName(name string) Option {
	return func(cfg *SessionOptions) {
		cfg.Name = name
	}
}
