package types

// NodeType represents type of node
type NodeType string

const (
	NodeStart     NodeType = "start"
	NodeApproval  NodeType = "approval"
	NodeService   NodeType = "service"
	NodeExclusive NodeType = "exclusive"
	NodeParallel  NodeType = "parallel"
	NodeEnd       NodeType = "end"
)

// Node represents a workflow node

type Node struct {
	ID         string          // Node ID
	Type       NodeType        // Node type
	ProcessID  string          // Process ID
	PrevNodes  []string        // Previous nodes
	NextNodes  []string        // Next nodes
	Properties map[string]any  // Node properties
	Variables  map[string]any  // Node variables
	Status     ExecutionStatus // Node status
	StartTime  *int64          // Start time
	EndTime    *int64          // End time
}
