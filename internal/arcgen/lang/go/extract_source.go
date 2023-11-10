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

//nolint:cyclop,funlen,gocognit
func extractSource(_ context.Context, fset *token.FileSet, f *goast.File) (*ARCSourceSet, error) {
	// NOTE: Use map to avoid duplicate entries.
	arcSrcMap := make(map[string]*ARCSource)

	goast.Inspect(f, func(node goast.Node) bool {
		switch n := node.(type) {
		case *goast.TypeSpec:
			typeSpec := n
			switch t := n.Type.(type) {
			case *goast.StructType:
				structType := t
				if hasColumnTagGo(t) {
					pos := fset.Position(structType.Pos())
					logs.Debug.Printf("🔍: %s: type=%s", pos.String(), n.Name.Name)
					arcSrcMap[pos.String()] = &ARCSource{
						Source:     pos,
						Package:    f.Name,
						TypeSpec:   typeSpec,
						StructType: structType,
					}
				}
				return false
			default: // noop
			}
		default: // noop
		}
		return true
	})

	// Since it is not possible to extract the comment group associated with the position of struct,
	// search for the struct associated with the comment group and overwrite it.
	for commentedNode, commentGroups := range goast.NewCommentMap(fset, f, f.Comments) {
		for _, commentGroup := range commentGroups {
		CommentGroupLoop:
			for _, commentLine := range commentGroup.List {
				commentGroup := commentGroup // MEMO: Using the variable on range scope `commentGroup` in function literal (scopelint)
				logs.Trace.Printf("commentLine=%s: %s", filepathz.Short(fset.Position(commentGroup.Pos()).String()), commentLine.Text)
				// NOTE: If the comment line matches the ColumnTagGo, it is assumed to be a comment line for the struct.
				if matches := ColumnTagGoCommentLineRegex().FindStringSubmatch(commentLine.Text); len(matches) > _ColumnTagGoCommentLineRegexContentIndex {
					goast.Inspect(commentedNode, func(node goast.Node) bool {
						switch n := node.(type) {
						case *goast.TypeSpec:
							typeSpec := n
							switch t := n.Type.(type) {
							case *goast.StructType:
								structType := t
								if hasColumnTagGo(t) {
									pos := fset.Position(structType.Pos())
									logs.Debug.Printf("🖋️: %s: overwrite with comment group: type=%s", fset.Position(t.Pos()).String(), n.Name.Name)
									arcSrcMap[pos.String()] = &ARCSource{
										Source:       pos,
										Package:      f.Name,
										TypeSpec:     typeSpec,
										StructType:   structType,
										CommentGroup: commentGroup,
									}
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

	arcSrcSet := &ARCSourceSet{
		Filename:    fset.Position(f.Pos()).Filename,
		PackageName: f.Name.Name,
		Source:      fset.Position(f.Pos()),
		ARCSources:  make([]*ARCSource, 0),
	}

	for _, arcSrc := range arcSrcMap {
		arcSrcSet.ARCSources = append(arcSrcSet.ARCSources, arcSrc)
	}

	if len(arcSrcSet.ARCSources) == 0 {
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
