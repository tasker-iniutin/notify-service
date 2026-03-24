package domain

import "errors"

var (
	ErrValidation    = errors.New("validation failed")
	ErrBadPagination = errors.New("bad pagination")
	ErrNotFound      = errors.New("not found")
)
