//nolint:testpackage
package arcgengo

import (
	"bytes"
	"context"
	goast "go/ast"
	"io"
	"testing"

	"github.com/kunitsucom/util.go/testing/require"

	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/contexts"
	"github.com/kunitsucom/arcgen/pkg/errors"
)

var _ buffer = (*testBuffer)(nil)

type testBuffer struct {
	WriteFunc  func(p []byte) (n int, err error)
	StringFunc func() string
	NameFunc   func() string
}

func (w *testBuffer) Write(p []byte) (n int, err error) {
	return w.WriteFunc(p)
}

func (w *testBuffer) String() string {
	return w.StringFunc()
}

func (w *testBuffer) Name() string {
	return w.NameFunc()
}

//nolint:paralleltest
func Test_generate(t *testing.T) {
	t.Run("failure,os.OpenFile", func(t *testing.T) {
		ctx := contexts.WithOSArgs(context.Background(), []string{"--go-column-tag=dbtest"})
		require.NoError(t, func() error {
			_, _, err := config.Load(ctx)
			return err
		}())

		arcSrcSets := ARCSourceSetSlice{
			&ARCSourceSet{
				Filename: "tests/invalid-source-set",
				ARCSourceSlice: []*ARCSource{
					nil,
				},
			},
		}
		backup := fileExt
		t.Cleanup(func() { fileExt = backup })
		fileExt = ".invalid-source-set"
		err := generate(arcSrcSets)
		require.ErrorIs(t, err, errors.ErrInvalidSourceSet)
	})
}

func newTestARCSourceSet() *ARCSourceSet {
	return &ARCSourceSet{
		PackageName: "testpkg",
		ARCSourceSlice: []*ARCSource{
			{
				TypeSpec: &goast.TypeSpec{
					Name: &goast.Ident{
						Name: "Test",
					},
				},
				StructType: &goast.StructType{
					Fields: &goast.FieldList{
						List: []*goast.Field{
							{
								Names: []*goast.Ident{
									{
										Name: "ID",
									},
								},
								Tag: &goast.BasicLit{
									Value: "`dbtest:\"id\"`",
								},
							},
						},
					},
				},
			},
		},
	}
}

func Test_sprint(t *testing.T) {
	t.Parallel()
	t.Run("failure,buffer", func(t *testing.T) {
		t.Parallel()
		buf := &testBuffer{
			WriteFunc: func(_ []byte) (n int, err error) {
				return 0, io.ErrClosedPipe
			},
			StringFunc: func() string {
				return ""
			},
		}
		arcSrcSet := newTestARCSourceSet()
		err := fprintColumns(nil, buf, arcSrcSet)
		require.ErrorIs(t, err, io.ErrClosedPipe)
	})

	t.Run("failure,File", func(t *testing.T) {
		t.Parallel()
		f := &testBuffer{
			WriteFunc: func(_ []byte) (n int, err error) {
				return 0, io.ErrClosedPipe
			},
		}
		arcSrcSet := newTestARCSourceSet()
		err := fprintColumns(f, bytes.NewBuffer(nil), arcSrcSet)
		require.ErrorIs(t, err, io.ErrClosedPipe)
	})
}
