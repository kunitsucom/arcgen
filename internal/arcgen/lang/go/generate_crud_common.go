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
)

func fprintCRUDCommon(osFile osFile, buf buffer, arcSrcSetSlice ARCSourceSetSlice) error {
	content, err := generateCRUDCommonFileContent(buf, arcSrcSetSlice)
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
func generateCRUDCommonFileContent(buf buffer, arcSrcSetSlice ARCSourceSetSlice) (string, error) {
	astFile := &ast.File{
		// package
		Name: &ast.Ident{
			Name: config.GoCRUDPackageName(),
		},
		// methods
		Decls: []ast.Decl{},
	}

	// Since all directories are the same from arcSrcSetSlice[0].Filename to arcSrcSetSlice[len(-1)].Filename,
	// get the package path from arcSrcSetSlice[0].Filename.
	dir := filepath.Dir(arcSrcSetSlice[0].Filename)
	structPackagePath, err := util.GetPackagePath(dir)
	if err != nil {
		return "", errorz.Errorf("GetPackagePath: %w", err)
	}

	astFile.Decls = append(astFile.Decls,
		//	import (
		//		"context"
		//		"database/sql"
		//
		//		dao "path/to/your/dao"
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
				&ast.ImportSpec{
					Name: &ast.Ident{Name: "dao"},
					Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(structPackagePath)},
				},
			},
		},
	)

	//	type sqlQueryerContext interface {
	//		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	//		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	//		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: "sqlQueryerContext"},
					Type: &ast.InterfaceType{
						Methods: &ast.FieldList{
							List: []*ast.Field{
								{
									Names: []*ast.Ident{{Name: "QueryContext"}},
									Type: &ast.FuncType{
										Params: &ast.FieldList{List: []*ast.Field{
											{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
											{Names: []*ast.Ident{{Name: "query"}}, Type: &ast.Ident{Name: "string"}},
											{Names: []*ast.Ident{{Name: "args"}}, Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "interface{}"}}},
										}},
										Results: &ast.FieldList{List: []*ast.Field{
											{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "Rows"}}}},
											{Type: &ast.Ident{Name: "error"}},
										}},
									},
								},
								{
									Names: []*ast.Ident{{Name: "QueryRowContext"}},
									Type: &ast.FuncType{
										Params: &ast.FieldList{List: []*ast.Field{
											{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
											{Names: []*ast.Ident{{Name: "query"}}, Type: &ast.Ident{Name: "string"}},
											{Names: []*ast.Ident{{Name: "args"}}, Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "interface{}"}}},
										}},
										Results: &ast.FieldList{List: []*ast.Field{
											{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "Row"}}}},
										}},
									},
								},
								{
									Names: []*ast.Ident{{Name: "ExecContext"}},
									Type: &ast.FuncType{
										Params: &ast.FieldList{List: []*ast.Field{
											{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
											{Names: []*ast.Ident{{Name: "query"}}, Type: &ast.Ident{Name: "string"}},
											{Names: []*ast.Ident{{Name: "args"}}, Type: &ast.Ellipsis{Elt: &ast.Ident{Name: "interface{}"}}},
										}},
										Results: &ast.FieldList{List: []*ast.Field{
											{Type: &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "Result"}}},
											{Type: &ast.Ident{Name: "error"}},
										}},
									},
								},
							},
						},
					},
				},
			},
		},
	)

	//	type query struct {}
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: "query"},
					Type: &ast.StructType{Fields: &ast.FieldList{}},
				},
			},
		},
	)

	//	type Query interface {
	//		Create{StructName}(ctx context.Context, queryer sqlQueryerContext, s *{Struct}) error
	//		Find{StructName}(ctx context.Context, queryer sqlQueryerContext, pk1 pk1type, ...) (*{Struct}, error)
	//		 	...
	//	}
	methods := make([]*ast.Field, 0)
	for _, arcSrcSet := range arcSrcSetSlice {
		for _, arcSrc := range arcSrcSet.ARCSourceSlice {
			structName := arcSrc.extractStructName()
			pks := arcSrc.extractFieldNamesAndColumnNames().PrimaryKeys()
			methods = append(methods,
				&ast.Field{
					Names: []*ast.Ident{{Name: "Create" + structName}},
					Type: &ast.FuncType{
						Params: &ast.FieldList{List: []*ast.Field{
							{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
							{Names: []*ast.Ident{{Name: "queryer"}}, Type: &ast.Ident{Name: "sqlQueryerContext"}},
							{Names: []*ast.Ident{{Name: "s"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "dao." + structName}}},
						}},
						Results: &ast.FieldList{List: []*ast.Field{
							{Type: &ast.Ident{Name: "error"}},
						}},
					},
				},
				&ast.Field{
					Names: []*ast.Ident{{Name: "Find" + structName}},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: append([]*ast.Field{
								{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
								{Names: []*ast.Ident{{Name: "queryer"}}, Type: &ast.Ident{Name: "sqlQueryerContext"}},
							},
								func() []*ast.Field {
									fields := make([]*ast.Field, 0)
									for _, pk := range pks {
										fields = append(fields, &ast.Field{
											Names: []*ast.Ident{{Name: pk.FieldName}},
											Type:  &ast.Ident{Name: pk.FieldType},
										})
									}
									return fields
								}()...),
						},
						Results: &ast.FieldList{List: []*ast.Field{
							{Type: &ast.StarExpr{X: &ast.Ident{Name: "dao." + structName}}},
							{Type: &ast.Ident{Name: "error"}},
						}},
					},
				},
			)
		}
	}

	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: "Query"},
					Type: &ast.InterfaceType{Methods: &ast.FieldList{List: methods}},
				},
			},
		},
	)

	// func NewQuery() Query {
	//	return &query{}
	// }
	astFile.Decls = append(astFile.Decls,
		&ast.FuncDecl{
			Name: &ast.Ident{Name: "NewQuery"},
			Type: &ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "Query"}}}}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.UnaryExpr{Op: token.AND, X: &ast.Ident{Name: "query{}"}}}}}},
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
