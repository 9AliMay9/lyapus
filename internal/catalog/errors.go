package catalog

import "errors"

var (
	ErrInvalidArgument = errors.New("catalog: invalid argument")
	ErrNotFound        = errors.New("catalog: not found")
	ErrConflict        = errors.New("catalog: conflict")
)

type InvalidArgumentError struct {
	Message string
}

func (e *InvalidArgumentError) Error() string {
	return e.Message
}

func (e *InvalidArgumentError) Unwrap() error {
	return ErrInvalidArgument
}
