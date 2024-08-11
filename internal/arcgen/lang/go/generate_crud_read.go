package arcgengo

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

//nolint:funlen,unparam
func generateREADContent(astFile *ast.File, arcSrcSet *ARCSourceSet) {
	for _, arcSrc := range arcSrcSet.ARCSourceSlice {
		structName := arcSrc.extractStructName()
		tableName := arcSrc.extractTableNameFromCommentGroup()
		tableInfo := arcSrc.extractFieldNamesAndColumnNames()
		columnNames := tableInfo.ColumnNames()

		// const Find{StructName}Query = `SELECT {column_name1}, {column_name2} FROM {table_name} WHERE {pk1} = ? [AND ...]`
		//
		//	// Find{StructName}
		//	func (q *query) Find{StructName}(ctx context.Context, queryer sqlQueryerContext, pk1 pk1type, ...) ({Struct}, error) {
		//		row := queryer.QueryRowContext(ctx, Find{StructName}Query, pk1, ...)
		//		var s {Struct}
		//		if err := row.Scan(
		//			&s.{ColumnName1},
		//			&i.{ColumnName2},
		//		) err != nil {
		//			return nil, fmt.Errorf("row.Scan: %w", err)
		//		}
		//		return &s, nil
		//	}
		pks := arcSrc.extractFieldNamesAndColumnNames().PrimaryKeys()
		astFile.Decls = append(astFile.Decls,
			&ast.GenDecl{
				Tok: token.CONST,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: "Find" + structName + "Query"}},
						Values: []ast.Expr{&ast.BasicLit{
							Kind: token.STRING,
							Value: "`SELECT " + strings.Join(columnNames, ", ") + " FROM " + tableName + " WHERE " + func() string {
								var where []string
								for _, pk := range pks {
									where = append(where, pk.ColumnName+" = ?")
								}
								return strings.Join(where, " AND ")
							}() + "`",
						}},
					},
				},
			},
			&ast.FuncDecl{
				Name: &ast.Ident{Name: "Find" + structName},
				Recv: &ast.FieldList{List: []*ast.Field{{
					Names: []*ast.Ident{{Name: "q"}},
					Type:  &ast.StarExpr{X: &ast.Ident{Name: "query"}},
				}}},
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: append([]*ast.Field{
							{
								Names: []*ast.Ident{{Name: "ctx"}},
								Type:  &ast.Ident{Name: "context.Context"},
							},
							{
								Names: []*ast.Ident{{Name: "queryer"}},
								Type:  &ast.Ident{Name: "sqlQueryerContext"},
							},
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
				Body: &ast.BlockStmt{
					// row, err := queryer.QueryRowContext(ctx, Find{StructName}Query, pk1, ...)
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{&ast.Ident{Name: "row"}},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{&ast.CallExpr{
								Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "queryer"}, Sel: &ast.Ident{Name: "QueryRowContext"}},
								Args: append(
									[]ast.Expr{
										&ast.Ident{Name: "ctx"},
										&ast.Ident{Name: "Find" + structName + "Query"},
									},
									func() []ast.Expr {
										var args []ast.Expr
										for _, c := range tableInfo.PrimaryKeys() {
											args = append(args, &ast.Ident{Name: c.FieldName})
										}
										return args
									}()...),
							}},
						},
						// var s {Struct}
						&ast.DeclStmt{Decl: &ast.GenDecl{
							Tok: token.VAR,
							Specs: []ast.Spec{&ast.ValueSpec{
								Names: []*ast.Ident{{Name: "s"}},
								Type:  &ast.Ident{Name: "dao." + structName},
							}},
						}},
						// if err := row.Scan(&s.{ColumnName1}, &s.{ColumnName2}); err != nil {
						&ast.IfStmt{
							Init: &ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
								Tok: token.DEFINE,
								// row.Scan(&s.{ColumnName1}, &s.{ColumnName2})
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "row"}, Sel: &ast.Ident{Name: "Scan"}},
									Args: func() []ast.Expr {
										var args []ast.Expr
										for _, c := range tableInfo.Columns {
											args = append(args, &ast.UnaryExpr{
												Op: token.AND,
												X: &ast.SelectorExpr{
													X:   &ast.Ident{Name: "s"},
													Sel: &ast.Ident{Name: c.FieldName},
												},
											})
										}
										return args
									}(),
								}},
							},
							Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
							Body: &ast.BlockStmt{List: []ast.Stmt{
								// return fmt.Errorf("queryer.ExecContext: %w", err)
								&ast.ReturnStmt{Results: []ast.Expr{
									&ast.Ident{Name: "nil"},
									&ast.CallExpr{
										Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
										Args: []ast.Expr{&ast.Ident{Name: strconv.Quote("row.Scan: %w")}, &ast.Ident{Name: "err"}},
									},
								}},
							}},
						},
						// return &s, nil
						&ast.ReturnStmt{Results: []ast.Expr{&ast.UnaryExpr{Op: token.AND, X: &ast.Ident{Name: "s"}}, &ast.Ident{Name: "nil"}}},
					},
				},
			},
		)
	}

	return
}
