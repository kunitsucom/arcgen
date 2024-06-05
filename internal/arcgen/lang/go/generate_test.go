//nolint:testpackage
package arcgengo

import (
	"bytes"
	"context"
	goast "go/ast"
	"io"
	"os"
	"testing"

	"github.com/kunitsucom/util.go/testing/assert"
	"github.com/kunitsucom/util.go/testing/require"

	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/contexts"
	"github.com/kunitsucom/arcgen/pkg/errors"
)

//nolint:paralleltest
func TestGenerate(t *testing.T) {
	t.Run("success,tests", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--method-name-table=GetTableName",
			"--method-name-columns=GetColumnNames",
			"--method-prefix-column=GetColumnName_",
			"--slice-type-suffix=Slice",
			// "tests/common.source",
			"tests",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".source"
		for _, src := range remainingArgs {
			require.NoError(t, Generate(ctx, src))
		}

		{
			expectedFile, err := os.Open("tests/common.golden")
			require.NoError(t, err)
			expectedBytes, err := io.ReadAll(expectedFile)
			require.NoError(t, err)
			expected := string(expectedBytes)

			actualFile, err := os.Open("tests/common.dbtest.gen.source")
			require.NoError(t, err)
			actualBytes, err := io.ReadAll(actualFile)
			require.NoError(t, err)
			actual := string(actualBytes)

			assert.Equal(t, expected, actual)
		}
	})

	t.Run("failure,no.errsource", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--method-name-table=GetTableName",
			"--method-name-columns=GetColumnNames",
			"--method-prefix-column=GetColumnName_",
			"tests/no.errsource",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".source"
		for _, src := range remainingArgs {
			require.ErrorContains(t, Generate(ctx, src), "expected 'package', found 'EOF'")
		}
	})

	t.Run("failure,no.errsource", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--method-name-table=GetTableName",
			"--method-name-columns=GetColumnNames",
			"--method-prefix-column=GetColumnName_",
			"tests",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".errsource"
		for _, src := range remainingArgs {
			require.ErrorContains(t, Generate(ctx, src), "expected 'package', found 'EOF'")
		}
	})

	t.Run("failure,no-such-file-or-directory", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--method-name-table=GetTableName",
			"--method-name-columns=GetColumnNames",
			"--method-prefix-column=GetColumnName_",
			"tests/no-such-file-or-directory",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".source"
		for _, src := range remainingArgs {
			require.ErrorContains(t, Generate(ctx, src), "no such file or directory")
		}
	})

	t.Run("failure,directory.dir", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--method-name-table=GetTableName",
			"--method-name-columns=GetColumnNames",
			"--method-prefix-column=GetColumnName_",
			"tests/directory.dir",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".dir"
		for _, src := range remainingArgs {
			require.ErrorContains(t, Generate(ctx, src), "is a directory")
		}
	})
}

var _ buffer = (*testBuffer)(nil)

type testBuffer struct {
	WriteFunc  func(p []byte) (n int, err error)
	StringFunc func() string
}

func (w *testBuffer) Write(p []byte) (n int, err error) {
	return w.WriteFunc(p)
}

func (w *testBuffer) String() string {
	return w.StringFunc()
}

//nolint:paralleltest
func Test_generate(t *testing.T) {
	t.Run("failure,os.OpenFile", func(t *testing.T) {
		arcSrcSets := ARCSourceSets{
			&ARCSourceSet{
				Filename: "tests/invalid-source-set",
				ARCSources: []*ARCSource{
					nil,
				},
			},
		}
		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })
		fileSuffix = ".invalid-source-set"
		err := generate(arcSrcSets)
		require.ErrorIs(t, err, errors.ErrInvalidSourceSet)
	})
}

func newTestARCSourceSet() *ARCSourceSet {
	return &ARCSourceSet{
		PackageName: "testpkg",
		ARCSources: []*ARCSource{
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
		err := fprint(nil, buf, arcSrcSet)
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
		err := fprint(f, bytes.NewBuffer(nil), arcSrcSet)
		require.ErrorIs(t, err, io.ErrClosedPipe)
	})
}
