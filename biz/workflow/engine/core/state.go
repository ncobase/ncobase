package core

import (
	"fmt"
	"sync"
)

// State represents node state transitions
type State string

const (
	StateInit     State = "init"
	StateReady    State = "ready"
	StateRunning  State = "running"
	StateWaiting  State = "waiting"
	StateComplete State = "complete"
	StateFailed   State = "failed"
	StateCanceled State = "canceled"
)

// StateManager handles node state transitions
type StateManager struct {
	states map[string]State
	mu     sync.RWMutex

	// State transition callbacks
	onStateChange func(nodeID string, from, to State)
}

// NewStateManager creates a new node state manager
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]State),
	}
}

// ChangeState changes the state of a node
func (m *StateManager) ChangeState(nodeID string, to State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	from := m.states[nodeID]
	if !m.IsValidTransition(from, to) {
		return fmt.Errorf("invalid state transition: %s -> %s", from, to)
	}

	m.states[nodeID] = to
	if m.onStateChange != nil {
		m.onStateChange(nodeID, from, to)
	}

	return nil
}

// IsValidTransition checks if a state transition is valid
func (m *StateManager) IsValidTransition(from, to State) bool {
	// Define valid state transitions
	transitions := map[State][]State{
		StateInit:     {StateReady},
		StateReady:    {StateRunning, StateCanceled},
		StateRunning:  {StateWaiting, StateComplete, StateFailed, StateCanceled},
		StateWaiting:  {StateRunning, StateCanceled},
		StateComplete: {},
		StateFailed:   {StateReady},
		StateCanceled: {StateReady},
	}

	valid := transitions[from]
	for _, state := range valid {
		if state == to {
			return true
		}
	}
	return false
}
