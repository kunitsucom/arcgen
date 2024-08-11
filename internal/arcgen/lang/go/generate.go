package arcgengo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/config"
)

//nolint:cyclop,funlen
func Generate(ctx context.Context, src string) error {
	arcSrcSetSlice, err := parse(ctx, src)
	if err != nil {
		return errorz.Errorf("parse: %w", err)
	}

	if err := generate(arcSrcSetSlice); err != nil {
		return errorz.Errorf("generate: %w", err)
	}

	return nil
}

const rw_r__r__ = 0o644 //nolint:revive,stylecheck // rw-r--r--

//nolint:cyclop,funlen,gocognit
func generate(arcSrcSetSlice ARCSourceSetSlice) error {
	newExt := fmt.Sprintf(".%s.gen%s", config.GoColumnTag(), fileExt)

	for _, arcSrcSet := range arcSrcSetSlice {
		// closure for defer
		if err := func() error {
			filePathWithoutExt := strings.TrimSuffix(arcSrcSet.Filename, fileExt)
			filename := filePathWithoutExt + newExt
			f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, rw_r__r__)
			if err != nil {
				return errorz.Errorf("os.OpenFile: %w", err)
			}
			defer f.Close()

			if err := fprintColumns(f, bytes.NewBuffer(nil), arcSrcSet); err != nil {
				return errorz.Errorf("sprint: %w", err)
			}
			return nil
		}(); err != nil {
			return errorz.Errorf("f: %w", err)
		}
	}

	if config.GenerateGoCRUDPackage() {
		newExt := fmt.Sprintf(".%s.crud.gen%s", config.GoColumnTag(), fileExt)

		if err := func() error {
			filename := filepath.Join(config.GoCRUDPackagePath(), "common"+newExt)
			f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, rw_r__r__)
			if err != nil {
				return errorz.Errorf("os.OpenFile: %w", err)
			}
			defer f.Close()

			if err := fprintCRUDCommon(f, bytes.NewBuffer(nil), arcSrcSetSlice); err != nil {
				return errorz.Errorf("sprint: %w", err)
			}

			return nil
		}(); err != nil {
			return errorz.Errorf("f: %w", err)
		}

		for _, arcSrcSet := range arcSrcSetSlice {
			// closure for defer
			if err := func() error {
				filePathWithoutExt := strings.TrimSuffix(filepath.Base(arcSrcSet.Filename), fileExt)
				filename := filepath.Join(config.GoCRUDPackagePath(), filePathWithoutExt+newExt)
				f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, rw_r__r__)
				if err != nil {
					return errorz.Errorf("os.OpenFile: %w", err)
				}
				defer f.Close()
				f.Name()

				if err := fprintCRUD(
					f,
					bytes.NewBuffer(nil),
					arcSrcSet,
				); err != nil {
					return errorz.Errorf("sprint: %w", err)
				}
				return nil
			}(); err != nil {
				return errorz.Errorf("f: %w", err)
			}
		}
	}

	return nil
}

type buffer = interface {
	io.Writer
	fmt.Stringer
}

type osFile = interface {
	io.Writer
	Name() string
}
