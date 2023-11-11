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
	"reflect"
	"strconv"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"
	filepathz "github.com/kunitsucom/util.go/path/filepath"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/logs"
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

func generate(arcSrcSets ARCSourceSets) error {
	for _, arcSrcSet := range arcSrcSets {
		filePrefix := strings.TrimSuffix(arcSrcSet.Filename, fileSuffix)
		filename := fmt.Sprintf("%s.%s.gen%s", filePrefix, config.ColumnTagGo(), fileSuffix)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return errorz.Errorf("os.OpenFile: %w", err)
		}

		if err := fprint(f, bytes.NewBuffer(nil), arcSrcSet); err != nil {
			return errorz.Errorf("sprint: %w", err)
		}
	}

	return nil
}

type buffer interface {
	io.Writer
	fmt.Stringer
}

func fprint(osFile io.Writer, buf buffer, arcSrcSet *ARCSourceSet) error {
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

	for _, arcSrc := range arcSrcSet.ARCSources {
		structName := arcSrc.TypeSpec.Name.Name
		tableName := extractTableNameFromCommentGroup(arcSrc.CommentGroup)
		fieldNames, columnNames := make([]string, 0), make([]string, 0)
		for _, field := range arcSrc.StructType.Fields.List {
			if field.Tag != nil {
				tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
				switch columnName := tag.Get(config.ColumnTagGo()); columnName {
				case "", "-":
					logs.Trace.Printf("SKIP: %s: field.Names=%s, columnName=%q", arcSrc.Source.String(), field.Names, columnName)
					// noop
				default:
					logs.Trace.Printf("%s: field.Names=%s, columnName=%q", arcSrc.Source.String(), field.Names, columnName)
					fieldNames, columnNames = append(fieldNames, field.Names[0].Name), append(columnNames, columnName)
				}
			}
		}

		appendAST(astFile, structName, tableName, config.MethodNameTable(), config.MethodNameColumns(), config.MethodPrefixColumn(), fieldNames, columnNames)
	}

	if err := printer.Fprint(buf, token.NewFileSet(), astFile); err != nil {
		return errorz.Errorf("printer.Fprint: %w", err)
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

	return nil
}

func extractTableNameFromCommentGroup(commentGroup *ast.CommentGroup) string {
	if commentGroup != nil {
		for _, comment := range commentGroup.List {
			if matches := util.RegexIndexTableName.Regex.FindStringSubmatch(comment.Text); len(matches) > util.RegexIndexTableName.Index {
				return strings.Trim(matches[util.RegexIndexTableName.Index], "`")
			}
		}
	}

	logs.Debug.Printf("WARN: table name in comment not found: `// \"%s\": table: *`: comment=%q", config.ColumnTagGo(), commentGroup.Text())
	return ""
}

//nolint:funlen
func appendAST(file *ast.File, structName string, tableName string, methodNameTable string, methodNameColumns string, methodPrefixColumn string, fieldNames, columnNames []string) {
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
						Type: &ast.Ident{
							Name: structName, // MEMO: struct name
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
	}

	file.Decls = append(file.Decls, generateASTColumnMethods(structName, methodNameColumns, methodPrefixColumn, fieldNames, columnNames)...)

	return //nolint:gosimple
}

//nolint:funlen
func generateASTColumnMethods(structName string, methodNameColumns string, prefixColumn string, fieldNames, columnNames []string) []ast.Decl {
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
					Type: &ast.Ident{
						Name: structName, // MEMO: struct name
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
	})

	// each column name methods
	for i, columnName := range columnNames {
		decls = append(decls, &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: "s",
							},
						},
						Type: &ast.Ident{
							Name: structName, // MEMO: struct name
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
