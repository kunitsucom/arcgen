package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadMethodNameTable(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionMethodNameTable)
	return v
}

func MethodNameTable() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.MethodNameTable
}
