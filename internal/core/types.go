package core

import (
	"fmt"
	"net/url"
	"strings"
)

// URI is a storage location routed to a backend by its scheme.
type URI string

func (u URI) Parse() (*url.URL, error) {
	return url.Parse(string(u))
}

// MultiError represents multiple keyed errors.
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
	errList := make([]error, 0, len(e.Errs))
	for _, err := range e.Errs {
		errList = append(errList, err)
	}
	return errList
}
