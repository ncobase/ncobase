package core

import (
	"sync"
)

// DataFlowManager manages node data flow
type DataFlowManager struct {
	dataStore map[string]map[string]any
	mu        sync.RWMutex
}

// NewDataFlowManager creates a new data flow manager
func NewDataFlowManager() *DataFlowManager {
	return &DataFlowManager{
		dataStore: make(map[string]map[string]any),
	}
}

func (m *DataFlowManager) SetNodeData(nodeID string, key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.dataStore[nodeID] == nil {
		m.dataStore[nodeID] = make(map[string]any)
	}
	m.dataStore[nodeID][key] = value
}

func (m *DataFlowManager) GetNodeData(nodeID string, key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if data, ok := m.dataStore[nodeID]; ok {
		value, exists := data[key]
		return value, exists
	}
	return nil, false
}

// Metrics returns core engine metrics
func (c *Core) Metrics() map[string]any {
	return map[string]any{}
}
