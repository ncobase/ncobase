package scheduler

import (
	"context"
	"fmt"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	nec "ncore/ext/core"
	"ncore/pkg/types"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

// Config represents scheduler configuration
type Config struct {
	Enabled       bool
	CheckInterval time.Duration
	BatchSize     int
	Workers       int
	QueueSize     int
	RetryInterval time.Duration
	MaxRetries    int
}

// DefaultConfig returns default scheduler configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		CheckInterval: time.Second * 10,
		BatchSize:     100,
		Workers:       10,
		QueueSize:     1000,
		RetryInterval: time.Second * 5,
		MaxRetries:    3,
	}
}

// Metrics represents scheduler metrics
type Metrics struct {
	TimeoutTasksScheduled  atomic.Int64
	TimeoutTasksCompleted  atomic.Int64
	TimeoutTasksFailed     atomic.Int64
	ReminderTasksScheduled atomic.Int64
	ReminderTasksCompleted atomic.Int64
	ReminderTasksFailed    atomic.Int64
	TimerTasksScheduled    atomic.Int64
	TimerTasksCompleted    atomic.Int64
	TimerTasksFailed       atomic.Int64
}

// Task represents a scheduled task
type Task struct {
	ID          string
	Type        string // timeout/reminder/timer
	ProcessID   string
	NodeID      string
	TriggerTime time.Time
	RetryCount  int
	Data        map[string]any
}

// Scheduler represents task scheduler
type Scheduler struct {
	// Dependencies
	services *service.Service
	em       nec.ManagerInterface
	config   *Config

	// Task queues
	timeoutTasks  chan *Task
	reminderTasks chan *Task
	timerTasks    chan *Task

	// Task map
	taskMap sync.Map

	// Runtime components
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	metrics *Metrics
}

// NewScheduler creates a new scheduler
//
// Usage:
//
//	// Create a new scheduler
//	svc := &service.Service{...}  // Provide necessary services
//	em := &nec.ManagerInterface{...} // Provide an extension manager
//	cfg := scheduler.DefaultConfig() // Use default configuration or create a custom one
//	s := scheduler.NewScheduler(svc, em, cfg)
//
//	// Start the scheduler
//	err := s.Start()
//	if err != nil {
//	    log.Fatalf("Failed to start scheduler: %v", err)
//	}
//
//	// Schedule a timeout task
//	err = s.ScheduleTimeout("process-1", "node-1", time.Now().Add(time.Hour))
//	if err != nil {
//	    log.Printf("Failed to schedule timeout task: %v", err)
//	}
//
//	// Schedule a reminder task
//	err = s.ScheduleReminder("process-2", "node-2", time.Now().Add(30*time.Minute))
//	if err != nil {
//	    log.Printf("Failed to schedule reminder task: %v", err)
//	}
//
//	// Schedule a timer task
//	err = s.ScheduleTimer("process-3", "node-3", time.Now().Add(2*time.Hour))
//	if err != nil {
//	    log.Printf("Failed to schedule timer task: %v", err)
//	}
//
//	// Cancel a timeout task
//	err = s.CancelTimeout("process-1", "node-1")
//	if err != nil {
//	    log.Printf("Failed to cancel timeout task: %v", err)
//	}
//
//	// Get scheduler metrics
//	metrics := s.GetMetrics()
//	log.Printf("Scheduler metrics: %+v", metrics)
//
//	// Stop the scheduler
//	s.Stop()
func NewScheduler(svc *service.Service, em nec.ManagerInterface, cfg *Config) *Scheduler {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &Scheduler{
		services:      svc,
		em:            em,
		config:        cfg,
		timeoutTasks:  make(chan *Task, cfg.QueueSize),
		reminderTasks: make(chan *Task, cfg.QueueSize),
		timerTasks:    make(chan *Task, cfg.QueueSize),
		metrics:       &Metrics{},
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	if !s.config.Enabled {
		return nil
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Start task processors
	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(3)
		go s.processTimeoutTasks()
		go s.processReminderTasks()
		go s.processTimerTasks()
	}

	// Start task checkers
	s.wg.Add(3)
	go s.checkTimeoutTasks()
	go s.checkReminderTasks()
	go s.checkTimerTasks()

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

// ScheduleTimeout schedules a timeout task
func (s *Scheduler) ScheduleTimeout(processID, nodeID string, deadline time.Time) error {
	task := &Task{
		ID:          fmt.Sprintf("timeout-%s-%s", processID, nodeID),
		Type:        "timeout",
		ProcessID:   processID,
		NodeID:      nodeID,
		TriggerTime: deadline,
	}

	select {
	case s.timeoutTasks <- task:
		s.taskMap.Store(task.ID, task)
		s.metrics.TimeoutTasksScheduled.Add(1)
		return nil
	default:
		return fmt.Errorf("timeout queue is full")
	}
}

// ScheduleReminder schedules a reminder task
func (s *Scheduler) ScheduleReminder(processID, nodeID string, remindAt time.Time) error {
	task := &Task{
		ID:          fmt.Sprintf("reminder-%s-%s", processID, nodeID),
		Type:        "reminder",
		ProcessID:   processID,
		NodeID:      nodeID,
		TriggerTime: remindAt,
	}

	select {
	case s.reminderTasks <- task:
		s.taskMap.Store(task.ID, task)
		s.metrics.ReminderTasksScheduled.Add(1)
		return nil
	default:
		return fmt.Errorf("reminder queue is full")
	}
}

// ScheduleTimer schedules a timer task
func (s *Scheduler) ScheduleTimer(processID, nodeID string, triggerAt time.Time) error {
	task := &Task{
		ID:          fmt.Sprintf("timer-%s-%s", processID, nodeID),
		Type:        "timer",
		ProcessID:   processID,
		NodeID:      nodeID,
		TriggerTime: triggerAt,
	}

	select {
	case s.timerTasks <- task:
		s.taskMap.Store(task.ID, task)
		s.metrics.TimerTasksScheduled.Add(1)
		return nil
	default:
		return fmt.Errorf("timer queue is full")
	}
}

// CancelTimeout cancels a timeout task
func (s *Scheduler) CancelTimeout(processID, nodeID string) error {
	taskID := fmt.Sprintf("timeout-%s-%s", processID, nodeID)
	return s.cancelTask(taskID)
}

// CancelReminder cancels a reminder task
func (s *Scheduler) CancelReminder(processID, nodeID string) error {
	taskID := fmt.Sprintf("reminder-%s-%s", processID, nodeID)
	return s.cancelTask(taskID)
}

// CancelTimer cancels a timer task
func (s *Scheduler) CancelTimer(processID, nodeID string) error {
	taskID := fmt.Sprintf("timer-%s-%s", processID, nodeID)
	return s.cancelTask(taskID)
}

// GetMetrics returns scheduler metrics
func (s *Scheduler) GetMetrics() map[string]int64 {
	return map[string]int64{
		"timeout_scheduled":  s.metrics.TimeoutTasksScheduled.Load(),
		"timeout_completed":  s.metrics.TimeoutTasksCompleted.Load(),
		"timeout_failed":     s.metrics.TimeoutTasksFailed.Load(),
		"reminder_scheduled": s.metrics.ReminderTasksScheduled.Load(),
		"reminder_completed": s.metrics.ReminderTasksCompleted.Load(),
		"reminder_failed":    s.metrics.ReminderTasksFailed.Load(),
		"timer_scheduled":    s.metrics.TimerTasksScheduled.Load(),
		"timer_completed":    s.metrics.TimerTasksCompleted.Load(),
		"timer_failed":       s.metrics.TimerTasksFailed.Load(),
	}
}

// Internal methods

// Task processors
func (s *Scheduler) processTimeoutTasks() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.timeoutTasks:
			if err := s.handleTimeout(task); err != nil {
				s.metrics.TimeoutTasksFailed.Add(1)
				s.retryTask(task)
			} else {
				s.metrics.TimeoutTasksCompleted.Add(1)
			}
			s.taskMap.Delete(task.ID)
		}
	}
}

// Reminder checkers
func (s *Scheduler) processReminderTasks() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.reminderTasks:
			if err := s.handleReminder(task); err != nil {
				s.metrics.ReminderTasksFailed.Add(1)
				s.retryTask(task)
			} else {
				s.metrics.ReminderTasksCompleted.Add(1)
			}
			s.taskMap.Delete(task.ID)
		}
	}
}

// Timer checkers
func (s *Scheduler) processTimerTasks() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.timerTasks:
			if err := s.handleTimer(task); err != nil {
				s.metrics.TimerTasksFailed.Add(1)
				s.retryTask(task)
			} else {
				s.metrics.TimerTasksCompleted.Add(1)
			}
			s.taskMap.Delete(task.ID)
		}
	}
}

// Task checkers
func (s *Scheduler) checkTimeoutTasks() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.findAndScheduleTimeouts()
		}
	}
}

// Reminder checkers
func (s *Scheduler) checkReminderTasks() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.findAndScheduleReminders()
		}
	}
}

// Timer checkers
func (s *Scheduler) checkTimerTasks() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.findAndScheduleTimers()
		}
	}
}

// Task handlers
func (s *Scheduler) handleTimeout(task *Task) error {
	node, err := s.services.Node.Get(s.ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeID,
	})
	if err != nil {
		return err
	}

	strategy := s.getTimeoutStrategy(node)
	switch strategy {
	case structs.TimeoutAutoPass:
		return s.handleTimeoutAutoPass(task)
	case structs.TimeoutAutoFail:
		return s.handleTimeoutAutoFail(task)
	case structs.TimeoutAlert:
		return s.handleTimeoutAlert(task)
	default:
		return fmt.Errorf("unknown timeout strategy: %s", strategy)
	}
}

func (s *Scheduler) handleReminder(task *Task) error {
	s.em.PublishEvent(structs.EventTaskUrged, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeID,
	})
	return nil
}

func (s *Scheduler) handleTimer(task *Task) error {
	// Handle different timer types based on task data
	return s.completeTimerNode(task)
}

// retryTask retries a task
func (s *Scheduler) retryTask(task *Task) {
	if task.RetryCount >= s.config.MaxRetries {
		return
	}

	task.RetryCount++
	time.Sleep(s.config.RetryInterval)

	switch task.Type {
	case "timeout":
		s.timeoutTasks <- task
	case "reminder":
		s.reminderTasks <- task
	case "timer":
		s.timerTasks <- task
	}

	s.taskMap.Store(task.ID, task)
}

func (s *Scheduler) cancelTask(taskID string) error {
	if task, ok := s.taskMap.Load(taskID); ok {
		s.taskMap.Delete(taskID)

		switch task.(*Task).Type {
		case "timeout":
			s.removeTaskFromQueue(s.timeoutTasks, taskID)
		case "reminder":
			s.removeTaskFromQueue(s.reminderTasks, taskID)
		case "timer":
			s.removeTaskFromQueue(s.timerTasks, taskID)
		}
	}

	return nil
}

// Remove task from queue
func (s *Scheduler) removeTaskFromQueue(queue chan *Task, taskID string) {
	var newQueue []*Task
	for task := range queue {
		if task.ID != taskID {
			newQueue = append(newQueue, task)
		}
	}
	close(queue)
	for _, task := range newQueue {
		queue <- task
	}
}

// Get timeout strategy
func (s *Scheduler) getTimeoutStrategy(node *structs.ReadNode) structs.TimeoutStrategy {
	if config, ok := node.Properties["timeoutConfig"].(map[string]any); ok {
		if strategy, ok := config["strategy"].(string); ok {
			return structs.TimeoutStrategy(strategy)
		}
	}
	return structs.TimeoutNone
}

// findAndScheduleTimeouts checks and schedules timeout tasks
func (s *Scheduler) findAndScheduleTimeouts() error {
	tasks, err := s.services.Task.List(s.ctx, &structs.ListTaskParams{
		Status: string(structs.StatusPending),
		Limit:  s.config.BatchSize,
	})
	if err != nil {
		return fmt.Errorf("list timeout tasks: %w", err)
	}

	for _, task := range tasks.Items {
		if task.DueTime != nil && time.Now().UnixMilli() > *task.DueTime {
			if err := s.ScheduleTimeout(task.ProcessID, task.NodeKey, types.ToValue(types.UnixMilliToTime(task.DueTime))); err != nil {
				// Log error but continue processing other tasks
				s.em.PublishEvent(structs.EventTaskError, &structs.EventData{
					ProcessID: task.ProcessID,
					TaskID:    task.ID,
					NodeID:    task.NodeKey,
					Details: types.JSON{
						"error": err.Error(),
					},
				})
			}
		}
	}
	return nil
}

// findAndScheduleReminders checks and schedules reminder tasks
func (s *Scheduler) findAndScheduleReminders() error {
	tasks, err := s.services.Task.List(s.ctx, &structs.ListTaskParams{
		Status: string(structs.StatusPending),
		Limit:  s.config.BatchSize,
	})
	if err != nil {
		return fmt.Errorf("list reminder tasks: %w", err)
	}

	for _, task := range tasks.Items {
		node, err := s.getNodeConfig(task)
		if err != nil {
			continue
		}

		if reminderConfig := s.getReminderConfig(node); reminderConfig != nil {
			if s.shouldSendReminder(task, reminderConfig) {
				remindAt := s.calculateNextReminder(task, reminderConfig)
				if err := s.ScheduleReminder(task.ProcessID, task.NodeKey, remindAt); err != nil {
					s.em.PublishEvent(structs.EventTaskError, &structs.EventData{
						ProcessID: task.ProcessID,
						TaskID:    task.ID,
						NodeID:    task.NodeKey,
						Details: types.JSON{
							"error": err.Error(),
						},
					})
				}
			}
		}
	}
	return nil
}

// findAndScheduleTimers checks and schedules timer tasks
func (s *Scheduler) findAndScheduleTimers() error {
	nodes, err := s.services.Node.List(s.ctx, &structs.ListNodeParams{
		Type:  "timer",
		Limit: s.config.BatchSize,
	})
	if err != nil {
		return fmt.Errorf("list timer nodes: %w", err)
	}
	for _, node := range nodes.Items {
		if timerConfig := s.getTimerConfig(node); timerConfig != nil {
			if triggerTime := s.getNextTriggerTime(node, timerConfig); !triggerTime.IsZero() {
				if err := s.ScheduleTimer(node.ProcessID, node.NodeKey, triggerTime); err != nil {
					s.em.PublishEvent(structs.EventNodeError, &structs.EventData{
						ProcessID: node.ProcessID,
						NodeID:    node.NodeKey,
						Details: types.JSON{
							"error": err.Error(),
						},
					})
				}
			}
		}
	}
	return nil
}

func (s *Scheduler) getNodeConfig(task *structs.ReadTask) (*structs.ReadNode, error) {
	return s.services.Node.Get(s.ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeKey,
	})
}

func (s *Scheduler) getReminderConfig(node *structs.ReadNode) map[string]any {
	if config, ok := node.Properties["reminderConfig"].(map[string]any); ok {
		return config
	}
	return nil
}

func (s *Scheduler) getTimerConfig(node *structs.ReadNode) map[string]any {
	if config, ok := node.Properties["timerConfig"].(map[string]any); ok {
		return config
	}
	return nil
}

func (s *Scheduler) shouldSendReminder(task *structs.ReadTask, config map[string]any) bool {
	maxReminders, _ := config["maxReminders"].(int)
	if maxReminders > 0 && task.UrgeCount >= maxReminders {
		return false
	}

	interval, _ := config["interval"].(int)
	if interval <= 0 {
		return false
	}

	lastUrgeTime := task.UpdatedAt
	if lastUrgeTime == nil {
		return true
	}

	return time.Since(time.Unix(*lastUrgeTime, 0)).Minutes() >= float64(interval)
}

func (s *Scheduler) calculateNextReminder(task *structs.ReadTask, config map[string]any) time.Time {
	interval, _ := config["interval"].(int)
	if interval <= 0 {
		interval = 30 // default 30 minutes
	}
	return time.Now().Add(time.Duration(interval) * time.Minute)
}

func (s *Scheduler) getNextTriggerTime(node *structs.ReadNode, config map[string]any) time.Time {
	timerType, _ := config["type"].(string)
	switch timerType {
	case "delay":
		return s.getDelayTriggerTime(node, config)
	case "cron":
		return s.getCronTriggerTime(node, config)
	case "date":
		return s.getDateTriggerTime(node, config)
	default:
		return time.Time{}
	}
}

// Specific timeout handlers

func (s *Scheduler) handleTimeoutAutoPass(task *Task) error {
	req := &structs.CompleteTaskRequest{
		TaskID:   task.ID,
		Action:   structs.ActionApprove,
		Operator: "system",
		Comment:  "Auto approved due to timeout",
	}

	if _, err := s.services.Task.Complete(s.ctx, req); err != nil {
		return fmt.Errorf("complete task: %w", err)
	}

	s.em.PublishEvent(structs.EventTaskCompleted, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeID,
		Action:    structs.ActionApprove,
	})

	return nil
}

func (s *Scheduler) handleTimeoutAutoFail(task *Task) error {
	req := &structs.CompleteTaskRequest{
		TaskID:   task.ID,
		Action:   structs.ActionReject,
		Operator: "system",
		Comment:  "Auto rejected due to timeout",
	}

	if _, err := s.services.Task.Complete(s.ctx, req); err != nil {
		return fmt.Errorf("complete task: %w", err)
	}

	s.em.PublishEvent(structs.EventTaskCompleted, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeID,
		Action:    structs.ActionReject,
	})

	return nil
}

func (s *Scheduler) handleTimeoutAlert(task *Task) error {
	// Update task status
	_, err := s.services.Task.Update(s.ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			IsTimeout: true,
		},
	})
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	// Publish timeout event
	s.em.PublishEvent(structs.EventTaskTimeout, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeID,
	})

	return nil
}

func (s *Scheduler) completeTimerNode(task *Task) error {
	node, err := s.services.Node.Get(s.ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeID,
	})
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	// Update node status
	_, err = s.services.Node.Update(s.ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusCompleted),
		},
	})
	if err != nil {
		return fmt.Errorf("update node: %w", err)
	}

	// Publish event
	s.em.PublishEvent(structs.EventNodeCompleted, &structs.EventData{
		ProcessID: task.ProcessID,
		NodeID:    task.NodeID,
	})

	return nil
}

// Timer related methods
func (s *Scheduler) getDelayTriggerTime(node *structs.ReadNode, config map[string]any) time.Time {
	duration, ok := config["duration"].(string)
	if !ok {
		return time.Time{}
	}

	// Parse duration string
	d, err := time.ParseDuration(duration)
	if err != nil {
		return time.Time{}
	}

	// For delay timer, trigger time is creation time + delay duration
	startTime := time.Unix(*node.CreatedAt, 0)
	return startTime.Add(d)
}
func (s *Scheduler) getCronTriggerTime(node *structs.ReadNode, config map[string]any) time.Time {
	expr, ok := config["cron"].(string)
	if !ok {
		return time.Time{}
	}

	// Parse cron expression
	schedule, err := cron.ParseStandard(expr)
	if err != nil {
		return time.Time{}
	}

	// Get last update time or create time
	var lastTime time.Time
	if node.UpdatedAt != nil {
		lastTime = time.Unix(*node.UpdatedAt, 0)
	} else {
		lastTime = time.Unix(*node.CreatedAt, 0)
	}

	// Next trigger time is the next schedule after last execution
	return schedule.Next(lastTime)
}

func (s *Scheduler) getDateTriggerTime(_ *structs.ReadNode, config map[string]any) time.Time {
	dateStr, ok := config["date"].(string)
	if !ok {
		return time.Time{}
	}

	// Parse date string in RFC3339 format
	triggerTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}
	}

	return triggerTime
}

// Utility method to check if a timer should trigger

func (s *Scheduler) shouldTriggerTimer(node *structs.ReadNode, config map[string]any) bool {
	timerType, _ := config["type"].(string)
	switch timerType {

	case "delay":
		triggerTime := s.getDelayTriggerTime(node, config)
		return !triggerTime.IsZero() && time.Now().After(triggerTime)
	case "cron":
		triggerTime := s.getCronTriggerTime(node, config)
		return !triggerTime.IsZero() && time.Now().After(triggerTime)
	case "date":
		triggerTime := s.getDateTriggerTime(node, config)
		return !triggerTime.IsZero() && time.Now().After(triggerTime)
	default:
		return false
	}
}

// Additional helper method for timer management

func (s *Scheduler) rescheduleTimer(node *structs.ReadNode, config map[string]any) error {
	timerType, _ := config["type"].(string)
	if timerType != "cron" {
		return nil // Only cron timers need rescheduling
	}

	// Calculate next trigger time for cron timer
	nextTrigger := s.getCronTriggerTime(node, config)
	if nextTrigger.IsZero() {
		return fmt.Errorf("failed to calculate next trigger time")
	}

	// Schedule next execution
	return s.ScheduleTimer(node.ProcessID, node.NodeKey, nextTrigger)
}
