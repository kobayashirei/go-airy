package taskpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

// Task represents an asynchronous task that can be executed
type Task interface {
	Execute(ctx context.Context) error
}

// TaskFunc is a function type that implements Task interface
type TaskFunc func(ctx context.Context) error

// Execute implements Task interface for TaskFunc
func (f TaskFunc) Execute(ctx context.Context) error {
	return f(ctx)
}

// Pool wraps ants pool and provides task submission with error handling
type Pool struct {
	pool   *ants.Pool
	logger *zap.Logger
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// Config holds configuration for the task pool
type Config struct {
	// Size is the maximum number of goroutines in the pool
	Size int
	// ExpiryDuration is the interval time to clean up goroutines
	ExpiryDuration time.Duration
	// PreAlloc indicates whether to pre-allocate memory for goroutines
	PreAlloc bool
	// MaxBlockingTasks is the maximum number of tasks that can be blocked
	MaxBlockingTasks int
	// Nonblocking indicates whether the pool should be nonblocking
	Nonblocking bool
	// PanicHandler is called when a task panics
	PanicHandler func(interface{})
	// Logger for logging task execution
	Logger *zap.Logger
}

// DefaultConfig returns a default configuration for the task pool
func DefaultConfig() *Config {
	return &Config{
		Size:             10000,
		ExpiryDuration:   10 * time.Second,
		PreAlloc:         false,
		MaxBlockingTasks: 0,
		Nonblocking:      false,
		Logger:           zap.NewNop(),
	}
}

// NewPool creates a new task pool with the given configuration
func NewPool(config *Config) (*Pool, error) {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	options := []ants.Option{
		ants.WithExpiryDuration(config.ExpiryDuration),
		ants.WithPreAlloc(config.PreAlloc),
	}

	if config.MaxBlockingTasks > 0 {
		options = append(options, ants.WithMaxBlockingTasks(config.MaxBlockingTasks))
	}

	if config.Nonblocking {
		options = append(options, ants.WithNonblocking(true))
	}

	if config.PanicHandler != nil {
		options = append(options, ants.WithPanicHandler(config.PanicHandler))
	} else {
		// Default panic handler that logs the panic
		options = append(options, ants.WithPanicHandler(func(p interface{}) {
			if config.Logger != nil {
				config.Logger.Error("task panicked",
					zap.Any("panic", p),
					zap.Stack("stack"))
			}
		}))
	}

	pool, err := ants.NewPool(config.Size, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create ants pool: %w", err)
	}

	return &Pool{
		pool:   pool,
		logger: config.Logger,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Submit submits a task to the pool for execution
// Returns an error if the pool is closed or if submission fails
func (p *Pool) Submit(task Task) error {
	if p.pool.IsClosed() {
		return fmt.Errorf("task pool is closed")
	}

	p.wg.Add(1)
	
	err := p.pool.Submit(func() {
		defer p.wg.Done()
		
		// Create a context with timeout for the task
		ctx, cancel := context.WithTimeout(p.ctx, 5*time.Minute)
		defer cancel()

		// Execute the task and handle errors
		if err := task.Execute(ctx); err != nil {
			if p.logger != nil {
				p.logger.Error("task execution failed",
					zap.Error(err),
					zap.String("task_type", fmt.Sprintf("%T", task)))
			}
		}
	})

	if err != nil {
		p.wg.Done() // Decrement if submission failed
		return fmt.Errorf("failed to submit task: %w", err)
	}

	return nil
}

// SubmitFunc submits a function as a task to the pool
func (p *Pool) SubmitFunc(fn func(ctx context.Context) error) error {
	return p.Submit(TaskFunc(fn))
}

// Running returns the number of currently running goroutines
func (p *Pool) Running() int {
	return p.pool.Running()
}

// Free returns the number of available goroutines
func (p *Pool) Free() int {
	return p.pool.Free()
}

// Cap returns the capacity of the pool
func (p *Pool) Cap() int {
	return p.pool.Cap()
}

// Waiting returns the number of tasks waiting to be executed
func (p *Pool) Waiting() int {
	return p.pool.Waiting()
}

// IsClosed returns whether the pool is closed
func (p *Pool) IsClosed() bool {
	return p.pool.IsClosed()
}

// Release closes the pool and waits for all tasks to complete
func (p *Pool) Release() {
	p.cancel() // Cancel the context to signal all tasks to stop
	p.pool.Release()
}

// ReleaseTimeout closes the pool and waits for all tasks to complete with a timeout
func (p *Pool) ReleaseTimeout(timeout time.Duration) error {
	p.cancel() // Cancel the context to signal all tasks to stop
	
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.pool.Release()
		return nil
	case <-time.After(timeout):
		p.pool.Release()
		return fmt.Errorf("timeout waiting for tasks to complete")
	}
}

// Wait waits for all submitted tasks to complete
func (p *Pool) Wait() {
	p.wg.Wait()
}

// WaitWithTimeout waits for all submitted tasks to complete with a timeout
func (p *Pool) WaitWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for tasks to complete")
	}
}
