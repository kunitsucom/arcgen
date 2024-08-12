package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoORMTypeName(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoORMTypeName)
	return v
}

func GoORMTypeName() string {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	return globalConfig.GoORMTypeName
}
