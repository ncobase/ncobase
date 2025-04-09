package scheduler

//
// import (
// 	"context"
// 	"fmt"
// 	nec "ncore/ext/core"
// 	"ncobase/core/workflow/service"
// 	"ncobase/core/workflow/structs"
// 	"testing"
// 	"time"
//
// 	"github.com/stretchr/testify/assert"
// )
//
// func TestScheduler_ScheduleTimeout(t *testing.T) {
// 	s := newTestScheduler()
//
// 	err := s.ScheduleTimeout("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.ScheduleTimeout("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.ScheduleTimeout("process-2", "node-2", time.Now().Add(time.Minute))
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, int64(3), s.metrics.TimeoutTasksScheduled.Load())
// }
//
// func TestScheduler_ScheduleReminder(t *testing.T) {
// 	s := newTestScheduler()
//
// 	err := s.ScheduleReminder("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.ScheduleReminder("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.ScheduleReminder("process-2", "node-2", time.Now().Add(time.Minute))
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, int64(3), s.metrics.ReminderTasksScheduled.Load())
// }
//
// func TestScheduler_ScheduleTimer(t *testing.T) {
// 	s := newTestScheduler()
//
// 	err := s.ScheduleTimer("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.ScheduleTimer("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.ScheduleTimer("process-2", "node-2", time.Now().Add(time.Minute))
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, int64(3), s.metrics.TimerTasksScheduled.Load())
// }
//
// func TestScheduler_CancelTimeout(t *testing.T) {
// 	s := newTestScheduler()
//
// 	err := s.ScheduleTimeout("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.CancelTimeout("process-1", "node-1")
// 	assert.NoError(t, err)
//
// 	_, ok := s.taskMap.Load("timeout-process-1-node-1")
// 	assert.False(t, ok)
// }
//
// func TestScheduler_CancelReminder(t *testing.T) {
// 	s := newTestScheduler()
//
// 	err := s.ScheduleReminder("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.CancelReminder("process-1", "node-1")
// 	assert.NoError(t, err)
//
// 	_, ok := s.taskMap.Load("reminder-process-1-node-1")
// 	assert.False(t, ok)
// }
//
// func TestScheduler_CancelTimer(t *testing.T) {
// 	s := newTestScheduler()
//
// 	err := s.ScheduleTimer("process-1", "node-1", time.Now().Add(time.Hour))
// 	assert.NoError(t, err)
//
// 	err = s.CancelTimer("process-1", "node-1")
// 	assert.NoError(t, err)
//
// 	_, ok := s.taskMap.Load("timer-process-1-node-1")
// 	assert.False(t, ok)
// }
//
// func TestScheduler_HandleTimeout(t *testing.T) {
// 	s := newTestScheduler()
// 	task := testTask("timeout", "process-1", "node-1", time.Now())
//
// 	err := s.handleTimeout(task)
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, int64(1), s.metrics.TimeoutTasksCompleted.Load())
// }
//
// func TestScheduler_HandleReminder(t *testing.T) {
// 	s := newTestScheduler()
// 	task := testTask("reminder", "process-1", "node-1", time.Now())
//
// 	err := s.handleReminder(task)
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, int64(1), s.metrics.ReminderTasksCompleted.Load())
// }
//
// func TestScheduler_HandleTimer(t *testing.T) {
// 	s := newTestScheduler()
// 	task := testTask("timer", "process-1", "node-1", time.Now())
//
// 	err := s.handleTimer(task)
// 	assert.NoError(t, err)
//
// 	assert.Equal(t, int64(1), s.metrics.TimerTasksCompleted.Load())
// }
//
// func TestScheduler_RetryTask(t *testing.T) {
// 	s := newTestScheduler()
// 	task := testTask("timeout", "process-1", "node-1", time.Now())
//
// 	s.retryTask(task)
// 	assert.Equal(t, 1, task.RetryCount)
//
// 	s.retryTask(task)
// 	assert.Equal(t, 2, task.RetryCount)
//
// 	s.retryTask(task)
// 	assert.Equal(t, 3, task.RetryCount)
//
// 	s.retryTask(task)
// 	assert.Equal(t, 3, task.RetryCount)
// }
//
// func BenchmarkScheduler_ScheduleTimeout(b *testing.B) {
// 	s := newTestScheduler()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.ScheduleTimeout(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i), time.Now().Add(time.Hour))
// 	}
// }
//
// func BenchmarkScheduler_ScheduleReminder(b *testing.B) {
// 	s := newTestScheduler()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.ScheduleReminder(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i), time.Now().Add(time.Hour))
// 	}
// }
//
// func BenchmarkScheduler_ScheduleTimer(b *testing.B) {
// 	s := newTestScheduler()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.ScheduleTimer(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i), time.Now().Add(time.Hour))
// 	}
// }
//
// func BenchmarkScheduler_CancelTimeout(b *testing.B) {
// 	s := newTestScheduler()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.ScheduleTimeout(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i), time.Now().Add(time.Hour))
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.CancelTimeout(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i))
// 	}
// }
//
// func BenchmarkScheduler_CancelReminder(b *testing.B) {
// 	s := newTestScheduler()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.ScheduleReminder(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i), time.Now().Add(time.Hour))
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.CancelReminder(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i))
// 	}
// }
//
// func BenchmarkScheduler_CancelTimer(b *testing.B) {
// 	s := newTestScheduler()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.ScheduleTimer(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i), time.Now().Add(time.Hour))
// 	}
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = s.CancelTimer(fmt.Sprintf("process-%d", i), fmt.Sprintf("node-%d", i))
// 	}
// }
//
// // Mock services
// type mockNodeService struct{}
//
// func (m *mockNodeService) Get(_ context.Context, params *structs.FindNodeParams) (*structs.ReadNode, error) {
// 	return &structs.ReadNode{
// 		ID:         params.NodeKey,
// 		NodeKey:    params.NodeKey,
// 		ProcessID:  params.ProcessID,
// 		Properties: map[string]any{},
// 	}, nil
// }
//
// func (m *mockNodeService) List(_ context.Context, _ *structs.ListNodeParams) (*structs.ListNodeParams, error) {
// 	return &structs.ListNodeParams{}, nil
// }
//
// func (m *mockNodeService) Update(_ context.Context, _ *structs.UpdateNodeBody) (*structs.ReadNode, error) {
// 	return &structs.ReadNode{}, nil
// }
//
// type mockTaskService struct{}
//
// func (m *mockTaskService) List(_ context.Context, _ *structs.ListTaskParams) (*structs.ListTaskParams, error) {
// 	return &structs.ListTaskParams{}, nil
// }
//
// func (m *mockTaskService) Update(_ context.Context, _ *structs.UpdateTaskBody) (*structs.ReadTask, error) {
// 	return &structs.ReadTask{}, nil
// }
//
// func (m *mockTaskService) Complete(_ context.Context, req *structs.CompleteTaskRequest) (*structs.ReadTask, error) {
// 	return &structs.ReadTask{
// 		ID:     req.TaskID,
// 		Action: string(req.Action),
// 		Status: string(structs.StatusCompleted),
// 	}, nil
// }
//
// // Helper functions
// func newTestScheduler() *Scheduler {
// 	svc := &service.Service{
// 		Node: &mockNodeService{},
// 		Task: &mockTaskService{},
// 	}
// 	em := &nec.ManagerInterface{}
// 	cfg := DefaultConfig()
// 	return NewScheduler(svc, em, cfg)
// }
//
// func testTask(taskType, processID, nodeID string, triggerTime time.Time) *Task {
// 	return &Task{
// 		ID:          fmt.Sprintf("%s-%s-%s", taskType, processID, nodeID),
// 		Type:        taskType,
// 		ProcessID:   processID,
// 		NodeID:      nodeID,
// 		TriggerTime: triggerTime,
// 	}
// }
