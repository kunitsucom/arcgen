package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoORMPackagePath(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionGoORMPackagePath)
	return v
}

func GoORMPackagePath() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoORMPackagePath
}

func GenerateGoORMPackage() bool {
	return GoORMPackagePath() != ""
}
