package arcgengo

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"sync"

	"github.com/kunitsucom/arcgen/internal/config"
)

type (
	ARCSource struct {
		// Source for sorting
		Source token.Position
		// TypeSpec is used to guess the table name if the CREATE TABLE annotation is not found.
		TypeSpec *ast.TypeSpec
		// StructType is used to determine the column name. If the tag specified by --column-tag-go is not found, the field name is used.
		StructType   *ast.StructType
		CommentGroup *ast.CommentGroup
	}
	ARCSourceSet struct {
		Source      token.Position
		Filename    string
		PackageName string
		ARCSources  []*ARCSource
	}
	ARCSourceSets []*ARCSourceSet
)

//nolint:gochecknoglobals
var (
	_ColumnTagGoCommentLineRegex     *regexp.Regexp
	_ColumnTagGoCommentLineRegexOnce sync.Once
)

const (
	//	                                             _____________ <- 1. comment prefix
	//	                                                             __ <- 2. tag name
	//	                                                                               ___ <- 4. comment suffix
	_ColumnTagGoCommentLineRegexFormat       = `^\s*(//+\s*|/\*\s*)?(%s)\s*:\s*(.*)\s*(\*/)?`
	_ColumnTagGoCommentLineRegexContentIndex = /*                               ^^ 3. tag value */ 3
)

func ColumnTagGoCommentLineRegex() *regexp.Regexp {
	_ColumnTagGoCommentLineRegexOnce.Do(func() {
		_ColumnTagGoCommentLineRegex = regexp.MustCompile(fmt.Sprintf(_ColumnTagGoCommentLineRegexFormat, config.ColumnTagGo()))
	})
	return _ColumnTagGoCommentLineRegex
}
