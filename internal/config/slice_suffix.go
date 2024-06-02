package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadSliceTypeSuffix(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionSliceTypeSuffix)
	return v
}

func SliceTypeSuffix() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.SliceTypeSuffix
}
