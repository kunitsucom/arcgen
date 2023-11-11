package errors

import "errors"

var (
	ErrInvalidSourceSet                      = errors.New("invalid source set")
	ErrColumnTagGoAnnotationNotFoundInSource = errors.New("column-tag-go annotation not found in source")
)
