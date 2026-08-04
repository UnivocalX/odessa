package tasks

import (
	"fmt"
	"strings"
)

// MultiError represents multiple keyed errors.
// It supports errors.Is/errors.As via the Go 1.20+ multi-unwrap interface.
type MultiError struct {
	Errs map[string]error
}

func (e *MultiError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d error(s):", len(e.Errs))
	for key, err := range e.Errs {
		fmt.Fprintf(&b, "\n  %s: %s", key, err)
	}
	return b.String()
}

func (e *MultiError) Unwrap() []error {
	errs := make([]error, 0, len(e.Errs))
	for _, err := range e.Errs {
		errs = append(errs, err)
	}
	return errs
}
