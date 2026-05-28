package checks

import (
	"context"
	"fmt"
)

var registered []Check

// Register adds a check to the global registry.
// Called by each check's init() when the check has real logic.
func Register(c Check) {
	registered = append(registered, c)
}

// RunAll executes every registered check and returns all results.
// If a check errors, the error is wrapped and returned alongside any results collected so far.
func RunAll(ctx context.Context, input CheckInput) ([]DeterministicCheckResult, error) {
	var all []DeterministicCheckResult
	for _, c := range registered {
		results, err := c.Run(ctx, input)
		if err != nil {
			return all, fmt.Errorf("check %s: %w", c.ID(), err)
		}
		all = append(all, results...)
	}
	return all, nil
}
