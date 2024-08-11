package errors

import "errors"

var (
	ErrLanguageNotSupported                  = errors.New("language not supported")
	ErrSourceFileOrDirectoryIsNotSet         = errors.New("source file or directory is not set")
	ErrInvalidSourceSet                      = errors.New("invalid source set")
	ErrGoColumnTagAnnotationNotFoundInSource = errors.New("go-column-tag annotation not found in source")
)
