package tasks

import (
	"errors"
)

// ErrScanCancelled is returned by Handle when a scan was cancelled.
var ErrScanCancelled = errors.New("scan cancelled")