package config

import (
	"context"
	"path/filepath"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoCRUDPackageName(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(_OptionGoCRUDPackageName)
	return v
}

func GoCRUDPackageName() string {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	if globalConfig.GoCRUDPackageName == "" {
		globalConfig.GoCRUDPackageName = filepath.Base(globalConfig.GoCRUDPackagePath)
	}

	return globalConfig.GoCRUDPackageName
}
