//nolint:testpackage
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
		ctx := contexts.WithArgs(context.Background(), []string{
			"ddlgen",
			"--column-tag-go=dbtest",
			"--method-prefix-global=Get",
			// "--src=tests/common.source",
			"--src=tests",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".source"
		require.NoError(t, Generate(ctx, config.Source()))

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
			"ddlgen",
			"--column-tag-go=dbtest",
			"--method-prefix-global=Get",
			"--src=tests/no.errsource",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".source"
		require.ErrorsContains(t, Generate(ctx, config.Source()), "expected 'package', found 'EOF'")
	})

	t.Run("failure,no.errsource", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"ddlgen",
			"--column-tag-go=dbtest",
			"--method-prefix-global=Get",
			"--src=tests",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".errsource"
		require.ErrorsContains(t, Generate(ctx, config.Source()), "expected 'package', found 'EOF'")
	})

	t.Run("failure,no-such-file-or-directory", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"ddlgen",
			"--column-tag-go=dbtest",
			"--method-prefix-global=Get",
			"--src=tests/no-such-file-or-directory",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".source"
		require.ErrorsContains(t, Generate(ctx, config.Source()), "no such file or directory")
	})

	t.Run("failure,directory.dir", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"ddlgen",
			"--column-tag-go=dbtest",
			"--method-prefix-global=Get",
			"--src=tests/directory.dir",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".dir"
		require.ErrorsContains(t, Generate(ctx, config.Source()), "is a directory")
	})
}
