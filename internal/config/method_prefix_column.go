package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadMethodPrefixColumn(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionMethodPrefixColumn)
	return v
}

func MethodPrefixColumn() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.MethodPrefixColumn
}
