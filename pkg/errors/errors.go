package errors

import (
	"errors"

	"github.com/kunitsucom/arcgen/internal/config"
)

var (
	ErrLanguageNotSupported                  = errors.New("language not supported")
	ErrFailedToDetectPackageImportPath       = errors.New("failed to detect package import path. Please use the --" + config.OptionGoORMStructPackageImportPath + " option, or run include the package in your GOPATH or module (GO111MODULE=auto may be required)")
	ErrSourceFileOrDirectoryIsNotSet         = errors.New("source file or directory is not set")
	ErrInvalidSourceSet                      = errors.New("invalid source set")
	ErrGoColumnTagAnnotationNotFoundInSource = errors.New("go-column-tag annotation not found in source")
)
