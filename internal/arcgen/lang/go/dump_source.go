package arcgengo

import (
	"bytes"
	goast "go/ast"
	"go/token"
	"io"

	"github.com/kunitsucom/arcgen/internal/logs"
)

func dumpSource(fset *token.FileSet, arcSrcSet *FileSource) {
	if arcSrcSet != nil {
		for _, arcSrc := range arcSrcSet.StructSourceSlice {
			logs.Trace.Print("== Source ================================================================================================================================")
			_, _ = io.WriteString(logs.Trace.LineWriter("r.CommentGroup.Text: "), arcSrc.CommentGroup.Text())
			logs.Trace.Print("-- CommentGroup --------------------------------------------------------------------------------------------------------------------------------")
			{
				commentGroupAST := bytes.NewBuffer(nil)
				goast.Fprint(commentGroupAST, fset, arcSrc.CommentGroup, goast.NotNilFilter)
				_, _ = logs.Trace.LineWriter("").Write(commentGroupAST.Bytes())
			}
			logs.Trace.Print("-- TypeSpec --------------------------------------------------------------------------------------------------------------------------------")
			{
				typeSpecAST := bytes.NewBuffer(nil)
				goast.Fprint(typeSpecAST, fset, arcSrc.TypeSpec, goast.NotNilFilter)
				_, _ = logs.Trace.LineWriter("").Write(typeSpecAST.Bytes())
			}
			logs.Trace.Print("-- StructType --------------------------------------------------------------------------------------------------------------------------------")
			{
				structTypeAST := bytes.NewBuffer(nil)
				goast.Fprint(structTypeAST, fset, arcSrc.StructType, goast.NotNilFilter)
				_, _ = logs.Trace.LineWriter("").Write(structTypeAST.Bytes())
			}
		}
	}
}
