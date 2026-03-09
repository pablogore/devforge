package testkit

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/pablogore/devforge/internal/ports"
)

// CommandCall records one Run/RunCombinedOutput invocation (dir, name, args).
type CommandCall struct {
	Dir  string
	Name string
	Args []string
}

// CommandResult is a single (Stdout, Err) result for a command.
type CommandResult struct {
	Stdout string
	Err    error
}

// CmdResponse is an alias for CommandResult for backward compatibility (Out/Err).
type CmdResponse struct {
	Out string
	Err error
}

// queuedEntry holds a name, args, and result for Enqueue.
type queuedEntry struct {
	name   string
	args   []string
	result CommandResult
}

// FakeCommandRunner implements ports.CommandRunner with exact stubbing, queued results,
// and optional default fallback. Safe for concurrent use.
type FakeCommandRunner struct {
	mu          sync.Mutex
	Calls       []CommandCall
	exactStub   map[string]CommandResult // key(name, args) -> result
	queue       []queuedEntry
	Responses   []CmdResponse // legacy: consumed in order when no stub/queue match
	idx         int
	Default     *CommandResult // fallback for unstubbed commands; nil = return "", nil
	defaultUsed bool           // set when Default was returned (for RequireNoUnexpectedCalls)
}

// NewFakeCommandRunner returns a new FakeCommandRunner ready for Stub/Enqueue or Default.
func NewFakeCommandRunner() *FakeCommandRunner {
	return &FakeCommandRunner{
		exactStub: make(map[string]CommandResult),
		queue:     nil,
		Calls:     nil,
	}
}

func key(name string, args []string) string {
	return name + "\x00" + strings.Join(args, "\x00")
}

// Stub sets an exact result for (name, args). Subsequent Run*(ctx, dir, name, args...) return this result regardless of dir.
func (f *FakeCommandRunner) Stub(name string, args []string, stdout string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.exactStub == nil {
		f.exactStub = make(map[string]CommandResult)
	}
	f.exactStub[key(name, args)] = CommandResult{Stdout: stdout, Err: err}
}

// Enqueue appends a result for (name, args). The next Run* call that matches (name, args) will consume this and return it.
func (f *FakeCommandRunner) Enqueue(name string, args []string, stdout string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queue = append(f.queue, queuedEntry{name: name, args: args, result: CommandResult{Stdout: stdout, Err: err}})
}

func (f *FakeCommandRunner) Run(_ context.Context, dir, name string, args ...string) (string, error) {
	return f.run(dir, name, args...)
}

func (f *FakeCommandRunner) RunCombinedOutput(_ context.Context, dir, name string, args ...string) (string, error) {
	return f.run(dir, name, args...)
}

func (f *FakeCommandRunner) RunCombinedOutputWithEnv(_ context.Context, dir string, _ []string, name string, args ...string) (string, error) {
	return f.run(dir, name, args...)
}

func (f *FakeCommandRunner) run(dir, name string, args ...string) (string, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, CommandCall{Dir: dir, Name: name, Args: args})
	out, err := f.lookup(name, args)
	f.mu.Unlock()
	return out, err
}

func (f *FakeCommandRunner) lookup(name string, args []string) (string, error) {
	k := key(name, args)
	if r, ok := f.exactStub[k]; ok {
		return r.Stdout, r.Err
	}
	if len(f.queue) > 0 {
		q := &f.queue[0]
		if q.name == name && sliceEqual(q.args, args) {
			r := q.result
			f.queue = f.queue[1:]
			return r.Stdout, r.Err
		}
	}
	if f.idx < len(f.Responses) {
		r := f.Responses[f.idx]
		f.idx++
		return r.Out, r.Err
	}
	if f.Default != nil {
		f.defaultUsed = true
		return f.Default.Stdout, f.Default.Err
	}
	return "", nil
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CallCount returns the number of recorded calls.
func (f *FakeCommandRunner) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.Calls)
}

// LastCall returns the last recorded call and true, or zero value and false if no calls.
func (f *FakeCommandRunner) LastCall() (CommandCall, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Calls) == 0 {
		return CommandCall{}, false
	}
	return f.Calls[len(f.Calls)-1], true
}

// WasCalled returns true if any recorded call had the given name and args (order matters).
func (f *FakeCommandRunner) WasCalled(name string, args ...string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.Calls {
		if c.Name == name && sliceEqual(c.Args, args) {
			return true
		}
	}
	return false
}

// RequireNoUnexpectedCalls returns an error if any call returned the Default (unstubbed) result.
func (f *FakeCommandRunner) RequireNoUnexpectedCalls() error {
	f.mu.Lock()
	used := f.defaultUsed
	f.mu.Unlock()
	if used {
		return errors.New("fake_command_runner: one or more calls had no stub or queued result (used Default)")
	}
	return nil
}

// Reset clears Calls, queue index for Responses, and defaultUsed. Stubs and Default are unchanged.
func (f *FakeCommandRunner) Reset() {
	f.mu.Lock()
	f.Calls = nil
	f.idx = 0
	f.defaultUsed = false
	f.queue = nil
	f.mu.Unlock()
}

var _ ports.CommandRunner = (*FakeCommandRunner)(nil)
