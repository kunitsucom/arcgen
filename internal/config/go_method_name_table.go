package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoMethodNameTable(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoMethodNameTable)
	return v
}

func GoMethodNameTable() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoMethodNameTable
}
