package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoColumnTag(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoColumnTag)
	return v
}

func GoColumnTag() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoColumnTag
}
