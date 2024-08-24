package arcgengo

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/logs"
	"github.com/kunitsucom/arcgen/internal/util"
	apperr "github.com/kunitsucom/arcgen/pkg/errors"
)

func parse(ctx context.Context, src string) (ARCSourceSetSlice, error) {
	// MEMO: get absolute path for parser.ParseFile()
	sourceAbs := util.Abs(src)

	info, err := os.Stat(sourceAbs)
	if err != nil {
		return nil, errorz.Errorf("os.Stat: %w", err)
	}

	if info.IsDir() {
		arcSrcSetSlice := make(ARCSourceSetSlice, 0)
		if err := filepath.WalkDir(sourceAbs, walkDirFn(ctx, &arcSrcSetSlice)); err != nil {
			return nil, errorz.Errorf("filepath.WalkDir: %w", err)
		}

		return arcSrcSetSlice, nil
	}

	arcSrcSet, err := parseFile(ctx, sourceAbs)
	if err != nil {
		return nil, errorz.Errorf("parseFile: file=%s: %v", sourceAbs, err)
	}

	return ARCSourceSetSlice{arcSrcSet}, nil
}

//nolint:gochecknoglobals
var fileExt = ".go"

func walkDirFn(ctx context.Context, arcSrcSetSlice *ARCSourceSetSlice) func(path string, d os.DirEntry, err error) error {
	return func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err //nolint:wrapcheck
		}

		if d.IsDir() || !strings.HasSuffix(path, fileExt) || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		arcSrcSet, err := parseFile(ctx, path)
		if err != nil {
			if errors.Is(err, apperr.ErrGoColumnTagAnnotationNotFoundInSource) {
				logs.Debug.Printf("SKIP: parseFile: %v", err)
				return nil
			}

			logs.Info.Printf("SKIP NON-GO FILE: parseFile: %v", err)
			return nil
		}

		*arcSrcSetSlice = append(*arcSrcSetSlice, arcSrcSet)

		return nil
	}
}

func parseFile(ctx context.Context, filename string) (*FileSource, error) {
	fset := token.NewFileSet()
	rootNode, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		// MEMO: parser.ParseFile err contains file path, so no need to log it
		return nil, errorz.Errorf("parser.ParseFile: %w", err)
	}

	arcSrcSet, err := extractSource(ctx, fset, rootNode)
	if err != nil {
		return nil, errorz.Errorf("extractSource: filename=%s: %w", filename, err)
	}

	dumpSource(fset, arcSrcSet)

	return arcSrcSet, nil
}
