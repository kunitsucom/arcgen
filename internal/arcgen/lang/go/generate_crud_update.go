package arcgengo

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

//nolint:funlen
func generateUPDATEContent(astFile *ast.File, arcSrcSet *ARCSourceSet) {
	for _, arcSrc := range arcSrcSet.ARCSourceSlice {
		structName := arcSrc.extractStructName()
		tableName := arcSrc.extractTableNameFromCommentGroup()
		tableInfo := arcSrc.extractFieldNamesAndColumnNames()

		// const Update{StructName}Query = `UPDATE {table_name} SET ({column_name1}, {column_name2}) = (?, ?) WHERE {pk1} = ? [AND {pk2} = ?]`
		//
		//	func (q *query) Update{StructName}(ctx context.Context, queryer sqlContext, s *{Struct}) error {
		//		if _, err := queryer.ExecContext(ctx, Update{StructName}Query, s.{ColumnName1}, s.{ColumnName2}, s.{PK1} [, s.{PK2}]); err != nil {
		//			return fmt.Errorf("q.queryer.ExecContext: %w", err)
		//		}
		//		return nil
		//	}
		funcName := "Update" + structName
		queryName := funcName + "Query"
		pkColumns := tableInfo.PrimaryKeys()
		pkColumnNames := func() (pkColumnNames []string) {
			for _, c := range pkColumns {
				pkColumnNames = append(pkColumnNames, c.ColumnName)
			}
			return pkColumnNames
		}()
		nonPKColumns := tableInfo.NonPrimaryKeys()
		nonPKColumnNames := func() (nonPKColumnNames []string) {
			for _, c := range nonPKColumns {
				nonPKColumnNames = append(nonPKColumnNames, c.ColumnName)
			}
			return nonPKColumnNames
		}()
		astFile.Decls = append(astFile.Decls,
			&ast.GenDecl{
				Tok: token.CONST,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: queryName}},
						Values: []ast.Expr{&ast.BasicLit{
							Kind:  token.STRING,
							Value: "`UPDATE " + tableName + " SET (" + strings.Join(nonPKColumnNames, ", ") + ") = (?" + strings.Repeat(", ?", len(nonPKColumns)-1) + ") WHERE " + strings.Join(pkColumnNames, " = ? AND ") + " = ?`",
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
							//		if _, err := queryer.ExecContext(ctx, Update{StructName}Query, s.{ColumnName1}, s.{ColumnName2}, s.{PK1} [, s.{PK2}]); err != nil {
							Init: &ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "_"}, &ast.Ident{Name: "err"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: "sqlCtx"},
										Sel: &ast.Ident{Name: "ExecContext"},
									},
									Args: append(
										append(
											[]ast.Expr{
												&ast.Ident{Name: "ctx"},
												&ast.Ident{Name: queryName},
											},
											func() []ast.Expr {
												var args []ast.Expr
												for _, c := range nonPKColumns {
													args = append(args, &ast.SelectorExpr{X: &ast.Ident{Name: "s"}, Sel: &ast.Ident{Name: c.FieldName}})
												}
												return args
											}()...),
										func() []ast.Expr {
											var args []ast.Expr
											for _, c := range pkColumns {
												args = append(args, &ast.SelectorExpr{X: &ast.Ident{Name: "s"}, Sel: &ast.Ident{Name: c.FieldName}})
											}
											return args
										}()...,
									),
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
