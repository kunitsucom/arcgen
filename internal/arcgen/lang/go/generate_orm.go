package arcgengo

import (
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/pkg/errors"
)

func fprintORM(osFile osFile, buf buffer, arcSrcSet *ARCSourceSet) error {
	content, err := generateORMFileContent(buf, arcSrcSet)
	if err != nil {
		return errorz.Errorf("generateORMFileContent: %w", err)
	}

	// write to file
	if _, err := io.WriteString(osFile, content); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}

	return nil
}

//nolint:funlen
func generateORMFileContent(buf buffer, arcSrcSet *ARCSourceSet) (string, error) {
	if arcSrcSet == nil || arcSrcSet.PackageName == "" {
		return "", errors.ErrInvalidSourceSet
	}
	astFile := &ast.File{
		// package
		Name: &ast.Ident{
			Name: config.GoORMPackageName(),
		},
		// methods
		Decls: []ast.Decl{},
	}

	structPackagePath, err := util.GetPackagePath(filepath.Dir(arcSrcSet.Filename))
	if err != nil {
		return "", errorz.Errorf("GetPackagePath: %w", err)
	}

	// import
	astFile.Decls = append(astFile.Decls,
		//	import (
		//		"context"
		//		"fmt"
		//
		//		orm "path/to/your/orm"
		//	)
		&ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("context")},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("fmt")},
				},
				&ast.ImportSpec{
					Name: &ast.Ident{Name: importName},
					Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(structPackagePath)},
				},
			},
		},
	)

	generateCREATEContent(astFile, arcSrcSet)
	generateREADContent(astFile, arcSrcSet)
	generateUPDATEContent(astFile, arcSrcSet)
	generateDELETEContent(astFile, arcSrcSet)

	if err := printer.Fprint(buf, token.NewFileSet(), astFile); err != nil {
		return "", errorz.Errorf("printer.Fprint: %w", err)
	}

	// add header comment
	content := arcSrcSet.generateGoFileHeader() + buf.String()

	// add blank line between methods
	content = strings.ReplaceAll(content, "\n}\nfunc ", "\n}\n\nfunc ")

	return content, nil
}
