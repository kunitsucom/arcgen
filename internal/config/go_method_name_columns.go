package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoMethodNameColumns(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoMethodNameColumns)
	return v
}

func GoMethodNameColumns() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoMethodNameColumns
}
