package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoCRUDTypeName(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionGoCRUDTypeName)
	return v
}

func GoCRUDTypeName() string {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	return globalConfig.GoCRUDTypeName
}

func GoCRUDTypeNameUnexported() string {
	return "_" + GoCRUDTypeName()
}
