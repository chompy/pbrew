package cli

import "errors"

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrInvalidService  = errors.New("invalid service")
)
