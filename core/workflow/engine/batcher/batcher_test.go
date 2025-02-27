package batcher

//
// import (
// 	"context"
// 	"fmt"
// 	"ncobase/core/workflow/service"
// 	"ncobase/core/workflow/structs"
// 	"testing"
// 	"time"
// )
//
// // Mock service for testing
// type mockService struct{}
//
// func (m *mockService) Task() *service.TaskService {
// 	return &service.TaskService{}
// }
//
// func TestBatcher_Add(t *testing.T) {
// 	cfg := DefaultConfig()
// 	b := NewBatcher(&mockService{}, nil, cfg)
//
// 	// Start the batcher
// 	err := b.Start()
// 	if err != nil {
// 		t.Fatalf("Failed to start batcher: %v", err)
// 	}
// 	defer b.Stop()
//
// 	// Add a task
// 	task := &structs.TaskBody{TaskKey: "task-1"}
// 	err = b.Add(task, 1)
// 	if err != nil {
// 		t.Fatalf("Failed to add task: %v", err)
// 	}
// 	if b.Size() != 1 {
// 		t.Fatalf("Expected 1 task in queue, got %d", b.Size())
// 	}
// }
//
// func TestBatcher_AddSync(t *testing.T) {
// 	cfg := DefaultConfig()
// 	b := NewBatcher(&mockService{}, nil, cfg)
//
// 	// Start the batcher
// 	err := b.Start()
// 	if err != nil {
// 		t.Fatalf("Failed to start batcher: %v", err)
// 	}
// 	defer b.Stop()
//
// 	// Add a task synchronously
// 	task := &structs.TaskBody{TaskKey: "task-sync"}
// 	ctx := context.Background()
// 	err = b.AddSync(ctx, task, 1)
// 	if err != nil {
// 		t.Fatalf("Failed to add task synchronously: %v", err)
// 	}
// 	if b.Size() != 1 {
// 		t.Fatalf("Expected 1 task in queue, got %d", b.Size())
// 	}
// }
//
// func TestBatcher_MaxRetries(t *testing.T) {
// 	cfg := DefaultConfig()
// 	cfg.MaxRetries = 3
// 	cfg.RetryInterval = time.Millisecond * 50
// 	b := NewBatcher(&mockService{}, nil, cfg)
//
// 	// Start the batcher
// 	err := b.Start()
// 	if err != nil {
// 		t.Fatalf("Failed to start batcher: %v", err)
// 	}
// 	defer b.Stop()
//
// 	// Add a task that will fail
// 	task := &structs.TaskBody{TaskKey: "task-fail"}
// 	err = b.Add(task, 1)
// 	if err != nil {
// 		t.Fatalf("Failed to add task: %v", err)
// 	}
//
// 	// Wait for retries to finish
// 	time.Sleep(time.Millisecond * 200)
//
// 	metrics := b.GetMetrics()
// 	if metrics["batches_failure"] == 0 {
// 		t.Fatalf("Expected at least one batch failure")
// 	}
// 	if metrics["items_processed"] != 1 {
// 		t.Fatalf("Expected 1 item processed, got %d", metrics["items_processed"])
// 	}
// }
//
// func TestBatcher_QueueFull(t *testing.T) {
// 	cfg := DefaultConfig()
// 	cfg.QueueSize = 2
// 	b := NewBatcher(&mockService{}, nil, cfg)
//
// 	// Start the batcher
// 	err := b.Start()
// 	if err != nil {
// 		t.Fatalf("Failed to start batcher: %v", err)
// 	}
// 	defer b.Stop()
//
// 	// Fill the queue
// 	task := &structs.TaskBody{TaskKey: "task-queue-full"}
// 	err = b.Add(task, 1)
// 	if err != nil {
// 		t.Fatalf("Failed to add task: %v", err)
// 	}
// 	err = b.Add(task, 1)
// 	if err != nil {
// 		t.Fatalf("Failed to add task: %v", err)
// 	}
//
// 	// Try adding another task, expecting queue to be full
// 	err = b.Add(task, 1)
// 	if err == nil {
// 		t.Fatalf("Expected error for full queue, got nil")
// 	}
// }
//
// func TestBatcher_Metrics(t *testing.T) {
// 	cfg := DefaultConfig()
// 	b := NewBatcher(&mockService{}, nil, cfg)
//
// 	// Start the batcher
// 	err := b.Start()
// 	if err != nil {
// 		t.Fatalf("Failed to start batcher: %v", err)
// 	}
// 	defer b.Stop()
//
// 	// Add tasks
// 	for i := 0; i < 10; i++ {
// 		task := &structs.TaskBody{TaskKey: fmt.Sprintf("task-%d", i)}
// 		err := b.Add(task, 1)
// 		if err != nil {
// 			t.Fatalf("Failed to add task %d: %v", i, err)
// 		}
// 	}
//
// 	// Wait for tasks to process
// 	time.Sleep(time.Second * 2)
//
// 	// Check metrics
// 	metrics := b.GetMetrics()
// 	if metrics["items_received"] != 10 {
// 		t.Fatalf("Expected 10 items received, got %d", metrics["items_received"])
// 	}
// }
//
// func BenchmarkBatcher(b *testing.B) {
// 	cfg := DefaultConfig()
// 	cfg.BatchSize = 50
// 	cfg.QueueSize = 100
// 	cfg.Workers = 10
//
// 	bc := NewBatcher(&mockService{}, nil, cfg)
// 	_ = bc.Start()
// 	defer bc.Stop()
//
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		task := &structs.TaskBody{TaskKey: fmt.Sprintf("task-%d", i)}
// 		_ = bc.Add(task, 1)
// 	}
// }
//
// func BenchmarkBatcher_LargeQueue(b *testing.B) {
// 	cfg := DefaultConfig()
// 	cfg.BatchSize = 100
// 	cfg.QueueSize = 1000
// 	cfg.Workers = 20
//
// 	bc := NewBatcher(&mockService{}, nil, cfg)
// 	_ = bc.Start()
// 	defer bc.Stop()
//
// 	b.ResetTimer()
//
// 	for i := 0; i < b.N; i++ {
// 		task := &structs.TaskBody{TaskKey: fmt.Sprintf("task-%d", i)}
// 		_ = bc.Add(task, 1)
// 	}
// }
