package util

import (
	"fmt"
	"go/build"
	"path/filepath"

	apperr "github.com/kunitsucom/arcgen/pkg/errors"
)

func GetPackageImportPath(path string) (string, error) {
	absDir, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("filepath.Abs: path=%s %w", path, err)
	}

	pkg, err := build.ImportDir(absDir, build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("build.ImportDir: path=%s: %w", path, err)
	}

	if pkg.ImportPath == "." {
		// If ImportPath is ".", it means the directory is not in GOPATH or inside a module
		return "", fmt.Errorf("path=%s: %w", absDir, apperr.ErrFailedToDetectPackageImportPath)
	}

	return pkg.ImportPath, nil
}
