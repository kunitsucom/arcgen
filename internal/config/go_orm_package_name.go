package config

import (
	"context"
	"path/filepath"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoORMPackageName(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionGoORMPackageName)
	return v
}

func GoORMPackageName() string {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	if globalConfig.GoORMPackageName == "" {
		globalConfig.GoORMPackageName = filepath.Base(globalConfig.GoORMPackagePath)
	}

	return globalConfig.GoORMPackageName
}