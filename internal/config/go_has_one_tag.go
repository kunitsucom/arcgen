package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoHasOneTag(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoHasOneTag)
	return v
}

func GoHasOneTag() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoHasOneTag
}
