package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoPKTag(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionGoPKTag)
	return v
}

func GoPKTag() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoPKTag
}
