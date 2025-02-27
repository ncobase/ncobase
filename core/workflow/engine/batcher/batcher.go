package batcher

import (
	"context"
	"fmt"
	"ncobase/common/extension"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Config represents batcher configuration
type Config struct {
	Enabled       bool
	BatchSize     int           // maximum items in one batch
	FlushInterval time.Duration // interval to flush pending items
	QueueSize     int           // maximum pending items
	Workers       int           // number of worker goroutines
	RetryInterval time.Duration // retry interval
	MaxRetries    int           // maximum retry times
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		BatchSize:     100,
		FlushInterval: time.Second * 5,
		QueueSize:     1000,
		Workers:       5,
		RetryInterval: time.Second * 3,
		MaxRetries:    3,
	}
}

// BatchItem represents an item to be batched
type BatchItem struct {
	Priority   int
	CreateTime time.Time
	RetryCount int
	Task       *structs.TaskBody
	Done       chan error
}

// Batcher handles batch processing of tasks
type Batcher struct {
	services *service.Service
	em       *extension.Manager
	config   *Config

	items   chan *BatchItem
	batches chan []*BatchItem
	pending []*BatchItem
	metrics *BatcherMetrics

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// BatcherMetrics tracks operational metrics
type BatcherMetrics struct {
	ItemsReceived  atomic.Int64
	ItemsProcessed atomic.Int64
	BatchesCreated atomic.Int64
	BatchesSuccess atomic.Int64
	BatchesFailure atomic.Int64
	ProcessingTime atomic.Int64
}

// NewBatcher creates a new batcher
//
// # Usage
//
// // Create batcher
// cfg := DefaultConfig()
// batcher := NewBatcher(svc, em, cfg)
//
// // Start batcher
// batcher.Start()
//
// // Add task async
// task := &structs.TaskBody{
// TaskKey: "task-1",
// // ... other fields
// }
// err := batcher.Add(task, 1)
//
// // Add task sync
// ctx := context.Background()
// err = batcher.AddSync(ctx, task, 1)
//
// // Get metrics
// metrics := batcher.GetMetrics()
// fmt.Printf("Batcher metrics: %+v\n", metrics)
//
// // Stop batcher
// batcher.Stop()
func NewBatcher(svc *service.Service, em *extension.Manager, cfg *Config) *Batcher {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Batcher{
		services: svc,
		em:       em,
		config:   cfg,
		items:    make(chan *BatchItem, cfg.QueueSize),
		batches:  make(chan []*BatchItem, cfg.Workers),
		pending:  make([]*BatchItem, 0, cfg.BatchSize),
		metrics:  &BatcherMetrics{},
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the batcher
func (b *Batcher) Start() error {
	if !b.config.Enabled {
		return nil
	}

	// Start batch collector
	b.wg.Add(1)
	go b.collectBatches()

	// Start batch processors
	for i := 0; i < b.config.Workers; i++ {
		b.wg.Add(1)
		go b.processBatches()
	}

	return nil
}

// Stop stops the batcher
func (b *Batcher) Stop() {
	b.cancel()
	close(b.items)
	b.wg.Wait()
}

// Add adds a task to be batched
func (b *Batcher) Add(task *structs.TaskBody, priority int) error {
	item := &BatchItem{
		Priority:   priority,
		CreateTime: time.Now(),
		Task:       task,
		Done:       make(chan error, 1),
	}

	select {
	case b.items <- item:
		b.metrics.ItemsReceived.Add(1)
		return nil
	default:
		return fmt.Errorf("batcher queue is full")
	}
}

// AddSync adds a task and waits for completion
func (b *Batcher) AddSync(ctx context.Context, task *structs.TaskBody, priority int) error {
	item := &BatchItem{
		Priority:   priority,
		CreateTime: time.Now(),
		Task:       task,
		Done:       make(chan error, 1),
	}

	select {
	case b.items <- item:
		b.metrics.ItemsReceived.Add(1)
	default:
		return fmt.Errorf("batcher queue is full")
	}

	select {
	case err := <-item.Done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Size returns current pending items count
func (b *Batcher) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.pending) + len(b.items)
}

// collectBatches collects items into batches
func (b *Batcher) collectBatches() {
	defer b.wg.Done()

	timer := time.NewTicker(b.config.FlushInterval)
	defer timer.Stop()

	for {
		select {
		case <-b.ctx.Done():
			b.flush() // Final flush
			return

		case item := <-b.items:
			b.addToPending(item)
			if len(b.pending) >= b.config.BatchSize {
				b.flush()
			}

		case <-timer.C:
			if len(b.pending) > 0 {
				b.flush()
			}
		}
	}
}

// addToPending adds an item to pending list
func (b *Batcher) addToPending(item *BatchItem) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pending = append(b.pending, item)
}

// flush creates a batch from pending items
func (b *Batcher) flush() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.pending) == 0 {
		return
	}

	// Sort by priority and creation time
	sort.Slice(b.pending, func(i, j int) bool {
		if b.pending[i].Priority != b.pending[j].Priority {
			return b.pending[i].Priority > b.pending[j].Priority
		}
		return b.pending[i].CreateTime.Before(b.pending[j].CreateTime)
	})

	batch := make([]*BatchItem, len(b.pending))
	copy(batch, b.pending)
	b.pending = b.pending[:0]

	b.metrics.BatchesCreated.Add(1)
	b.batches <- batch
}

// processBatches processes batches of items
func (b *Batcher) processBatches() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return

		case batch := <-b.batches:
			start := time.Now()
			if err := b.processBatch(batch); err != nil {
				b.metrics.BatchesFailure.Add(1)
				b.handleBatchError(batch, err)
			} else {
				b.metrics.BatchesSuccess.Add(1)
				b.completeBatch(batch, nil)
			}
			b.metrics.ProcessingTime.Add(time.Since(start).Nanoseconds())
		}
	}
}

// processBatch processes a batch of items
func (b *Batcher) processBatch(batch []*BatchItem) error {
	// Prepare batch of tasks
	var tasks []structs.TaskBody
	for _, item := range batch {
		tasks = append(tasks, *item.Task)
	}

	// Batch create tasks
	ctx := context.Background()
	if err := b.createTasks(ctx, tasks); err != nil {
		return fmt.Errorf("create tasks: %w", err)
	}

	b.metrics.ItemsProcessed.Add(int64(len(batch)))
	return nil
}

// createTasks creates multiple tasks
func (b *Batcher) createTasks(ctx context.Context, tasks []structs.TaskBody) error {
	for _, task := range tasks {
		if _, err := b.services.Task.Create(ctx, &task); err != nil {
			return err
		}

		// Publish event
		b.em.PublishEvent(structs.EventTaskCreated, &structs.EventData{
			ProcessID: task.ProcessID,
			TaskID:    task.TaskKey,
			NodeID:    task.NodeKey,
			NodeType:  structs.NodeType(task.NodeType),
		})
	}
	return nil
}

// handleBatchError handles batch processing error
func (b *Batcher) handleBatchError(batch []*BatchItem, err error) {
	for _, item := range batch {
		item.RetryCount++
		if item.RetryCount < b.config.MaxRetries {
			time.Sleep(b.config.RetryInterval)
			// Retry by adding back to queue
			select {
			case b.items <- item:
			default:
				b.completeBatch([]*BatchItem{item}, fmt.Errorf("retry queue full: %w", err))
			}
		} else {
			b.completeBatch([]*BatchItem{item}, fmt.Errorf("max retries exceeded: %w", err))
		}
	}
}

// completeBatch completes a batch with result
func (b *Batcher) completeBatch(batch []*BatchItem, err error) {
	for _, item := range batch {
		select {
		case item.Done <- err:
		default:
		}
		close(item.Done)
	}
}

// GetMetrics returns batcher metrics
func (b *Batcher) GetMetrics() map[string]int64 {
	return map[string]int64{
		"items_received":  b.metrics.ItemsReceived.Load(),
		"items_processed": b.metrics.ItemsProcessed.Load(),
		"batches_created": b.metrics.BatchesCreated.Load(),
		"batches_success": b.metrics.BatchesSuccess.Load(),
		"batches_failure": b.metrics.BatchesFailure.Load(),
		"processing_time": b.metrics.ProcessingTime.Load(),
	}
}
