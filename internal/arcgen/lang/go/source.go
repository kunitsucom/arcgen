package arcgengo

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	filepathz "github.com/kunitsucom/util.go/path/filepath"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/logs"
)

type (
	ARCSource struct {
		// Source for sorting
		Source token.Position
		// TypeSpec is used to guess the table name if the CREATE TABLE annotation is not found.
		TypeSpec *ast.TypeSpec
		// StructType is used to determine the column name. If the tag specified by --go-column-tag is not found, the field name is used.
		StructType   *ast.StructType
		CommentGroup *ast.CommentGroup
	}
	ARCSourceSet struct {
		Source         token.Position
		Filename       string
		PackageName    string
		ARCSourceSlice []*ARCSource
	}
	ARCSourceSetSlice []*ARCSourceSet
)

//nolint:gochecknoglobals
var (
	_GoColumnTagCommentLineRegex     *regexp.Regexp
	_GoColumnTagCommentLineRegexOnce sync.Once
)

const (
	//	                                             _____________ <- 1. comment prefix
	//	                                                             __ <- 2. tag name
	//	                                                                               ___ <- 4. comment suffix
	_GoColumnTagCommentLineRegexFormat       = `^\s*(//+\s*|/\*\s*)?(%s)\s*:\s*(.*)\s*(\*/)?`
	_GoColumnTagCommentLineRegexContentIndex = /*                               ^^ 3. tag value */ 3
)

func GoColumnTagCommentLineRegex() *regexp.Regexp {
	_GoColumnTagCommentLineRegexOnce.Do(func() {
		_GoColumnTagCommentLineRegex = regexp.MustCompile(fmt.Sprintf(_GoColumnTagCommentLineRegexFormat, config.GoColumnTag()))
	})
	return _GoColumnTagCommentLineRegex
}

func (a *ARCSource) extractStructName() string {
	return a.TypeSpec.Name.Name
}

func (a *ARCSource) extractTableNameFromCommentGroup() string {
	if a.CommentGroup != nil {
		for _, comment := range a.CommentGroup.List {
			if matches := util.RegexIndexTableName.Regex.FindStringSubmatch(comment.Text); len(matches) > util.RegexIndexTableName.Index {
				return strings.Trim(strings.Trim(strings.Trim(matches[util.RegexIndexTableName.Index], "`"), `"`), "'")
			}
		}
	}

	logs.Debug.Printf("WARN: table name in comment not found: `// \"%s\": table: *`: comment=%q", config.GoColumnTag(), a.CommentGroup.Text())
	return ""
}

type TableInfo struct {
	HasOneTags  []string
	HasManyTags []string
	Columns     ColumnInfos
}

type ColumnInfos []*ColumnInfo

func (ss ColumnInfos) ColumnNames() []string {
	columnNames := make([]string, len(ss))
	for i := range ss {
		columnNames[i] = ss[i].ColumnName
	}
	return columnNames
}

func (ss ColumnInfos) PrimaryKeys() ColumnInfos {
	pks := make(ColumnInfos, 0, len(ss))
	for _, column := range ss {
		if column.PK {
			pks = append(pks, column)
		}
	}
	return pks
}

func (ss ColumnInfos) NonPrimaryKeys() ColumnInfos {
	nonPks := make(ColumnInfos, 0, len(ss))
	for _, column := range ss {
		if !column.PK {
			nonPks = append(nonPks, column)
		}
	}
	return nonPks
}

func (t *TableInfo) HasOneTagColumnsByTag() map[string]ColumnInfos {
	columns := make(map[string]ColumnInfos)
	for _, hasOneTagInTable := range t.HasOneTags {
		columns[hasOneTagInTable] = make(ColumnInfos, 0, len(t.Columns))
		for _, column := range t.Columns {
			for _, hasOneTag := range column.HasOneTags {
				if hasOneTagInTable == hasOneTag {
					columns[hasOneTag] = append(columns[hasOneTag], column)
				}
			}
		}
	}

	return columns
}

func (t *TableInfo) HasManyTagColumnsByTag() map[string]ColumnInfos {
	columns := make(map[string]ColumnInfos)
	for _, hasManyTagInTable := range t.HasManyTags {
		columns[hasManyTagInTable] = make(ColumnInfos, 0, len(t.Columns))
		for _, column := range t.Columns {
			for _, hasManyTag := range column.HasManyTags {
				if hasManyTagInTable == hasManyTag {
					columns[hasManyTag] = append(columns[hasManyTag], column)
				}
			}
		}
	}

	return columns
}

type ColumnInfo struct {
	FieldName   string
	FieldType   string
	ColumnName  string
	PK          bool
	HasOneTags  []string
	HasManyTags []string
}

func columnValuesPlaceholder(columns []string) string {
	switch config.Dialect() {
	case "mysql":
		// ?, ?, ?, ...
		return "?" + strings.Repeat(", ?", len(columns)-1)
	default:
		return func() string {
			// $1, $2, $3, ...
			var s strings.Builder
			s.WriteString("$1")
			for i := 2; i <= len(columns); i++ {
				s.WriteString(", $")
				s.WriteString(strconv.Itoa(i))
			}
			return s.String()
		}()
	}
}

//nolint:unparam
func whereColumnsPlaceholder(columns []string, op string) string {
	switch config.Dialect() {
	case "mysql":
		// column1 = ? AND column2 = ? AND column3 = ...
		return strings.Join(columns, " = ? "+op+" ") + " = ?"
	default:
		return func() string {
			// column1 = $1 AND column2 = $2 AND column3 = ...
			var s strings.Builder
			for i, column := range columns {
				if i > 0 {
					s.WriteString(" " + op + " ")
				}
				s.WriteString(column)
				s.WriteString(" = $")
				s.WriteString(strconv.Itoa(i + 1))
			}
			return s.String()
		}()
	}
}

func fieldName(x ast.Expr) *ast.Ident {
	switch t := x.(type) {
	case *ast.Ident:
		return t
	case *ast.SelectorExpr:
		if _, ok := t.X.(*ast.Ident); ok {
			return t.Sel
		}
	case *ast.StarExpr:
		return fieldName(t.X)
	}
	return nil
}

//nolint:cyclop
func (a *ARCSource) extractFieldNamesAndColumnNames() *TableInfo {
	tableInfo := &TableInfo{
		Columns: make([]*ColumnInfo, 0, len(a.StructType.Fields.List)),
	}
	for _, field := range a.StructType.Fields.List {
		if field.Tag != nil {
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			// db tag
			switch columnName := tag.Get(config.GoColumnTag()); columnName {
			case "", "-":
				logs.Trace.Printf("SKIP: %s: field.Names=%s, columnName=%q", a.Source.String(), field.Names, columnName)
				// noop
			default:
				logs.Trace.Printf("%s: field.Names=%s, columnName=%q", a.Source.String(), field.Names, columnName)
				columnInfo := &ColumnInfo{
					FieldName:  field.Names[0].Name,
					FieldType:  fieldName(field.Type).String(),
					ColumnName: columnName,
				}
				// pk tag
				switch pk := tag.Get(config.GoPKTag()); pk {
				case "", "-":
					logs.Trace.Printf("SKIP: %s: field.Names=%s, pk=%q", a.Source.String(), field.Names, pk)
					// noop
				default:
					logs.Trace.Printf("%s: field.Names=%s, pk=%q", a.Source.String(), field.Names, pk)
					columnInfo.PK = true
				}
				// hasOne tag
				for _, hasOneTag := range strings.Split(tag.Get(config.GoHasOneTag()), ",") {
					if hasOneTag != "" {
						logs.Trace.Printf("%s: field.Names=%s, hasOneTag=%q", a.Source.String(), field.Names, hasOneTag)
						tableInfo.HasOneTags = append(tableInfo.HasOneTags, hasOneTag)
						columnInfo.HasOneTags = append(columnInfo.HasOneTags, hasOneTag)
					}
				}
				// hasMany tag
				for _, hasManyTag := range strings.Split(tag.Get(config.GoHasManyTag()), ",") {
					if hasManyTag != "" {
						logs.Trace.Printf("%s: field.Names=%s, hasManyTag=%q", a.Source.String(), field.Names, hasManyTag)
						tableInfo.HasManyTags = append(tableInfo.HasManyTags, hasManyTag)
						columnInfo.HasManyTags = append(columnInfo.HasManyTags, hasManyTag)
					}
				}

				tableInfo.Columns = append(tableInfo.Columns, columnInfo)
			}
		}
	}

	slices.Sort(tableInfo.HasOneTags)
	tableInfo.HasOneTags = slices.Compact(tableInfo.HasOneTags)

	return tableInfo
}

func (ss *ARCSourceSet) generateGoFileHeader() string {
	return generateGoFileHeader() +
		"// source: " + filepathz.Short(ss.Source.Filename) + "\n"
}

func generateGoFileHeader() string {
	return "" +
		"// Code generated by arcgen. DO NOT EDIT." + "\n" +
		"//" + "\n"
}
