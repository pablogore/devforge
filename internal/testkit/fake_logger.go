package testkit

import (
	"github.com/pablogore/devforge/internal/ports"
)

// LogCall records one Info or Error call.
type LogCall struct {
	Msg  string
	Args []any
}

// FakeLogger implements ports.Logger and records the last Info/Error and optional history for assertions.
type FakeLogger struct {
	LastInfoMsg   string
	LastInfoArgs  []any
	LastErrorMsg  string
	LastErrorArgs []any
	InfoCalls     int
	ErrorCalls    int
	// RecordInfoHistory appends each Info call to InfoHistory when true (default false).
	RecordInfoHistory bool
	InfoHistory       []LogCall
}

// Debug is a no-op.
func (f *FakeLogger) Debug(string, ...any) {}

// Info records msg and args, increments InfoCalls.
func (f *FakeLogger) Info(msg string, args ...any) {
	f.LastInfoMsg = msg
	f.LastInfoArgs = args
	f.InfoCalls++
	if f.RecordInfoHistory {
		f.InfoHistory = append(f.InfoHistory, LogCall{Msg: msg, Args: args})
	}
}

// Warn is a no-op.
func (f *FakeLogger) Warn(string, ...any) {}

// Error records msg and args, increments ErrorCalls.
func (f *FakeLogger) Error(msg string, args ...any) {
	f.LastErrorMsg = msg
	f.LastErrorArgs = args
	f.ErrorCalls++
}

// With returns the same logger.
func (f *FakeLogger) With(...any) ports.Logger { return f }

// Sync returns nil.
func (f *FakeLogger) Sync() error { return nil }

var _ ports.Logger = (*FakeLogger)(nil)
