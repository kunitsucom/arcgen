package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoCRUDPackagePath(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionGoCRUDPackagePath)
	return v
}

func GoCRUDPackagePath() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoCRUDPackagePath
}

func GenerateGoCRUDPackage() bool {
	return GoCRUDPackagePath() != ""
}
