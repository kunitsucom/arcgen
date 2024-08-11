package arcgengo

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/kunitsucom/util.go/testing/assert"
	"github.com/kunitsucom/util.go/testing/require"

	"github.com/kunitsucom/arcgen/internal/config"
	"github.com/kunitsucom/arcgen/internal/contexts"
)

//nolint:paralleltest
func TestGenerate(t *testing.T) {
	t.Run("success,tests", func(t *testing.T) {
		ctx := contexts.WithOSArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--go-method-name-table=GetTableName",
			"--go-method-name-columns=GetColumnNames",
			"--go-method-prefix-column=GetColumnName_",
			"--go-slice-type-suffix=Slice",
			// "tests/common.source",
			"tests",
		})

		backup := fileExt
		t.Cleanup(func() { fileExt = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileExt = ".source"
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

	t.Run("success,skipErrInDir", func(t *testing.T) {
		ctx := contexts.WithOSArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--go-method-name-table=GetTableName",
			"--go-method-name-columns=GetColumnNames",
			"--go-method-prefix-column=GetColumnName_",
			"tests",
		})

		backup := fileExt
		t.Cleanup(func() { fileExt = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileExt = ".errsource"
		for _, src := range remainingArgs {
			require.NoError(t, Generate(ctx, src))
		}
	})

	t.Run("failure,no.errsource", func(t *testing.T) {
		ctx := contexts.WithOSArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--go-method-name-table=GetTableName",
			"--go-method-name-columns=GetColumnNames",
			"--go-method-prefix-column=GetColumnName_",
			"tests/no.errsource",
		})

		backup := fileExt
		t.Cleanup(func() { fileExt = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileExt = ".source"
		for _, src := range remainingArgs {
			require.ErrorContains(t, Generate(ctx, src), "expected 'package', found ")
		}
	})

	t.Run("failure,no-such-file-or-directory", func(t *testing.T) {
		ctx := contexts.WithOSArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--go-method-name-table=GetTableName",
			"--go-method-name-columns=GetColumnNames",
			"--go-method-prefix-column=GetColumnName_",
			"tests/no-such-file-or-directory",
		})

		backup := fileExt
		t.Cleanup(func() { fileExt = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileExt = ".source"
		for _, src := range remainingArgs {
			require.ErrorContains(t, Generate(ctx, src), "no such file or directory")
		}
	})

	t.Run("failure,directory.dir", func(t *testing.T) {
		ctx := contexts.WithOSArgs(context.Background(), []string{
			"arcgen",
			"--go-column-tag=dbtest",
			"--go-method-name-table=GetTableName",
			"--go-method-name-columns=GetColumnNames",
			"--go-method-prefix-column=GetColumnName_",
			"tests/directory.dir",
		})

		backup := fileExt
		t.Cleanup(func() { fileExt = backup })

		_, remainingArgs, err := config.Load(ctx)
		require.NoError(t, err)

		fileExt = ".dir"
		for _, src := range remainingArgs {
			require.ErrorContains(t, Generate(ctx, src), "is a directory")
		}
	})
}
