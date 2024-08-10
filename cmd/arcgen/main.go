package main

import (
	"context"
	"log"
	"os"

	"github.com/kunitsucom/arcgen/internal/contexts"
	"github.com/kunitsucom/arcgen/pkg/entrypoint/arcgen"
)

func main() {
	if err := arcgen.Run(contexts.WithArgs(context.Background(), os.Args)); err != nil {
		log.Fatalf("arcgen: %+v", err)
	}
}
