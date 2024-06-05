package arcgen

import (
	"context"
	"errors"
	"fmt"

	errorz "github.com/kunitsucom/util.go/errors"
	cliz "github.com/kunitsucom/util.go/exp/cli"
	"github.com/kunitsucom/util.go/version"

	arcgengo "github.com/kunitsucom/arcgen/internal/arcgen/lang/go"
	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/logs"
)

func ARCGen(ctx context.Context) error {
	_, remainingArgs, err := config.Load(ctx)
	if err != nil {
		if errors.Is(err, cliz.ErrHelp) {
			return nil
		}
		return fmt.Errorf("config.Load: %w", err)
	}

	if config.Version() {
		fmt.Printf("version: %s\n", version.Version())           //nolint:forbidigo
		fmt.Printf("revision: %s\n", version.Revision())         //nolint:forbidigo
		fmt.Printf("build branch: %s\n", version.Branch())       //nolint:forbidigo
		fmt.Printf("build timestamp: %s\n", version.Timestamp()) //nolint:forbidigo
		return nil
	}

	if len(remainingArgs) == 0 {
		remainingArgs = []string{"/dev/stdin"}
	}

	for _, src := range remainingArgs {
		logs.Info.Printf("source: %s", src)
		if err := generate(ctx, src); err != nil {
			return errorz.Errorf("parse: %w", err)
		}
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
