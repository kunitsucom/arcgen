package main

import (
	"context"
	"log"
	"os"

	"github.com/kunitsucom/arcgen/internal/contexts"
	"github.com/kunitsucom/arcgen/pkg/arcgen"
)

func main() {
	if err := arcgen.ARCGen(contexts.WithArgs(context.Background(), os.Args)); err != nil {
		log.Fatalf("arcgen: %+v", err)
	}
}
