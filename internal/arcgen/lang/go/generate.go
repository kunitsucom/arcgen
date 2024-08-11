package arcgengo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/config"
)

//nolint:cyclop,funlen
func Generate(ctx context.Context, src string) error {
	arcSrcSets, err := parse(ctx, src)
	if err != nil {
		return errorz.Errorf("parse: %w", err)
	}

	if err := generate(arcSrcSets); err != nil {
		return errorz.Errorf("generate: %w", err)
	}

	return nil
}

func generate(arcSrcSets ARCSourceSetSlice) error {
	for _, arcSrcSet := range arcSrcSets {
		const rw_r__r__ = 0o644 //nolint:revive,stylecheck // rw-r--r--

		// closure for defer
		if err := func() error {
			filePathWithoutExt := strings.TrimSuffix(arcSrcSet.Filename, fileExt)
			newExt := fmt.Sprintf(".%s.gen%s", config.GoColumnTag(), fileExt)
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

	return nil
}

type buffer = interface {
	io.Writer
	fmt.Stringer
}
