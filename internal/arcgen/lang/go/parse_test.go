//nolint:testpackage
package arcgengo

import (
	"context"
	"os"
	"testing"

	"github.com/kunitsucom/util.go/testing/require"
)

func Test_walkDirFn(t *testing.T) {
	t.Parallel()

	t.Run("failure,os.Stat", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		arcSrcSets := make(ARCSourceSets, 0)
		require.ErrorIs(t, walkDirFn(ctx, &arcSrcSets)("tests", nil, os.ErrNotExist), os.ErrNotExist)
	})
}
