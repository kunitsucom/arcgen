package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadMethodPrefixGlobal(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionMethodPrefixGlobal)
	return v
}

func MethodPrefixGlobal() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.MethodPrefixGlobal
}
