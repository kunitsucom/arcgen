//nolint:testpackage
package arcgengo

import (
	"context"
	"testing"

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
			"--src=tests",
		})

		backup := fileSuffix
		t.Cleanup(func() { fileSuffix = backup })

		_, err := config.Load(ctx)
		require.NoError(t, err)

		fileSuffix = ".source"
		require.NoError(t, Generate(ctx, config.Source()))
	})

	t.Run("failure,no.errsource", func(t *testing.T) {
		ctx := contexts.WithArgs(context.Background(), []string{
			"ddlgen",
			"--column-tag-go=dbtest",
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
