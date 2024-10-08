package arcgengo

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
	"github.com/kunitsucom/arcgen/internal/config"
)

//nolint:cyclop,funlen,gocognit,maintidx,unparam
func generateREADContent(astFile *ast.File, arcSrcSet *ARCSourceSet) {
	for _, arcSrc := range arcSrcSet.ARCSourceSlice {
		structName := arcSrc.extractStructName()
		tableName := arcSrc.extractTableNameFromCommentGroup()
		tableInfo := arcSrc.extractFieldNamesAndColumnNames()
		columnNames := tableInfo.Columns.ColumnNames()

		{
			// const Get{StructName}ByPKQuery = `SELECT {column_name1}, {column_name2} FROM {table_name} WHERE {pk1} = ? [AND ...]`
			//
			//	func (q *query) Get{StructName}ByPK(ctx context.Context, queryer QueryerContext, pk1 pk1type, ...) ({Struct}, error) {
			//		row := queryer.QueryRowContext(ctx, Get{StructName}Query, pk1, ...)
			//		var s {Struct}
			//		if err := row.Scan(
			//			&s.{ColumnName1},
			//			&i.{ColumnName2},
			//		) err != nil {
			//			return nil, fmt.Errorf("row.Scan: %w", err)
			//		}
			//		return &s, nil
			//	}
			pks := tableInfo.Columns.PrimaryKeys()
			byPKFuncName := readOneFuncPrefix + structName + "ByPK"
			byPKQueryName := byPKFuncName + "Query"
			astFile.Decls = append(astFile.Decls,
				&ast.GenDecl{
					Tok: token.CONST,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{{Name: byPKQueryName}},
							Values: []ast.Expr{&ast.BasicLit{
								Kind:  token.STRING,
								Value: "`SELECT " + util.JoinStringsWithQuote(columnNames, ", ", `"`) + " FROM " + tableName + " WHERE " + whereColumnsPlaceholder(pks.ColumnNames(), "AND", 1) + "`",
							}},
						},
					},
				},
				&ast.FuncDecl{
					Name: &ast.Ident{Name: byPKFuncName},
					Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: receiverName}}, Type: &ast.StarExpr{X: &ast.Ident{Name: config.GoORMStructName()}}}}},
					Type: &ast.FuncType{
						Params: &ast.FieldList{
							List: append([]*ast.Field{
								{
									Names: []*ast.Ident{{Name: "ctx"}},
									Type:  &ast.Ident{Name: "context.Context"},
								},
								{
									Names: []*ast.Ident{{Name: queryerContextVarName}},
									Type:  &ast.Ident{Name: queryerContextTypeName},
								},
							},
								func() []*ast.Field {
									fields := make([]*ast.Field, 0)
									for _, pk := range pks {
										fields = append(fields, &ast.Field{
											Names: []*ast.Ident{{Name: util.PascalCaseToCamelCase(pk.FieldName)}},
											Type:  &ast.Ident{Name: pk.FieldType},
										})
									}
									return fields
								}()...),
						},
						Results: &ast.FieldList{List: []*ast.Field{
							{Type: &ast.StarExpr{X: &ast.Ident{Name: importName + "." + structName}}},
							{Type: &ast.Ident{Name: "error"}},
						}},
					},
					Body: &ast.BlockStmt{
						// row, err := queryer.QueryRowContext(ctx, Get{StructName}Query, pk1, ...)
						List: []ast.Stmt{
							&ast.ExprStmt{
								//		LoggerFromContext(ctx).Debug(queryName)
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   &ast.CallExpr{Fun: &ast.Ident{Name: "LoggerFromContext"}, Args: []ast.Expr{&ast.Ident{Name: "ctx"}}},
										Sel: &ast.Ident{Name: "Debug"},
									},
									Args: []ast.Expr{&ast.Ident{Name: byPKQueryName}},
								},
							},
							&ast.AssignStmt{
								Lhs: []ast.Expr{&ast.Ident{Name: "row"}},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{&ast.CallExpr{
									Fun: &ast.SelectorExpr{X: &ast.Ident{Name: queryerContextVarName}, Sel: &ast.Ident{Name: "QueryRowContext"}},
									Args: append(
										[]ast.Expr{
											&ast.Ident{Name: "ctx"},
											&ast.Ident{Name: byPKQueryName},
										},
										func() []ast.Expr {
											var args []ast.Expr
											for _, pk := range pks {
												args = append(args, &ast.Ident{Name: util.PascalCaseToCamelCase(pk.FieldName)})
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
									Type:  &ast.Ident{Name: importName + "." + structName},
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
									// return fmt.Errorf("row.Scan: %w", err)
									&ast.ReturnStmt{Results: []ast.Expr{
										&ast.Ident{Name: "nil"},
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
											Args: []ast.Expr{&ast.Ident{Name: strconv.Quote("row.Scan: %w")}, &ast.CallExpr{
												Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: receiverName}, Sel: &ast.Ident{Name: "HandleError"}},
												Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
											}},
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

		{
			hasOneColumnsByTag := tableInfo.HasOneTagColumnsByTag()
			for _, hasOneTag := range tableInfo.HasOneTags {
				// const Get{StructName}By{FieldName}Query = `SELECT {column_name1}, {column_name2} FROM {table_name} WHERE {column} = ? [AND ...]`
				//
				//	func (q *queryer) Get{StructName}ByColumn1[AndColumn2](ctx context.Context, queryer QueryerContext, {ColumnName} {ColumnType} [, {Column2Name} {Column2Type}]) ({Struct}Slice, error) {
				//		row := queryer.QueryRowContext(ctx, Get{StructName}Query, {ColumnName}, {Column2Name})
				//		var s {Struct}
				//		if err := row.Scan(
				//			&s.{ColumnName1},
				//			&i.{ColumnName2},
				//		) err != nil {
				//			return nil, fmt.Errorf("row.Scan: %w", err)
				//		}
				//		return &s, nil
				//	}
				byHasOneTagFuncName := readOneFuncPrefix + structName + "By" + hasOneTag
				byHasOneTagQueryName := byHasOneTagFuncName + "Query"
				hasOneColumns := hasOneColumnsByTag[hasOneTag]
				astFile.Decls = append(astFile.Decls,
					&ast.GenDecl{
						Tok: token.CONST,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{{Name: byHasOneTagQueryName}},
								Values: []ast.Expr{&ast.BasicLit{
									Kind:  token.STRING,
									Value: "`SELECT " + util.JoinStringsWithQuote(columnNames, ", ", `"`) + " FROM " + tableName + " WHERE " + whereColumnsPlaceholder(hasOneColumns.ColumnNames(), "AND", 1) + "`",
								}},
							},
						},
					},
					&ast.FuncDecl{
						Name: &ast.Ident{Name: byHasOneTagFuncName},
						Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: receiverName}}, Type: &ast.StarExpr{X: &ast.Ident{Name: config.GoORMStructName()}}}}},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: append([]*ast.Field{
									{
										Names: []*ast.Ident{{Name: "ctx"}},
										Type:  &ast.Ident{Name: "context.Context"},
									},
									{
										Names: []*ast.Ident{{Name: queryerContextVarName}},
										Type:  &ast.Ident{Name: queryerContextTypeName},
									},
								},
									func() []*ast.Field {
										fields := make([]*ast.Field, 0)
										for _, c := range hasOneColumns {
											fields = append(fields, &ast.Field{
												Names: []*ast.Ident{{Name: util.PascalCaseToCamelCase(c.FieldName)}},
												Type:  &ast.Ident{Name: c.FieldType},
											})
										}
										return fields
									}()...),
							},
							Results: &ast.FieldList{List: []*ast.Field{
								{Type: &ast.StarExpr{X: &ast.Ident{Name: importName + "." + structName}}},
								{Type: &ast.Ident{Name: "error"}},
							}},
						},
						Body: &ast.BlockStmt{
							// row, err := queryer.QueryRowContext(ctx, Get{StructName}Query, column1, ...)
							List: []ast.Stmt{
								&ast.ExprStmt{
									//		LoggerFromContext(ctx).Debug(queryName)
									X: &ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   &ast.CallExpr{Fun: &ast.Ident{Name: "LoggerFromContext"}, Args: []ast.Expr{&ast.Ident{Name: "ctx"}}},
											Sel: &ast.Ident{Name: "Debug"},
										},
										Args: []ast.Expr{&ast.Ident{Name: byHasOneTagQueryName}},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "row"}},
									Tok: token.DEFINE,
									Rhs: []ast.Expr{&ast.CallExpr{
										Fun: &ast.SelectorExpr{X: &ast.Ident{Name: queryerContextVarName}, Sel: &ast.Ident{Name: "QueryRowContext"}},
										Args: append(
											[]ast.Expr{
												&ast.Ident{Name: "ctx"},
												&ast.Ident{Name: byHasOneTagQueryName},
											},
											func() []ast.Expr {
												var args []ast.Expr
												for _, c := range hasOneColumns {
													args = append(args, &ast.Ident{Name: util.PascalCaseToCamelCase(c.FieldName)})
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
										Type:  &ast.Ident{Name: importName + "." + structName},
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
										// return nil, fmt.Errorf("row.Scan: %w", err)
										&ast.ReturnStmt{Results: []ast.Expr{
											&ast.Ident{Name: "nil"},
											&ast.CallExpr{
												Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
												Args: []ast.Expr{&ast.Ident{Name: strconv.Quote("row.Scan: %w")}, &ast.CallExpr{
													Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: receiverName}, Sel: &ast.Ident{Name: "HandleError"}},
													Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
												}},
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
		}

		{
			hasManyColumnsByTag := tableInfo.HasManyTagColumnsByTag()
			for _, hasManyTag := range tableInfo.HasManyTags {
				// const List{StructName}By{FieldName}Query = `SELECT {column_name1}, {column_name2} FROM {table_name} WHERE {pk1} = ? [AND ...]`
				//
				//	func (q *query) List{StructName}ByColumn1[AndColumn2](ctx context.Context, queryer QueryerContext, {ColumnName} {ColumnType} [, {Column2Name} {Column2Type}]) ({Struct}Slice, error) {
				//		rows, err := queryer.QueryContext(ctx, List{StructName}Query, {ColumnName}, {Column2Name})
				// 		if err != nil {
				// 			return nil, fmt.Errorf("queryer.QueryContext: %w", err)
				// 		}
				//		var ss {Struct}Slice
				//		for rows.Next() {
				//			var s {Struct}
				//			if err := rows.Scan(
				//				&i.{ColumnName1},
				//				&i.{ColumnName2},
				//			); err != nil {
				//				return nil, fmt.Errorf("rows.Scan: %w", err)
				//			}
				//			s = append(s, &i)
				//		}
				//		if err := rows.Close(); err != nil {
				//			return nil, fmt.Errorf("rows.Close: %w", err)
				//		}
				//		if err := rows.Err(); err != nil {
				//			return nil, fmt.Errorf("rows.Err: %w", err)
				//		}
				//		return ss, nil
				//	}
				structSliceType := "[]*" + importName + "." + structName
				if sliceSuffix := config.GoSliceTypeSuffix(); sliceSuffix != "" {
					structSliceType = importName + "." + structName + sliceSuffix
				}
				byHasOneTagFuncName := readManyFuncPrefix + structName + "By" + hasManyTag
				byHasOneTagQueryName := byHasOneTagFuncName + "Query"
				hasManyColumns := hasManyColumnsByTag[hasManyTag]
				astFile.Decls = append(astFile.Decls,
					&ast.GenDecl{
						Tok: token.CONST,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{{Name: byHasOneTagQueryName}},
								Values: []ast.Expr{&ast.BasicLit{
									Kind:  token.STRING,
									Value: "`SELECT " + util.JoinStringsWithQuote(columnNames, ", ", `"`) + " FROM " + tableName + " WHERE " + whereColumnsPlaceholder(hasManyColumns.ColumnNames(), "AND", 1) + "`",
								}},
							},
						},
					},
					&ast.FuncDecl{
						Name: &ast.Ident{Name: byHasOneTagFuncName},
						Recv: &ast.FieldList{List: []*ast.Field{{Names: []*ast.Ident{{Name: receiverName}}, Type: &ast.StarExpr{X: &ast.Ident{Name: config.GoORMStructName()}}}}},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: append(
									[]*ast.Field{
										{Names: []*ast.Ident{{Name: "ctx"}}, Type: &ast.Ident{Name: "context.Context"}},
										{Names: []*ast.Ident{{Name: queryerContextVarName}}, Type: &ast.Ident{Name: queryerContextTypeName}},
									},
									func() []*ast.Field {
										fields := make([]*ast.Field, 0)
										for _, c := range hasManyColumns {
											fields = append(fields, &ast.Field{Names: []*ast.Ident{{Name: util.PascalCaseToCamelCase(c.FieldName)}}, Type: &ast.Ident{Name: c.FieldType}})
										}
										return fields
									}()...,
								),
							},
							Results: &ast.FieldList{List: []*ast.Field{
								{Type: &ast.Ident{Name: structSliceType}},
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
										Args: []ast.Expr{&ast.Ident{Name: byHasOneTagQueryName}},
									},
								},
								&ast.AssignStmt{
									Lhs: []ast.Expr{&ast.Ident{Name: "rows"}, &ast.Ident{Name: "err"}},
									Tok: token.DEFINE,
									Rhs: []ast.Expr{&ast.CallExpr{
										Fun: &ast.SelectorExpr{X: &ast.Ident{Name: queryerContextVarName}, Sel: &ast.Ident{Name: "QueryContext"}},
										Args: append(
											[]ast.Expr{
												&ast.Ident{Name: "ctx"},
												&ast.Ident{Name: byHasOneTagQueryName},
											},
											func() []ast.Expr {
												var args []ast.Expr
												for _, c := range hasManyColumns {
													args = append(args, &ast.Ident{Name: util.PascalCaseToCamelCase(c.FieldName)})
												}
												return args
											}()...,
										),
									}},
								},
								// 		if err != nil {
								// 			return nil, fmt.Errorf("queryerContext.QueryContext: %w", err)
								// 		}
								&ast.IfStmt{
									Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
									Body: &ast.BlockStmt{List: []ast.Stmt{
										&ast.ReturnStmt{Results: []ast.Expr{
											&ast.Ident{Name: "nil"},
											&ast.CallExpr{
												Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
												Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(queryerContextVarName + ".QueryContext: %w")}, &ast.CallExpr{
													Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: receiverName}, Sel: &ast.Ident{Name: "HandleError"}},
													Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
												}},
											},
										}},
									}},
								},
								//		var ss {Struct}Slice
								&ast.DeclStmt{Decl: &ast.GenDecl{
									Tok: token.VAR,
									Specs: []ast.Spec{&ast.ValueSpec{
										Names: []*ast.Ident{{Name: "ss"}},
										Type:  &ast.Ident{Name: structSliceType},
									}},
								}},
								//		for rows.Next() {
								&ast.ForStmt{
									Cond: &ast.CallExpr{
										Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "rows"}, Sel: &ast.Ident{Name: "Next"}},
									},
									Body: &ast.BlockStmt{List: []ast.Stmt{
										//			var s {Struct}
										&ast.DeclStmt{Decl: &ast.GenDecl{
											Tok: token.VAR,
											Specs: []ast.Spec{&ast.ValueSpec{
												Names: []*ast.Ident{{Name: "s"}},
												Type:  &ast.Ident{Name: importName + "." + structName},
											}},
										}},
										//			if err := rows.Scan(&s.{ColumnName1}, &s.{ColumnName2}); err != nil {
										&ast.IfStmt{
											Init: &ast.AssignStmt{
												Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
												Tok: token.DEFINE,
												Rhs: []ast.Expr{&ast.CallExpr{
													Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "rows"}, Sel: &ast.Ident{Name: "Scan"}},
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
												//				return nil, fmt.Errorf("rows.Scan: %w", err)
												&ast.ReturnStmt{
													Results: []ast.Expr{
														&ast.Ident{Name: "nil"},
														&ast.CallExpr{
															Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
															Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("rows.Scan: %w")}, &ast.CallExpr{
																Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: receiverName}, Sel: &ast.Ident{Name: "HandleError"}},
																Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
															}},
														},
													},
												},
											}},
										},
										//			ss = append(ss, &s)
										&ast.AssignStmt{
											Lhs: []ast.Expr{&ast.Ident{Name: "ss"}},
											Tok: token.ASSIGN,
											Rhs: []ast.Expr{
												&ast.CallExpr{
													Fun: &ast.Ident{Name: "append"},
													Args: []ast.Expr{
														&ast.Ident{Name: "ss"},
														&ast.UnaryExpr{Op: token.AND, X: &ast.Ident{Name: "s"}},
													},
												},
											},
										},
									}},
								},
								//		if err := rows.Close(); err != nil {
								&ast.IfStmt{
									Init: &ast.AssignStmt{
										// err := rows.Close()
										Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
										Tok: token.DEFINE,
										Rhs: []ast.Expr{&ast.CallExpr{
											Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "rows"}, Sel: &ast.Ident{Name: "Close"}},
										}},
									},
									Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											//			return nil, fmt.Errorf("rows.Close: %w", err)
											&ast.ReturnStmt{
												Results: []ast.Expr{
													&ast.Ident{Name: "nil"},
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
														Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("rows.Close: %w")}, &ast.CallExpr{
															Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: receiverName}, Sel: &ast.Ident{Name: "HandleError"}},
															Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
														}},
													},
												},
											},
										},
									},
								},
								//		if err := rows.Err(); err != nil {
								&ast.IfStmt{
									Init: &ast.AssignStmt{
										// err := rows.Err()
										Lhs: []ast.Expr{&ast.Ident{Name: "err"}},
										Tok: token.DEFINE,
										Rhs: []ast.Expr{&ast.CallExpr{
											Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "rows"}, Sel: &ast.Ident{Name: "Err"}},
										}},
									},
									Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
									Body: &ast.BlockStmt{
										List: []ast.Stmt{
											//			return nil, fmt.Errorf("rows.Err: %w", err)
											&ast.ReturnStmt{
												Results: []ast.Expr{
													&ast.Ident{Name: "nil"},
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Errorf"}},
														Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("rows.Err: %w")}, &ast.CallExpr{
															Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: receiverName}, Sel: &ast.Ident{Name: "HandleError"}},
															Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.Ident{Name: "err"}},
														}},
													},
												},
											},
										},
									},
								},
								//		return ss, nil
								&ast.ReturnStmt{Results: []ast.Expr{&ast.Ident{Name: "ss"}, &ast.Ident{Name: "nil"}}},
							},
						},
					},
				)
			}
		}
	}
}
