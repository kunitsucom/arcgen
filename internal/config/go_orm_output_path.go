package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoORMOutputPath(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoORMOutputPath)
	return v
}

func GoORMOutputPath() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoORMOutputPath
}

func GenerateGoORMPackage() bool {
	return GoORMOutputPath() != ""
}
