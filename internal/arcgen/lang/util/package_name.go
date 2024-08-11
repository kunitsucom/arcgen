package util

import (
	"fmt"
	"go/build"
	"path/filepath"
)

func GetPackagePath(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	pkg, err := build.ImportDir(absDir, build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("failed to import directory: %w", err)
	}

	if pkg.ImportPath == "." {
		// If ImportPath is ".", it means the directory is not in GOPATH or inside a module
		// In this case, we'll use the last directory name as the package path
		return filepath.Base(absDir), nil
	}

	return pkg.ImportPath, nil
}
