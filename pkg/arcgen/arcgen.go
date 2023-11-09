package arcgen

import (
	"context"
	"errors"
	"fmt"
	"time"

	errorz "github.com/kunitsucom/util.go/errors"
	cliz "github.com/kunitsucom/util.go/exp/cli"

	arcgengo "github.com/kunitsucom/arcgen/internal/arcgen/lang/go"
	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/contexts"
	"github.com/kunitsucom/arcgen/internal/logs"
)

func ARCGen(ctx context.Context) error {
	if _, err := config.Load(ctx); err != nil {
		if errors.Is(err, cliz.ErrHelp) {
			return nil
		}
		return fmt.Errorf("config.Load: %w", err)
	}

	if config.Version() {
		fmt.Printf("version: %s\n", config.BuildVersion())           //nolint:forbidigo
		fmt.Printf("revision: %s\n", config.BuildRevision())         //nolint:forbidigo
		fmt.Printf("build branch: %s\n", config.BuildBranch())       //nolint:forbidigo
		fmt.Printf("build timestamp: %s\n", config.BuildTimestamp()) //nolint:forbidigo
		return nil
	}

	ctx = contexts.WithNowString(ctx, time.RFC3339, config.Timestamp())

	src := config.Source()
	logs.Info.Printf("source: %s", src)

	if err := generate(ctx, src); err != nil {
		return errorz.Errorf("parse: %w", err)
	}

	return nil
}

func generate(ctx context.Context, src string) error {
	switch language := config.Language(); language {
	case "go":
		if err := arcgengo.Generate(ctx, src); err != nil {
			return errorz.Errorf("arcgengo.Fprint: %w", err)
		}
		return nil
	default:
		return errorz.Errorf("unknown language: %s", language)
	}
}
