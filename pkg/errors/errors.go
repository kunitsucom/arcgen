package errors

import "errors"

var (
	ErrUnknownError                          = errors.New("unknown error")
	ErrNotSupported                          = errors.New("not supported")
	ErrUnformattedFileIsNotSupported         = errors.New("unformatted file is not supported")
	ErrColumnTagGoAnnotationNotFoundInSource = errors.New("column-tag-go annotation not found in source")
)
