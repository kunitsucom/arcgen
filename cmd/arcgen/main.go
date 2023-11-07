package main

import (
	"context"
	"log"

	"github.com/kunitsucom/arcgen/pkg/arcgen"
)

func main() {
	ctx := context.Background()

	if err := arcgen.ARCGen(ctx); err != nil {
		log.Fatalf("arcgen: %+v", err)
	}
}
