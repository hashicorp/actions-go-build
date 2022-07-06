package cli

import "errors"

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNoArgsAllowed  = errors.New("no args allowed")
)
