package taskpool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPool(t *testing.T) {
	t.Run("creates pool with default config", func(t *testing.T) {
		pool, err := NewPool(nil)
		require.NoError(t, err)
		require.NotNil(t, pool)
		defer pool.Release()

		assert.Equal(t, 10000, pool.Cap())
		assert.False(t, pool.IsClosed())
	})

	t.Run("creates pool with custom config", func(t *testing.T) {
		config := &Config{
			Size:           100,
			ExpiryDuration: 5 * time.Second,
			PreAlloc:       true,
			Logger:         zap.NewNop(),
		}

		pool, err := NewPool(config)
		require.NoError(t, err)
		require.NotNil(t, pool)
		defer pool.Release()

		assert.Equal(t, 100, pool.Cap())
	})
}

func TestPool_Submit(t *testing.T) {
	t.Run("submits and executes task successfully", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)
		defer pool.Release()

		var executed atomic.Bool
		task := TaskFunc(func(ctx context.Context) error {
			executed.Store(true)
			return nil
		})

		err = pool.Submit(task)
		require.NoError(t, err)

		pool.Wait()
		assert.True(t, executed.Load())
	})

	t.Run("handles task error", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		config := DefaultConfig()
		config.Logger = logger

		pool, err := NewPool(config)
		require.NoError(t, err)
		defer pool.Release()

		expectedErr := errors.New("task failed")
		task := TaskFunc(func(ctx context.Context) error {
			return expectedErr
		})

		err = pool.Submit(task)
		require.NoError(t, err)

		pool.Wait()
		// Task error should be logged but not returned
	})

	t.Run("returns error when pool is closed", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)
		pool.Release()

		task := TaskFunc(func(ctx context.Context) error {
			return nil
		})

		err = pool.Submit(task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed")
	})
}

func TestPool_SubmitFunc(t *testing.T) {
	t.Run("submits function as task", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)
		defer pool.Release()

		var counter atomic.Int32
		err = pool.SubmitFunc(func(ctx context.Context) error {
			counter.Add(1)
			return nil
		})
		require.NoError(t, err)

		pool.Wait()
		assert.Equal(t, int32(1), counter.Load())
	})
}

func TestPool_MultipleTasks(t *testing.T) {
	t.Run("executes multiple tasks concurrently", func(t *testing.T) {
		config := &Config{
			Size:   10,
			Logger: zap.NewNop(),
		}

		pool, err := NewPool(config)
		require.NoError(t, err)
		defer pool.Release()

		var counter atomic.Int32
		numTasks := 100

		for i := 0; i < numTasks; i++ {
			err := pool.SubmitFunc(func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				counter.Add(1)
				return nil
			})
			require.NoError(t, err)
		}

		pool.Wait()
		assert.Equal(t, int32(numTasks), counter.Load())
	})
}

func TestPool_PanicHandler(t *testing.T) {
	t.Run("handles panic in task", func(t *testing.T) {
		var panicValue interface{}
		config := &Config{
			Size: 10,
			PanicHandler: func(p interface{}) {
				panicValue = p
			},
			Logger: zap.NewNop(),
		}

		pool, err := NewPool(config)
		require.NoError(t, err)
		defer pool.Release()

		task := TaskFunc(func(ctx context.Context) error {
			panic("test panic")
		})

		err = pool.Submit(task)
		require.NoError(t, err)

		pool.Wait()
		assert.Equal(t, "test panic", panicValue)
	})
}

func TestPool_Metrics(t *testing.T) {
	t.Run("reports pool metrics", func(t *testing.T) {
		config := &Config{
			Size:   5,
			Logger: zap.NewNop(),
		}

		pool, err := NewPool(config)
		require.NoError(t, err)
		defer pool.Release()

		assert.Equal(t, 5, pool.Cap())
		assert.Equal(t, 5, pool.Free())
		assert.Equal(t, 0, pool.Running())
		assert.Equal(t, 0, pool.Waiting())

		// Submit a long-running task
		done := make(chan struct{})
		err = pool.SubmitFunc(func(ctx context.Context) error {
			<-done
			return nil
		})
		require.NoError(t, err)

		// Give it time to start
		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 1, pool.Running())
		assert.Equal(t, 4, pool.Free())

		close(done)
		pool.Wait()
	})
}

func TestPool_ReleaseTimeout(t *testing.T) {
	t.Run("releases with timeout", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)

		// Submit a quick task
		err = pool.SubmitFunc(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		require.NoError(t, err)

		err = pool.ReleaseTimeout(1 * time.Second)
		assert.NoError(t, err)
	})

	t.Run("times out when tasks take too long", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)

		// Submit a long-running task
		err = pool.SubmitFunc(func(ctx context.Context) error {
			time.Sleep(2 * time.Second)
			return nil
		})
		require.NoError(t, err)

		err = pool.ReleaseTimeout(100 * time.Millisecond)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})
}

func TestPool_WaitWithTimeout(t *testing.T) {
	t.Run("waits successfully within timeout", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)
		defer pool.Release()

		err = pool.SubmitFunc(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		require.NoError(t, err)

		err = pool.WaitWithTimeout(1 * time.Second)
		assert.NoError(t, err)
	})

	t.Run("times out when waiting too long", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)
		defer pool.Release()

		err = pool.SubmitFunc(func(ctx context.Context) error {
			time.Sleep(2 * time.Second)
			return nil
		})
		require.NoError(t, err)

		err = pool.WaitWithTimeout(100 * time.Millisecond)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})
}

func TestPool_ContextCancellation(t *testing.T) {
	t.Run("task receives cancelled context on pool release", func(t *testing.T) {
		pool, err := NewPool(DefaultConfig())
		require.NoError(t, err)

		var ctxCancelled atomic.Bool
		err = pool.SubmitFunc(func(ctx context.Context) error {
			<-ctx.Done()
			ctxCancelled.Store(true)
			return ctx.Err()
		})
		require.NoError(t, err)

		// Give task time to start
		time.Sleep(50 * time.Millisecond)

		pool.Release()
		time.Sleep(50 * time.Millisecond)

		assert.True(t, ctxCancelled.Load())
	})
}
