package arcgengo

import (
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/pkg/errors"
)

func fprintCRUD(osFile osFile, buf buffer, arcSrcSet *ARCSourceSet) error {
	content, err := generateCRUDFileContent(buf, arcSrcSet)
	if err != nil {
		return errorz.Errorf("generateCRUDFileContent: %w", err)
	}

	// write to file
	if _, err := io.WriteString(osFile, content); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}

	return nil
}

//nolint:funlen
func generateCRUDFileContent(buf buffer, arcSrcSet *ARCSourceSet) (string, error) {
	if arcSrcSet == nil || arcSrcSet.PackageName == "" {
		return "", errors.ErrInvalidSourceSet
	}
	astFile := &ast.File{
		// package
		Name: &ast.Ident{
			Name: config.GoCRUDPackageName(),
		},
		// methods
		Decls: []ast.Decl{},
	}

	if err := generateCREATEContent(astFile, arcSrcSet); err != nil {
		return "", errorz.Errorf("generateCREATEContent: %w", err)
	}

	if err := printer.Fprint(buf, token.NewFileSet(), astFile); err != nil {
		return "", errorz.Errorf("printer.Fprint: %w", err)
	}

	// add header comment
	content := arcSrcSet.generateGoFileHeader() + buf.String()

	// add blank line between methods
	content = strings.ReplaceAll(content, "\n}\nfunc ", "\n}\n\nfunc ")

	return content, nil
}
