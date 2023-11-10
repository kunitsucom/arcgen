package arcgengo

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"
	filepathz "github.com/kunitsucom/util.go/path/filepath"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/logs"
)

//nolint:cyclop
func Generate(ctx context.Context, src string) error {
	arcSrcSets, err := parse(ctx, src)
	if err != nil {
		return errorz.Errorf("parse: %w", err)
	}

	newFile := token.NewFileSet()

	for _, arcSrcSet := range arcSrcSets {
		filePrefix := strings.TrimSuffix(arcSrcSet.Filename, fileSuffix)
		filename := fmt.Sprintf("%s.%s.gen%s", filePrefix, config.ColumnTagGo(), fileSuffix)
		osFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return errorz.Errorf("os.OpenFile: %w", err)
		}

		astFile := &ast.File{
			// package
			Name: &ast.Ident{
				Name: arcSrcSet.PackageName,
			},
			// methods
			Decls: []ast.Decl{},
		}

		for _, arcSrc := range arcSrcSet.ARCSources {
			structName := arcSrc.TypeSpec.Name.Name
			tableName := extractTableNameFromCommentGroup(arcSrc.CommentGroup)
			columnNames := func() []string {
				columnNames := make([]string, 0)
				for _, field := range arcSrc.StructType.Fields.List {
					if field.Tag != nil {
						tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
						switch columnName := tag.Get(config.ColumnTagGo()); columnName {
						case "", "-":
							logs.Trace.Printf("SKIP: %s: field.Names=%s, columnName=%q", arcSrc.Source.String(), field.Names, columnName)
							// noop
						default:
							columnNames = append(columnNames, columnName)
						}
					}
				}
				return columnNames
			}()

			appendAST(astFile, structName, tableName, config.MethodPrefixGlobal(), config.MethodPrefixColumn(), columnNames)
		}

		buf := bytes.NewBuffer(nil)
		if err := format.Node(buf, newFile, astFile); err != nil {
			return errorz.Errorf("format.Node: %w", err)
		}

		// add header comment
		content := "" +
			"// Code generated by arcgen. DO NOT EDIT." + "\n" +
			"//" + "\n" +
			fmt.Sprintf("// source: %s", filepathz.Short(arcSrcSet.Source.Filename)) + "\n" +
			"\n" +
			buf.String()

		// add blank line between methods
		content = strings.ReplaceAll(content, "\n}\nfunc ", "\n}\n\nfunc ")

		// write to file
		if _, err := io.WriteString(osFile, content); err != nil {
			return errorz.Errorf("io.WriteString: %w", err)
		}
	}

	return nil
}

func extractTableNameFromCommentGroup(commentGroup *ast.CommentGroup) string {
	if commentGroup != nil {
		for _, comment := range commentGroup.List {
			if matches := util.RegexIndexTableName.Regex.FindStringSubmatch(comment.Text); len(matches) > util.RegexIndexTableName.Index {
				return matches[util.RegexIndexTableName.Index]
			}
		}
	}

	logs.Debug.Printf("WARN: table name in comment not found: `// \"%s\": table: *`: comment=%q", config.ColumnTagGo(), commentGroup.Text())
	return ""
}

//nolint:funlen
func appendAST(file *ast.File, structName string, tableName string, prefixGlobal string, prefixColumn string, columnNames []string) {
	if tableName != "" {
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
				Name: prefixGlobal + "TableName",
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
	}

	file.Decls = append(file.Decls, generateASTColumnMethods(structName, prefixGlobal, prefixColumn, columnNames)...)

	return
}

//nolint:funlen
func generateASTColumnMethods(structName string, prefixGlobal string, prefixColumn string, columnNames []string) []ast.Decl {
	decls := make([]ast.Decl, 0)

	// all column names method
	elts := make([]ast.Expr, 0)
	for _, columnName := range columnNames {
		elts = append(elts, &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(columnName),
		})
	}
	decls = append(decls, &ast.FuncDecl{
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
			Name: prefixGlobal + "ColumnNames",
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
	})

	// each column name methods
	for _, columnName := range columnNames {
		decls = append(decls, &ast.FuncDecl{
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
				Name: prefixGlobal + prefixColumn + columnName,
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
								Name: strconv.Quote(columnName),
							},
						},
					},
				},
			},
		})
	}

	return decls
}
