package arcgengo

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
)

//nolint:funlen
func generateCREATEContent(astFile *ast.File, arcSrcSet *ARCSourceSet) error {
	for _, arcSrc := range arcSrcSet.ARCSourceSlice {
		structName := arcSrc.extractStructName()
		tableName := arcSrc.extractTableNameFromCommentGroup()
		fieldNames, columnNames := arcSrc.extractFieldNamesAndColumnNames()

		structPackagePath, err := util.GetPackagePath(filepath.Dir(arcSrcSet.Filename))
		if err != nil {
			return errorz.Errorf("GetPackagePath: %w", err)
		}

		astFile.Decls = append(astFile.Decls,
			//	import (
			//		"context"
			//		"fmt"
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
						Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("fmt")},
					},
					&ast.ImportSpec{
						Name: &ast.Ident{Name: "dao"},
						Path: &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(structPackagePath)},
					},
				},
			},
		)

		// const Create{StructName}Query = `INSERT INTO {table_name} ({column_name1}, {column_name2}) VALUES (?, ?)`
		//
		//	// Create{StructName}
		//	func (q *query) Create{StructName}(ctx context.Context, queryer sqlQueryerContext, s *{Struct}) error {
		//		if _, err := q.queryer.ExecContext(ctx, Create{StructName}Query, s.{ColumnName1}, s.{ColumnName2}); err != nil {
		//			return fmt.Errorf("q.queryer.ExecContext: %w", err)
		//		}
		//		return nil
		//	}
		astFile.Decls = append(astFile.Decls,
			&ast.GenDecl{
				Tok: token.CONST,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: "Create" + structName + "Query"}},
						Values: []ast.Expr{&ast.BasicLit{
							Kind:  token.STRING,
							Value: "`INSERT INTO " + tableName + " (" + strings.Join(columnNames, ", ") + ") VALUES (?" + strings.Repeat(", ?", len(columnNames)-1) + ")`",
						}},
					},
				},
			},
			&ast.FuncDecl{
				Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "q"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "query"}}}}},
				Name: &ast.Ident{Name: "Create" + structName},
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
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.IfStmt{
							//		if _, err := q.queryer.ExecContext(ctx, Create{StructName}Query, s.{ColumnName1}, s.{ColumnName2}); err != nil {
							Init: &ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "_"}, &ast.Ident{Name: "err"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.SelectorExpr{X: &ast.Ident{Name: "q"}, Sel: &ast.Ident{Name: "queryer"}},
										Sel: &ast.Ident{Name: "ExecContext"},
									},
									Args: append(
										[]ast.Expr{
											&ast.Ident{Name: "ctx"},
											&ast.Ident{Name: "Create" + structName + "Query"},
										},
										func() []ast.Expr {
											var args []ast.Expr
											for _, fieldName := range fieldNames {
												args = append(args, &ast.SelectorExpr{X: &ast.Ident{Name: "s"}, Sel: &ast.Ident{Name: fieldName}})
											}
											return args
										}()...),
								}},
							},
							// err != nil {
							Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
							Body: &ast.BlockStmt{List: []ast.Stmt{
								// return fmt.Errorf("q.queryer.ExecContext: %w", err)
								&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
									Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
									Args: []ast.Expr{&ast.Ident{Name: strconv.Quote("q.queryer.ExecContext: %w")}, &ast.Ident{Name: "err"}},
								}}},
							}},
						},
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.Ident{Name: "nil"},
							},
						},
					},
				},
			},
		)
	}

	return nil
}
