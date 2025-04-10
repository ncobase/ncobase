package handler

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/scheduler"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"time"

	ext "github.com/ncobase/ncore/ext/types"

	"github.com/robfig/cron/v3"
)

// TimerHandler handles timer nodes
type TimerHandler struct {
	*BaseHandler
	// Configuration
	config *config.TimerHandlerConfig

	// Scheduler
	scheduler *scheduler.Scheduler
}

// NewTimerHandler creates a new timer handler
func NewTimerHandler(svc *service.Service, em ext.ManagerInterface, cfg *config.Config) *TimerHandler {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &TimerHandler{
		BaseHandler: NewBaseHandler("timer", "Timer Handler", svc, em, cfg.Handlers.Base),
		config:      cfg.Handlers.Timer,
		scheduler:   scheduler.NewScheduler(svc, em, cfg.Components.Scheduler),
	}
}

// Type returns handler type
func (h *TimerHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *TimerHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *TimerHandler) Priority() int { return h.priority }

// Start starts the timer handler
func (h *TimerHandler) Start() error {
	if err := h.BaseHandler.Start(); err != nil {
		return err
	}

	// Start scheduler
	if err := h.scheduler.Start(); err != nil {
		return err
	}

	return nil
}

// Stop stops the timer handler
func (h *TimerHandler) Stop() error {
	if err := h.BaseHandler.Stop(); err != nil {
		return err
	}

	// Stop scheduler
	if h.scheduler != nil {
		h.scheduler.Stop()
	}

	return nil
}

// executeInternal executes the timer node
func (h *TimerHandler) executeInternal(ctx context.Context, node *structs.ReadNode) error {
	// Parse timer config
	c, err := h.parseTimerConfig(node)
	if err != nil {
		return err
	}

	// Calculate trigger time
	triggerTime, err := h.calculateTriggerTime(c)
	if err != nil {
		return err
	}

	// For immediate execution
	if !triggerTime.After(time.Now()) {
		return h.completeTimer(ctx, node)
	}

	// Schedule timer task
	return h.scheduler.ScheduleTimer(node.ProcessID, node.NodeKey, triggerTime)
}

// completeInternal completes the timer node
func (h *TimerHandler) completeInternal(ctx context.Context, node *structs.ReadNode, _ *structs.CompleteTaskRequest) error {
	c, err := h.parseTimerConfig(node)
	if err != nil {
		return err
	}

	// For cyclic timers, schedule next cycle if needed
	if c.Type == "cycle" && h.shouldScheduleNextCycle(node) {
		nextTime, err := h.calculateNextCycle(node, c)
		if err != nil {
			return err
		}
		return h.scheduler.ScheduleTimer(node.ProcessID, node.NodeKey, nextTime)
	}

	return h.completeTimer(ctx, node)
}

// validateInternal validates the timer node
func (h *TimerHandler) validateInternal(node *structs.ReadNode) error {
	c, err := h.parseTimerConfig(node)
	if err != nil {
		return err
	}

	// Validate based on timer type
	switch c.Type {
	case string(types.TimerDelay):
		if _, err := time.ParseDuration(c.Duration); err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
	case string(types.TimerCron):
		if _, err := cron.ParseStandard(c.CronExpr); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
	case string(types.TimerCycle):
		if c.CycleCount <= 0 || c.CycleInterval <= 0 {
			return fmt.Errorf("invalid cycle configuration")
		}
	case string(types.TimerDate):
		if _, err := time.Parse(time.RFC3339, c.TriggerDate); err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
	default:
		return fmt.Errorf("unsupported timer type: %s", c.Type)
	}

	return nil
}

// rollbackInternal rollbacks the timer node
func (h *TimerHandler) rollbackInternal(_ context.Context, node *structs.ReadNode) error {
	// Cancel scheduled timer
	return h.scheduler.CancelTimer(node.ProcessID, node.ID)
}

// parseTimerConfig parses the timer configuration
func (h *TimerHandler) parseTimerConfig(node *structs.ReadNode) (*config.TimerHandlerConfig, error) {
	c, ok := node.Properties["timerConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing timer configuration", nil)
	}

	result := config.DefaultTimerHandlerConfig()

	// Parse type specific fields
	switch result.Type {
	case string(types.TimerDelay):
		if duration, ok := c["duration"].(string); ok {
			result.Duration = duration
		}
	case string(types.TimerCron):
		if expr, ok := c["cron"].(string); ok {
			result.CronExpr = expr
		}
	case string(types.TimerCycle):
		if count, ok := c["cycle_count"].(int); ok {
			result.CycleCount = count
		}
		if interval, ok := c["cycle_interval"].(int); ok {
			result.CycleInterval = time.Duration(interval)
		}
	case string(types.TimerDate):
		if date, ok := c["date"].(string); ok {
			result.TriggerDate = date
		}
	}

	// Parse common fields
	if timeout, ok := c["timeout_hours"].(int); ok {
		result.TimeoutHours = timeout
	}
	if isWorkingDay, ok := c["is_working_day"].(bool); ok {
		result.IsWorkingDay = isWorkingDay
	}
	if failureMode, ok := c["failure_mode"].(string); ok {
		result.FailureMode = failureMode
	}

	return result, nil
}

// calculateTriggerTime calculates the trigger time based on the timer type
func (h *TimerHandler) calculateTriggerTime(config *config.TimerHandlerConfig) (time.Time, error) {
	now := time.Now()

	switch config.Type {
	case string(types.TimerDelay):
		duration, err := time.ParseDuration(config.Duration)
		if err != nil {
			return time.Time{}, err
		}
		return now.Add(duration), nil
	case string(types.TimerCron):
		schedule, err := cron.ParseStandard(config.CronExpr)
		if err != nil {
			return time.Time{}, err
		}
		return schedule.Next(now), nil
	case string(types.TimerCycle):
		return now.Add(time.Duration(config.CycleInterval) * time.Second), nil
	case string(types.TimerDate):
		return time.Parse(time.RFC3339, config.TriggerDate)
	default:
		return time.Time{}, fmt.Errorf("unsupported timer type")
	}
}

// shouldScheduleNextCycle checks if the timer should schedule the next cycle
func (h *TimerHandler) shouldScheduleNextCycle(node *structs.ReadNode) bool {
	cycleCount, ok := node.Properties["cycleCount"].(int)
	if !ok {
		return false
	}

	maxCycles, ok := node.Properties["maxCycles"].(int)
	if !ok {
		return true // No max cycles defined
	}

	return cycleCount < maxCycles
}

// calculateNextCycle calculates the next cycle time
func (h *TimerHandler) calculateNextCycle(node *structs.ReadNode, config *config.TimerHandlerConfig) (time.Time, error) {
	lastTrigger, ok := node.Properties["lastTrigger"].(time.Time)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid last trigger time")
	}

	return lastTrigger.Add(time.Duration(config.CycleInterval) * time.Second), nil
}

// completeTimer completes the timer node
func (h *TimerHandler) completeTimer(ctx context.Context, node *structs.ReadNode) error {
	// Update node status
	_, err := h.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusCompleted),
		},
	})

	return err
}
