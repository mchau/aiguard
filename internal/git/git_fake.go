package git

import (
	"context"
	"fmt"
	"strings"
)

// FakeRunner returns canned outputs keyed by the first argument of the git command.
// Use in tests to avoid real git invocations.
type FakeRunner struct {
	// Responses maps "arg0 arg1..." to (stdout, error).
	Responses map[string]FakeResponse
}

// FakeResponse is a canned response for a git command.
type FakeResponse struct {
	Output []byte
	Err    error
}

func (f FakeRunner) Run(_ context.Context, args ...string) ([]byte, error) {
	key := strings.Join(args, " ")
	if r, ok := f.Responses[key]; ok {
		return r.Output, r.Err
	}
	return nil, fmt.Errorf("FakeRunner: no response for %q", key)
}
