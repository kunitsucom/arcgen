package arcgengo

import (
	"go/ast"
	"go/parser"
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

const (
	importName             = "orm"
	receiverName           = "_orm"
	queryerContextVarName  = "dbtx"
	queryerContextTypeName = "DBTX"
	readOneFuncPrefix      = "Get"
	readManyFuncPrefix     = "List"
)

func fprintORMCommon(osFile osFile, buf buffer, arcSrcSetSlice ARCSourceSetSlice, crudFiles []string) error {
	content, err := generateORMCommonFileContent(buf, arcSrcSetSlice, crudFiles)
	if err != nil {
		return errorz.Errorf("generateORMCommonFileContent: %w", err)
	}

	// write to file
	if _, err := io.WriteString(osFile, content); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}

	return nil
}

//nolint:cyclop,funlen,gocognit,maintidx
func generateORMCommonFileContent(buf buffer, arcSrcSetSlice ARCSourceSetSlice, crudFiles []string) (string, error) {
	astFile := &ast.File{
		// package
		Name: &ast.Ident{
			Name: config.GoORMPackageName(),
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
		//		"log/slog"
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
					Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("database/sql")},
				},
				&ast.ImportSpec{
					Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("log/slog")},
				},
				&ast.ImportSpec{
					Name: &ast.Ident{Name: importName},
					Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(structPackagePath)},
				},
			},
		},
	)

	//	type ORM interface {
	//		Create{StructName}(ctx context.Context, queryerContext QueryerContext, s *{Struct}) error
	//		...
	//	}
	methods := make([]*ast.Field, 0)
	fset := token.NewFileSet()
	for _, crudFile := range crudFiles {
		rootNode, err := parser.ParseFile(fset, crudFile, nil, parser.ParseComments)
		if err != nil {
			// MEMO: parser.ParseFile err contains file path, so no need to log it
			return "", errorz.Errorf("parser.ParseFile: %w", err)
		}

		// MEMO: Inspect is used to get the method declaration from the file
		ast.Inspect(rootNode, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.FuncDecl:
				//nolint:nestif
				if n.Recv != nil && len(n.Recv.List) > 0 {
					if t, ok := n.Recv.List[0].Type.(*ast.StarExpr); ok {
						if ident, ok := t.X.(*ast.Ident); ok {
							if ident.Name == config.GoORMStructName() {
								methods = append(methods, &ast.Field{
									Names: []*ast.Ident{{Name: n.Name.Name}},
									Type:  n.Type,
								})
							}
						}
					}
				}
			default:
				// noop
			}
			return true
		})
	}
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: config.GoORMTypeName()},
					Type: &ast.InterfaceType{
						Methods: &ast.FieldList{List: methods},
					},
				},
			},
		},
	)

	//	func NewORM(opts ...ORMOption) ORM {
	//		o := new(_ORM)
	//		for _, opt := range opts {
	//			opt.Apply(o)
	//		}
	//		return o
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.FuncDecl{
			Name: &ast.Ident{Name: "New" + config.GoORMTypeName()},
			Type: &ast.FuncType{
				Params:  &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "opts"}}, Type: &ast.Ellipsis{Elt: &ast.Ident{Name: config.GoORMTypeName() + "Option"}}}}},
				Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: config.GoORMTypeName()}}}},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{&ast.Ident{Name: "o"}},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{&ast.CallExpr{Fun: &ast.Ident{Name: "new"}, Args: []ast.Expr{&ast.Ident{Name: config.GoORMStructName()}}}},
					},
					&ast.RangeStmt{
						Key:   &ast.Ident{Name: "_"},
						Value: &ast.Ident{Name: "opt"},
						Tok:   token.DEFINE,
						X:     &ast.Ident{Name: "opts"},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "opt"}, Sel: &ast.Ident{Name: "Apply"}}, Args: []ast.Expr{&ast.Ident{Name: "o"}}}},
							},
						},
					},
					&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "o"}}},
				},
			},
		},
	)

	//	type ORMOption interface {
	//		Apply(o *_ORM)
	//	}
	astFile.Decls = append(astFile.Decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{&ast.TypeSpec{
			Name: &ast.Ident{Name: config.GoORMTypeName() + "Option"},
			Type: &ast.InterfaceType{Methods: &ast.FieldList{List: []*ast.Field{{
				Names: []*ast.Ident{{Name: "Apply"}},
				Type:  &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "o"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: config.GoORMStructName()}}}}}},
			}}}},
		}},
	})

	//nolint:dupword
	//	func NewORMOptionHandleErrorFunc(handleErrorFunc func(ctx context.Context, err error) error) ORMOption {
	//		return &ormOptionHandleErrorFunc{handleErrorFunc: handleErrorFunc}
	//	}
	//	type ormOptionHandleErrorFunc struct {
	//		handleErrorFunc func(ctx context.Context, err error) error
	//	}
	//	func (o *ormOptionHandleErrorFunc) Apply(s *_ORM) {
	//		s.HandleErrorFunc = o.handleErrorFunc
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.FuncDecl{
			Name: &ast.Ident{Name: "New" + config.GoORMTypeName() + "OptionHandleErrorFunc"},
			Type: &ast.FuncType{
				Params:  &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "handleErrorFunc"}}, Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}}, {Names: []*ast.Ident{{Name: "err"}}, Type: &ast.Ident{Name: "error"}}}}, Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}}}}}},
				Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: config.GoORMTypeName() + "Option"}}}},
			},
			Body: &ast.BlockStmt{List: []ast.Stmt{
				&ast.ReturnStmt{Results: []ast.Expr{
					&ast.UnaryExpr{Op: token.AND, X: &ast.CompositeLit{
						Type: &ast.Ident{Name: "ormOptionHandleErrorFunc"},
						Elts: []ast.Expr{&ast.KeyValueExpr{Key: &ast.Ident{Name: "handleErrorFunc"}, Value: &ast.Ident{Name: "handleErrorFunc"}}},
					}},
				}},
			}},
		},
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{&ast.TypeSpec{
				Name: &ast.Ident{Name: "ormOptionHandleErrorFunc"},
				Type: &ast.StructType{Fields: &ast.FieldList{List: []*ast.Field{{
					Names: []*ast.Ident{{Name: "handleErrorFunc"}},
					Type:  &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}}, {Names: []*ast.Ident{{Name: "err"}}, Type: &ast.Ident{Name: "error"}}}}, Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}}},
				}}}},
			}},
		},
		&ast.FuncDecl{
			Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "o"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "ormOptionHandleErrorFunc"}}}}},
			Name: &ast.Ident{Name: "Apply"},
			Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "s"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: config.GoORMStructName()}}}}}},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{&ast.AssignStmt{
					Lhs: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "s"}, Sel: &ast.Ident{Name: "HandleErrorFunc"}}},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{&ast.SelectorExpr{X: &ast.Ident{Name: "o"}, Sel: &ast.Ident{Name: "handleErrorFunc"}}},
				}},
			},
		},
	)

	//	type _ORM struct {
	//		HandleErrorFunc func(ctx context.Context, err error) error
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: config.GoORMStructName()},
					Type: &ast.StructType{Fields: &ast.FieldList{List: []*ast.Field{{
						Names: []*ast.Ident{{Name: "HandleErrorFunc"}},
						Type: &ast.FuncType{
							Params: &ast.FieldList{List: []*ast.Field{
								{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
								{Names: []*ast.Ident{{Name: "err"}}, Type: &ast.Ident{Name: "error"}},
							}},
							Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}},
						},
					}}}},
				},
			},
		},
	)

	//	func (o *_ORM) HandleError(ctx context.Context, err error) error {
	//		if o.HandleErrorFunc != nil {
	//			return o.HandleErrorFunc(ctx, err)
	//		}
	//		return err
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.FuncDecl{
			Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "o"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: config.GoORMStructName()}}}}},
			Name: &ast.Ident{Name: "HandleError"},
			Type: &ast.FuncType{
				Params: &ast.FieldList{List: []*ast.Field{
					{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
					{Names: []*ast.Ident{{Name: "err"}}, Type: &ast.Ident{Name: "error"}},
				}},
				Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "error"}}}},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "o"}, Sel: &ast.Ident{Name: "HandleErrorFunc"}}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
						Body: &ast.BlockStmt{List: []ast.Stmt{
							&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
								Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "o"}, Sel: &ast.Ident{Name: "HandleErrorFunc"}},
								Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
							}}},
						}},
					},
					&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "err"}}},
				},
			},
		},
	)

	//	type QueryerContext interface {
	//		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	//		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	//		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	//		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					// Assign: token.Pos(1),
					Name: &ast.Ident{Name: queryerContextTypeName},
					Type: &ast.InterfaceType{
						Methods: &ast.FieldList{List: []*ast.Field{
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
							{
								Names: []*ast.Ident{{Name: "PrepareContext"}},
								Type: &ast.FuncType{
									Params: &ast.FieldList{List: []*ast.Field{
										{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
										{Names: []*ast.Ident{{Name: "query"}}, Type: &ast.Ident{Name: "string"}},
									}},
									Results: &ast.FieldList{List: []*ast.Field{
										{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "sql"}, Sel: &ast.Ident{Name: "Stmt"}}}},
										{Type: &ast.Ident{Name: "error"}},
									}},
								},
							},
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
						}},
					},
				},
			},
		},
	)

	//	func LoggerFromContext(ctx context.Context) *slog.Logger {
	//		if ctx == nil {
	//			return slog.Default()
	//		}
	//		if logger, ok := ctx.Value((*slog.Logger)(nil)).(*slog.Logger); ok {
	//			return logger
	//		}
	//		return slog.Default()
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.FuncDecl{
			Name: &ast.Ident{Name: "LoggerFromContext"},
			Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}}}}, Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "slog"}, Sel: &ast.Ident{Name: "Logger"}}}}}}},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "ctx"}, Op: token.EQL, Y: &ast.Ident{Name: "nil"}},
						Body: &ast.BlockStmt{List: []ast.Stmt{
							&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "slog"}, Sel: &ast.Ident{Name: "Default"}}}}},
						}},
					},
					&ast.IfStmt{
						//		if logger, ok := ctx.Value((*slog.Logger)(nil)).(*slog.Logger); ok {
						Init: &ast.AssignStmt{
							Lhs: []ast.Expr{&ast.Ident{Name: "logger"}, &ast.Ident{Name: "ok"}},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.TypeAssertExpr{
									X: &ast.CallExpr{
										Fun:  &ast.Ident{Name: "ctx.Value"},
										Args: []ast.Expr{&ast.CallExpr{Fun: &ast.ParenExpr{X: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "slog"}, Sel: &ast.Ident{Name: "Logger"}}}}, Args: []ast.Expr{&ast.Ident{Name: "nil"}}}},
									},
									Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "slog"}, Sel: &ast.Ident{Name: "Logger"}}},
								},
							},
						},
						Cond: &ast.Ident{Name: "ok"},
						Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "logger"}}}}},
					},
					&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "slog"}, Sel: &ast.Ident{Name: "Default"}}}}},
				},
			},
		},
	)

	//	func LoggerWithContext(ctx context.Context, logger *slog.Logger) context.Context {
	//		return context.WithValue(ctx, (*slog.Logger)(nil), logger)
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.FuncDecl{
			Name: &ast.Ident{Name: "LoggerWithContext"},
			Type: &ast.FuncType{Params: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}}, {Names: []*ast.Ident{{Name: "logger"}}, Type: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "slog"}, Sel: &ast.Ident{Name: "Logger"}}}}}}, Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: "context.Context"}}}}},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "context"}, Sel: &ast.Ident{Name: "WithValue"}}, Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.CallExpr{Fun: &ast.ParenExpr{X: &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: "slog"}, Sel: &ast.Ident{Name: "Logger"}}}}, Args: []ast.Expr{&ast.Ident{Name: "nil"}}}, &ast.Ident{Name: "logger"}}}}},
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
