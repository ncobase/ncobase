package types

import "time"

// EventType represents event type
type EventType string

const (
	// Process events

	EventProcessStarted    EventType = "process.started"
	EventProcessCompleted  EventType = "process.completed"
	EventProcessTerminated EventType = "process.terminated"

	// Node events

	EventNodeStarted   EventType = "node.started"
	EventNodeCompleted EventType = "node.completed"
	EventNodeFailed    EventType = "node.failed"

	// Task events

	EventTaskCreated     EventType = "task.created"
	EventTaskCompleted   EventType = "task.completed"
	EventTaskFailed      EventType = "task.failed"
	EventTaskCancelled   EventType = "task.cancelled"
	EventTaskDelegated   EventType = "task.delegated"
	EventTaskTransferred EventType = "task.transferred"
	EventTaskWithdrawn   EventType = "task.withdrawn"
	EventTaskUrged       EventType = "task.urgent"
	EventTaskTimeout     EventType = "task.timeout"
)

// Event represents a workflow event
type Event struct {
	Type      EventType      // Event type
	ProcessID string         // Process ID
	NodeID    string         // Node ID
	TaskID    string         // Task ID
	Details   map[string]any // Event details
	Timestamp time.Time      // Event timestamp
}
