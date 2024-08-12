package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoORMStructName(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoORMStructName)
	return v
}

func GoORMStructName() string {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	if globalConfig.GoORMStructName == "" {
		globalConfig.GoORMStructName = "_" + globalConfig.GoORMTypeName
	}

	return globalConfig.GoORMStructName
}
