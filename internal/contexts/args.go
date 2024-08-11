package contexts

import (
	"context"
	"os"
)

type contextKeyArgs struct{}

func OSArgs(ctx context.Context) []string {
	if v, ok := ctx.Value(contextKeyArgs{}).([]string); ok {
		return v
	}

	return os.Args[0:]
}

func WithOSArgs(ctx context.Context, osArgs []string) context.Context {
	return context.WithValue(ctx, contextKeyArgs{}, osArgs)
}
