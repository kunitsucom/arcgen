package arcgengo

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"os"
	"strconv"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"

	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/pkg/errors"
)

//nolint:cyclop,funlen
func Generate(ctx context.Context, src string) error {
	arcSrcSets, err := parse(ctx, src)
	if err != nil {
		return errorz.Errorf("parse: %w", err)
	}

	if err := generate(arcSrcSets); err != nil {
		return errorz.Errorf("generate: %w", err)
	}

	return nil
}

func generate(arcSrcSets ARCSourceSetSlice) error {
	for _, arcSrcSet := range arcSrcSets {
		const rw_r__r__ = 0o644 //nolint:revive,stylecheck // rw-r--r--

		generateColumns := func() error {
			filePathWithoutExt := strings.TrimSuffix(arcSrcSet.Filename, fileExt)
			newExt := fmt.Sprintf(".%s.gen%s", config.GoColumnTag(), fileExt)
			filename := filePathWithoutExt + newExt
			f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, rw_r__r__)
			if err != nil {
				return errorz.Errorf("os.OpenFile: %w", err)
			}
			defer f.Close()

			if err := fprintColumns(f, bytes.NewBuffer(nil), arcSrcSet); err != nil {
				return errorz.Errorf("sprint: %w", err)
			}
			return nil
		}

		if err := generateColumns(); err != nil {
			return errorz.Errorf("f: %w", err)
		}
	}

	return nil
}

type buffer interface {
	io.Writer
	fmt.Stringer
}

func fprintColumns(osFile io.Writer, buf buffer, arcSrcSet *ARCSourceSet) error {
	if arcSrcSet == nil || arcSrcSet.PackageName == "" {
		return errors.ErrInvalidSourceSet
	}
	astFile := &ast.File{
		// package
		Name: &ast.Ident{
			Name: arcSrcSet.PackageName,
		},
		// methods
		Decls: []ast.Decl{},
	}

	for _, arcSrc := range arcSrcSet.ARCSourceSlice {
		structName := arcSrc.extractStructName()
		tableName := arcSrc.extractTableNameFromCommentGroup()
		fieldNames, columnNames := arcSrc.extractFieldNamesAndColumnNames()

		appendAST(
			astFile,
			structName,
			config.GoSliceTypeSuffix(),
			tableName,
			config.GoMethodNameTable(),
			config.GoMethodNameColumns(),
			config.GoMethodPrefixColumn(),
			fieldNames,
			columnNames,
		)
	}

	if err := printer.Fprint(buf, token.NewFileSet(), astFile); err != nil {
		return errorz.Errorf("printer.Fprint: %w", err)
	}

	// add header comment
	content := arcSrcSet.generateGoFileHeader() + buf.String()

	// add blank line between methods
	content = strings.ReplaceAll(content, "\n}\nfunc ", "\n}\n\nfunc ")

	// write to file
	if _, err := io.WriteString(osFile, content); err != nil {
		return errorz.Errorf("io.WriteString: %w", err)
	}

	return nil
}

//nolint:funlen
func appendAST(file *ast.File, structName string, sliceTypeSuffix string, tableName string, methodNameTable string, methodNameColumns string, methodPrefixColumn string, fieldNames, columnNames []string) {
	if tableName != "" {
		//	func (s *StructName) TableName() string {
		//		return "TableName"
		//	}
		file.Decls = append(file.Decls, &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: "s",
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: structName, // MEMO: struct name
							},
						},
					},
				},
			},
			Name: &ast.Ident{
				Name: methodNameTable,
			},
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.Ident{
								Name: "string",
							},
						},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: strconv.Quote(tableName),
							},
						},
					},
				},
			},
		})

		// type StructNameSlice []*StructName
		if sliceTypeSuffix != "" {
			file.Decls = append(
				file.Decls,
				&ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: structName + sliceTypeSuffix,
							},
							Type: &ast.ArrayType{
								Elt: &ast.StarExpr{
									X: &ast.Ident{
										Name: structName, // MEMO: struct name
									},
								},
							},
						},
					},
				},
				//	func (s StructNameSlice) TableName() string {
				//		return "TableName"
				//	}
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{
										Name: "s",
									},
								},
								Type: &ast.Ident{
									Name: structName + sliceTypeSuffix,
								},
							},
						},
					},
					Name: &ast.Ident{
						Name: methodNameTable,
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{
										Name: "string",
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{
										Name: strconv.Quote(tableName),
									},
								},
							},
						},
					},
				},
			)
		}
	}

	file.Decls = append(file.Decls, generateASTColumnMethods(structName, sliceTypeSuffix, methodNameColumns, methodPrefixColumn, fieldNames, columnNames)...)

	return //nolint:gosimple
}

//nolint:funlen
func generateASTColumnMethods(structName string, sliceTypeSuffix string, methodNameColumns string, prefixColumn string, fieldNames, columnNames []string) []ast.Decl {
	decls := make([]ast.Decl, 0)

	// all column names method
	elts := make([]ast.Expr, 0)
	for _, columnName := range columnNames {
		elts = append(elts, &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(columnName),
		})
	}
	decls = append(decls,
		//	func (s *StructName) Columns() []string {
		//		return []string{"column1", "column2", ...}
		//	}
		&ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: "s",
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: structName, // MEMO: struct name
							},
						},
					},
				},
			},
			Name: &ast.Ident{
				Name: methodNameColumns,
			},
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.Ident{
								Name: "[]string",
							},
						},
					},
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CompositeLit{
								Type: &ast.ArrayType{
									Elt: &ast.Ident{
										Name: "string",
									},
								},
								Elts: elts,
							},
						},
					},
				},
			},
		},
	)

	// each column name methods
	for i := range columnNames {
		decls = append(decls,
			//	func (s *StructName) Column1() string {
			//		return "column1"
			//	}
			&ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "s",
								},
							},
							Type: &ast.StarExpr{
								X: &ast.Ident{
									Name: structName, // MEMO: struct name
								},
							},
						},
					},
				},
				Name: &ast.Ident{
					Name: prefixColumn + fieldNames[i],
				},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.Ident{
									Name: "string",
								},
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.Ident{
									Name: strconv.Quote(columnNames[i]),
								},
							},
						},
					},
				},
			},
		)
	}

	if sliceTypeSuffix != "" {
		//	func (s StructNameSlice) Columns() []string {
		//		return []string{"column1", "column2", ...}
		//	}
		decls = append(decls,
			&ast.FuncDecl{
				Recv: &ast.FieldList{
					List: []*ast.Field{
						{
							Names: []*ast.Ident{
								{
									Name: "s",
								},
							},
							Type: &ast.Ident{
								Name: structName + sliceTypeSuffix,
							},
						},
					},
				},
				Name: &ast.Ident{
					Name: methodNameColumns,
				},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.Ident{
									Name: "[]string",
								},
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.ArrayType{
										Elt: &ast.Ident{
											Name: "string",
										},
									},
									Elts: elts,
								},
							},
						},
					},
				},
			},
		)

		// each column name methods
		for i := range columnNames {
			decls = append(decls,
				//	func (s StructNameSlice) Column1() string {
				//		return "column1"
				//	}
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{
										Name: "s",
									},
								},
								Type: &ast.Ident{
									Name: structName + sliceTypeSuffix,
								},
							},
						},
					},
					Name: &ast.Ident{
						Name: prefixColumn + fieldNames[i],
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{
										Name: "string",
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{
										Name: strconv.Quote(columnNames[i]),
									},
								},
							},
						},
					},
				},
			)
		}
	}

	return decls
}
