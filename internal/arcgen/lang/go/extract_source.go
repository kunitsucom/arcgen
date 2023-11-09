package arcgengo

import (
	"context"
	goast "go/ast"
	"go/token"
	"reflect"
	"strings"

	errorz "github.com/kunitsucom/util.go/errors"
	filepathz "github.com/kunitsucom/util.go/path/filepath"

	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/logs"
	apperr "github.com/kunitsucom/arcgen/pkg/errors"
)

//nolint:gocognit,cyclop
func extractSource(_ context.Context, fset *token.FileSet, f *goast.File) (ARCSourceSet, error) {
	arcSrcSet := make(ARCSourceSet, 0)
	for commentedNode, commentGroups := range goast.NewCommentMap(fset, f, f.Comments) {
		for _, commentGroup := range commentGroups {
		CommentGroupLoop:
			for _, commentLine := range commentGroup.List {
				logs.Trace.Printf("commentLine=%s: %s", filepathz.Short(fset.Position(commentGroup.Pos()).String()), commentLine.Text)
				// NOTE: If the comment line matches the ColumnTagGo, it is assumed to be a comment line for the struct.
				if matches := ColumnTagGoCommentLineRegex().FindStringSubmatch(commentLine.Text); len(matches) > _ColumnTagGoCommentLineRegexContentIndex {
					s := &ARCSource{
						Position:     fset.Position(commentLine.Pos()),
						Package:      f.Name,
						CommentGroup: commentGroup,
					}
					goast.Inspect(commentedNode, func(node goast.Node) bool {
						switch n := node.(type) {
						case *goast.TypeSpec:
							s.TypeSpec = n
							switch t := n.Type.(type) {
							case *goast.StructType:
								s.StructType = t
								if hasColumnTagGo(t) {
									logs.Debug.Printf("🔍: %s: type=%s", fset.Position(t.Pos()).String(), n.Name.Name)
									arcSrcSet = append(arcSrcSet, s)
								}
								return false
							default: // noop
							}
						default: // noop
						}
						return true
					})
					break CommentGroupLoop // NOTE: There may be multiple "ColumnTagGo"s in the same commentGroup, so once you find the first one, break.
				}
			}
		}
	}

	if len(arcSrcSet) == 0 {
		return nil, errorz.Errorf("column-tag-go=%s: %w", config.ColumnTagGo(), apperr.ErrColumnTagGoAnnotationNotFoundInSource)
	}

	return arcSrcSet, nil
}

func hasColumnTagGo(s *goast.StructType) bool {
	for _, field := range s.Fields.List {
		if field.Tag != nil {
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			if columnName := tag.Get(config.ColumnTagGo()); columnName != "" {
				return true
			}
		}
	}
	return false
}
