package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoSliceTypeSuffix(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoSliceTypeSuffix)
	return v
}

func GoSliceTypeSuffix() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoSliceTypeSuffix
}
