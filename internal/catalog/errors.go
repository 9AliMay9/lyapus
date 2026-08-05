package catalog

import "errors"

var (
	ErrInvalidArgument = errors.New("catalog: invalid argument")
	ErrNotFound        = errors.New("catalog: not found")
	ErrConflict        = errors.New("catalog: conflict")
)
