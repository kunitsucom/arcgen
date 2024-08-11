package arcgengo

import (
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"strconv"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/config"
)

func fprintCRUDCommon(osFile osFile, buf buffer) error {
	content, err := generateCRUDCommonFileContent(buf)
	if err != nil {
		return errorz.Errorf("generateCRUDCommonFileContent: %w", err)
	}

	// write to file
	if _, err := io.WriteString(osFile, content); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}

	return nil
}

//nolint:funlen
func generateCRUDCommonFileContent(buf buffer) (string, error) {
	astFile := &ast.File{
		// package
		Name: &ast.Ident{
			Name: config.GoCRUDPackageName(),
		},
		// methods
		Decls: []ast.Decl{},
	}

	astFile.Decls = append(astFile.Decls,
		//	import (
		//		"context"
		//		"database/sql"
		//	)
		&ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				&ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: strconv.Quote("context"),
					},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{
						Kind:  token.STRING,
						Value: strconv.Quote("database/sql"),
					},
				},
			},
		},
	)

	// type sqlQueryerContext = interface {
	//	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	//	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	//	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// }
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{
						Name: "sqlQueryerContext",
					},
					Type: &ast.InterfaceType{
						Methods: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{{Name: "QueryContext"}},
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
												{Names: []*ast.Ident{{Name: "query"}}, Type: &ast.Ident{Name: "string"}},
												{Names: []*ast.Ident{{Name: "args"}}, Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "interface{}"}}},
											},
										},
										Results: &ast.FieldList{
											List: []*ast.Field{
												{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "Rows"}}}},
												{Type: &ast.Ident{Name: "error"}},
											},
										},
									},
								},
								{
									Names: []*ast.Ident{{Name: "QueryRowContext"}},
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
												{Names: []*ast.Ident{{Name: "query"}}, Type: &ast.Ident{Name: "string"}},
												{Names: []*ast.Ident{{Name: "args"}}, Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "interface{}"}}},
											},
										},
										Results: &ast.FieldList{
											List: []*ast.Field{
												{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "Row"}}}},
											},
										},
									},
								},
								{
									Names: []*ast.Ident{{Name: "ExecContext"}},
									Type: &ast.FuncType{
										Params: &ast.FieldList{
											List: []*ast.Field{
												{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
												{Names: []*ast.Ident{{Name: "query"}}, Type: &ast.Ident{Name: "string"}},
												{Names: []*ast.Ident{{Name: "args"}}, Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "interface{}"}}},
											},
										},
										Results: &ast.FieldList{
											List: []*ast.Field{
												{Type: &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "Result"}}},
												{Type: &ast.Ident{Name: "error"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	)

	//	type query struct {
	//		queryer sqlQueryerContext
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: "query"},
					Type: &ast.StructType{
						Fields: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{{Name: "queryer"}},
									Type:  &ast.Ident{Name: "sqlQueryerContext"},
								},
							},
						},
					},
				},
			},
		},
	)

	if err := printer.Fprint(buf, token.NewFileSet(), astFile); err != nil {
		return "", errorz.Errorf("printer.Fprint: %w", err)
	}

	// add header comment
	content := generateGoFileHeader() + buf.String()

	// add blank line between methods
	content = strings.ReplaceAll(content, "\n}\nfunc ", "\n}\n\nfunc ")

	return content, nil
}
