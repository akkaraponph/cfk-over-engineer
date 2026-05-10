package saga

import "errors"

var (
	ErrSagaNotFound = errors.New("saga definition not found")
	ErrInstanceNotFound = errors.New("saga instance not found")
)
