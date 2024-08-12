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

func fprintCRUDCommon(osFile osFile, buf buffer, arcSrcSetSlice ARCSourceSetSlice, crudFiles []string) error {
	content, err := generateCRUDCommonFileContent(buf, arcSrcSetSlice, crudFiles)
	if err != nil {
		return errorz.Errorf("generateCRUDCommonFileContent: %w", err)
	}

	// write to file
	if _, err := io.WriteString(osFile, content); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}

	return nil
}

const (
	sqlQueryerContextVarName  = "sqlContext"
	sqlQueryerContextTypeName = "sqlQueryerContext"
)

//nolint:cyclop,funlen,gocognit,maintidx
func generateCRUDCommonFileContent(buf buffer, arcSrcSetSlice ARCSourceSetSlice, crudFiles []string) (string, error) {
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
		//		"log/slog"
		//
		//		dao "path/to/your/dao"
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
					// Assign: token.Pos(1),
					Name: &ast.Ident{Name: sqlQueryerContextTypeName},
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

	//	type _CRUD struct {
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{Name: config.GoCRUDTypeNameUnexported()},
					Type: &ast.StructType{Fields: &ast.FieldList{}},
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

	//	type CRUD interface {
	//		Create{StructName}(ctx context.Context, sqlQueryer sqlQueryerContext, s *{Struct}) error
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
							if ident.Name == config.GoCRUDTypeNameUnexported() {
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
					Name: &ast.Ident{Name: config.GoCRUDTypeName()},
					Type: &ast.InterfaceType{
						Methods: &ast.FieldList{List: methods},
					},
				},
			},
		},
	)

	//	func NewCRUD() CRUD {
	//		return &_CRUD{}
	//	}
	astFile.Decls = append(astFile.Decls,
		&ast.FuncDecl{
			Name: &ast.Ident{Name: "New" + config.GoCRUDTypeName()},
			Type: &ast.FuncType{Results: &ast.FieldList{List: []*ast.Field{{Type: &ast.Ident{Name: config.GoCRUDTypeName()}}}}},
			Body: &ast.BlockStmt{List: []ast.Stmt{&ast.ReturnStmt{Results: []ast.Expr{&ast.UnaryExpr{Op: token.AND, X: &ast.Ident{Name: config.GoCRUDTypeNameUnexported() + "{}"}}}}}},
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
