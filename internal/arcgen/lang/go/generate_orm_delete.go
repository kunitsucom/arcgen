package arcgengo

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
	"github.com/kunitsucom/arcgen/internal/config"
)

//nolint:funlen
func generateDELETEContent(astFile *ast.File, arcSrcSet *ARCSourceSet) {
	for _, arcSrc := range arcSrcSet.ARCSourceSlice {
		structName := arcSrc.extractStructName()
		tableName := arcSrc.extractTableNameFromCommentGroup()
		tableInfo := arcSrc.extractFieldNamesAndColumnNames()

		// const Delete{StructName}Query = `DELETE FROM {table_name} WHERE {pk1} = ? [AND {pk2} = ?]`
		//
		//	func (q *query) Delete{StructName}(ctx context.Context, queryer QueryerContext, pk1 pk1type [, pk2 pk2type]) error {
		//		if _, err := queryerContext.ExecContext(ctx, Delete{StructName}Query, pk1 [, pk2]); err != nil {
		//			return fmt.Errorf("queryerContext.ExecContext: %w", err)
		//		}
		//		return nil
		//	}
		funcName := "Delete" + structName + "ByPK"
		queryName := funcName + "Query"
		pkColumns := tableInfo.Columns.PrimaryKeys()
		pkColumnNames := func() (pkColumnNames []string) {
			for _, c := range pkColumns {
				pkColumnNames = append(pkColumnNames, c.ColumnName)
			}
			return pkColumnNames
		}()
		astFile.Decls = append(astFile.Decls,
			&ast.GenDecl{
				Tok: token.CONST,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{{Name: queryName}},
						Values: []ast.Expr{&ast.BasicLit{
							Kind:  token.STRING,
							Value: "`DELETE FROM " + tableName + " WHERE " + whereColumnsPlaceholder(pkColumnNames, "AND", 1) + "`",
						}},
					},
				},
			},
			&ast.FuncDecl{
				Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: receiverName}}, Type: &ast.StarExpr{X: &ast.Ident{Name: config.GoORMStructName()}}}}},
				Name: &ast.Ident{Name: funcName},
				Type: &ast.FuncType{
					Params: &ast.FieldList{List: append(
						[]*ast.Field{
							{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
							{Names: []*ast.Ident{{Name: queryerContextVarName}}, Type: &ast.Ident{Name: queryerContextTypeName}},
						},
						func() []*ast.Field {
							var fields []*ast.Field
							for _, c := range pkColumns {
								fields = append(fields, &ast.Field{Names: []*ast.Ident{{Name: util.PascalCaseToCamelCase(c.FieldName)}}, Type: &ast.Ident{Name: c.FieldType}})
							}
							return fields
						}()...,
					)},
					Results: &ast.FieldList{List: []*ast.Field{
						{Type: &ast.Ident{Name: "error"}},
					}},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							//		LoggerFromContext(ctx).Debug(queryName)
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   &ast.CallExpr{Fun: &ast.Ident{Name: "LoggerFromContext"}, Args: []ast.Expr{&ast.Ident{Name: "ctx"}}},
									Sel: &ast.Ident{Name: "Debug"},
								},
								Args: []ast.Expr{&ast.Ident{Name: queryName}},
							},
						},
						&ast.IfStmt{
							//		if _, err := queryerContext.ExecContext(ctx, Delete{StructName}Query, pk1 [, pk2]); err != nil {
							Init: &ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "_"}, &ast.Ident{Name: "err"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.Ident{Name: queryerContextVarName},
										Sel: &ast.Ident{Name: "ExecContext"},
									},
									Args: append(
										[]ast.Expr{
											&ast.Ident{Name: "ctx"},
											&ast.Ident{Name: queryName},
										},
										func() []ast.Expr {
											var args []ast.Expr
											for _, c := range pkColumns {
												args = append(args, &ast.Ident{Name: util.PascalCaseToCamelCase(c.FieldName)})
											}
											return args
										}()...,
									),
								}},
							},
							// err != nil {
							Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
							Body: &ast.BlockStmt{List: []ast.Stmt{
								// return fmt.Errorf("queryerContext.ExecContext: %w", err)
								&ast.ReturnStmt{Results: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
									Args: []ast.Expr{&ast.Ident{Name: strconv.Quote(queryerContextVarName + ".ExecContext: %w")}, &ast.CallExpr{
										Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: receiverName}, Sel: &ast.Ident{Name: "HandleError"}},
										Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
									}},
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
