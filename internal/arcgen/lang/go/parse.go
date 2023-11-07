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

func parse(ctx context.Context, src string) (ARCSourceSets, error) {
	// MEMO: get absolute path for parser.ParseFile()
	sourceAbs := util.Abs(src)

	info, err := os.Stat(sourceAbs)
	if err != nil {
		return nil, errorz.Errorf("os.Stat: %w", err)
	}

	arcSrcSets := make(ARCSourceSets, 0)

	if info.IsDir() {
		if err := filepath.WalkDir(sourceAbs, walkDirFn(ctx, &arcSrcSets)); err != nil {
			return nil, errorz.Errorf("filepath.WalkDir: %w", err)
		}

		return arcSrcSets, nil
	}

	arcSrcSet, err := parseFile(ctx, sourceAbs)
	if err != nil {
		return nil, errorz.Errorf("parseFile: file=%s: %v", sourceAbs, err)
	}

	arcSrcSets = append(arcSrcSets, arcSrcSet)
	return arcSrcSets, nil
}

//nolint:gochecknoglobals
var fileSuffix = ".go"

func walkDirFn(ctx context.Context, arcSrcSets *ARCSourceSets) func(path string, d os.DirEntry, err error) error {
	return func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err //nolint:wrapcheck
		}

		if d.IsDir() || !strings.HasSuffix(path, fileSuffix) || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		arcSrcSet, err := parseFile(ctx, path)
		if err != nil {
			if errors.Is(err, apperr.ErrColumnTagGoAnnotationNotFoundInSource) {
				logs.Debug.Printf("SKIP: parseFile: file=%s: %v", path, err)
				return nil
			}
			return errorz.Errorf("parseFile: file=%s: %v", path, err)
		}

		*arcSrcSets = append(*arcSrcSets, arcSrcSet)

		return nil
	}
}

func parseFile(ctx context.Context, filename string) (ARCSourceSet, error) {
	fset := token.NewFileSet()
	rootNode, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, errorz.Errorf("parser.ParseFile: %w", err)
	}

	arcSrcSet, err := extractSource(ctx, fset, rootNode)
	if err != nil {
		return nil, errorz.Errorf("extractSource: %w", err)
	}

	dumpSource(fset, arcSrcSet)

	return arcSrcSet, nil
}
