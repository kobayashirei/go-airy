package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchLoader(t *testing.T) {
	t.Run("Load multiple items", func(t *testing.T) {
		loader := NewBatchLoader(func(ctx context.Context, keys []int64) (map[int64]string, error) {
			result := make(map[int64]string)
			for _, key := range keys {
				result[key] = "value"
			}
			return result, nil
		})

		ctx := context.Background()
		result, err := loader.Load(ctx, []int64{1, 2, 3})
		
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, "value", result[1])
		assert.Equal(t, "value", result[2])
		assert.Equal(t, "value", result[3])
	})

	t.Run("LoadOne existing item", func(t *testing.T) {
		loader := NewBatchLoader(func(ctx context.Context, keys []int64) (map[int64]string, error) {
			result := make(map[int64]string)
			for _, key := range keys {
				result[key] = "found"
			}
			return result, nil
		})

		ctx := context.Background()
		val, ok, err := loader.LoadOne(ctx, 1)
		
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, "found", val)
	})

	t.Run("LoadOne non-existing item", func(t *testing.T) {
		loader := NewBatchLoader(func(ctx context.Context, keys []int64) (map[int64]string, error) {
			return make(map[int64]string), nil
		})

		ctx := context.Background()
		val, ok, err := loader.LoadOne(ctx, 1)
		
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, "", val)
	})

	t.Run("Load empty keys", func(t *testing.T) {
		callCount := 0
		loader := NewBatchLoader(func(ctx context.Context, keys []int64) (map[int64]string, error) {
			callCount++
			return make(map[int64]string), nil
		})

		ctx := context.Background()
		result, err := loader.Load(ctx, []int64{})
		
		require.NoError(t, err)
		assert.Empty(t, result)
		assert.Equal(t, 1, callCount) // Loader is still called
	})
}
