package arcgengo

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

//nolint:funlen
func generateCREATEContent(astFile *ast.File, arcSrcSet *ARCSourceSet) {
	for _, arcSrc := range arcSrcSet.ARCSourceSlice {
		structName := arcSrc.extractStructName()
		tableName := arcSrc.extractTableNameFromCommentGroup()
		tableInfo := arcSrc.extractFieldNamesAndColumnNames()
		columnNames := tableInfo.ColumnNames()

		// const Create{StructName}Query = `INSERT INTO {table_name} ({column_name1}, {column_name2}) VALUES (?, ?)`
		//
		//	func (q *query) Create{StructName}(ctx context.Context, queryer sqlContext, s *{Struct}) error {
		//		if _, err := queryer.ExecContext(ctx, Create{StructName}Query, s.{ColumnName1}, s.{ColumnName2}); err != nil {
		//			return fmt.Errorf("q.queryer.ExecContext: %w", err)
		//		}
		//		return nil
		//	}
		funcName := "Create" + structName
		queryName := funcName + "Query"
		astFile.Decls = append(astFile.Decls,
			&ast.GenDecl{
				Tok: token.CONST,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: queryName}},
						Values: []ast.Expr{&ast.BasicLit{
							Kind:  token.STRING,
							Value: "`INSERT INTO " + tableName + " (" + strings.Join(columnNames, ", ") + ") VALUES (?" + strings.Repeat(", ?", len(columnNames)-1) + ")`",
						}},
					},
				},
			},
			&ast.FuncDecl{
				Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: "q"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "Queryer"}}}}},
				Name: &ast.Ident{Name: funcName},
				Type: &ast.FuncType{
					Params: &ast.FieldList{List: []*ast.Field{
						{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
						{Names: []*ast.Ident{{Name: "sqlCtx"}}, Type: &ast.Ident{Name: "sqlContext"}},
						{Names: []*ast.Ident{{Name: "s"}}, Type: &ast.StarExpr{X: &ast.Ident{Name: "dao." + structName}}},
					}},
					Results: &ast.FieldList{List: []*ast.Field{
						{Type: &ast.Ident{Name: "error"}},
					}},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.IfStmt{
							//		if _, err := queryer.ExecContext(ctx, Create{StructName}Query, s.{ColumnName1}, s.{ColumnName2}); err != nil {
							Init: &ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "_"}, &ast.Ident{Name: "err"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "sqlCtx"},
										Sel: &ast.Ident{Name: "ExecContext"},
									},
									Args: append(
										[]ast.Expr{
											&ast.Ident{Name: "ctx"},
											&ast.Ident{Name: queryName},
										},
										func() []ast.Expr {
											var args []ast.Expr
											for _, c := range tableInfo.Columns {
												args = append(args, &ast.SelectorExpr{X: &ast.Ident{Name: "s"}, Sel: &ast.Ident{Name: c.FieldName}})
											}
											return args
										}()...),
								}},
							},
							// err != nil {
							Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
							Body: &ast.BlockStmt{List: []ast.Stmt{
								// return fmt.Errorf("queryer.ExecContext: %w", err)
								&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
									Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
									Args: []ast.Expr{&ast.Ident{Name: strconv.Quote("queryer.ExecContext: %w")}, &ast.Ident{Name: "err"}},
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
}
