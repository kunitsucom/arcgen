package config

import (
	"context"

	cliz "github.com/kunitsucom/util.go/exp/cli"
)

func loadGoHasManyTag(_ context.Context, cmd *cliz.Command) string {
	v, _ := cmd.GetOptionString(OptionGoHasManyTag)
	return v
}

func GoHasManyTag() string {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return globalConfig.GoHasManyTag
}
