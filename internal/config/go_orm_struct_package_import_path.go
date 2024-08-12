package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoORMStructPackageImportPath(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoORMStructPackageImportPath)
	return v
}

func GoORMStructPackageImportPath() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoORMStructPackageImportPath
}
