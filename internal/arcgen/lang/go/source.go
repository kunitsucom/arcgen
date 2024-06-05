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
		// StructType is used to determine the column name. If the tag specified by --go-column-tag is not found, the field name is used.
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
