package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadMethodNameColumns(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionMethodNameColumns)
	return v
}

func MethodNameColumns() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.MethodNameColumns
}
