package handler

import (
	"context"
	"fmt"
	nec "github.com/ncobase/ncore/ext/core"
	"github.com/ncobase/ncore/pkg/expression"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"sort"
	"time"
)

// ExclusiveHandler handles exclusive gateway nodes
type ExclusiveHandler struct {
	*BaseHandler
	// Configuration
	config *config.ExclusiveHandlerConfig

	// Dependencies
	expression *expression.Expression
}

// NewExclusiveHandler creates a new exclusive handler
func NewExclusiveHandler(svc *service.Service, em nec.ManagerInterface, expr *expression.Expression, cfg *config.Config) *ExclusiveHandler {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &ExclusiveHandler{
		BaseHandler: NewBaseHandler("exclusive", "Exclusive Gateway Handler", svc, em, cfg.Handlers.Base),
		config:      cfg.Handlers.Exclusive,
		expression:  expr,
	}
}

// Type returns handler type
func (h *ExclusiveHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *ExclusiveHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *ExclusiveHandler) Priority() int { return h.priority }

// executeInternal executes the exclusive gateway node
func (h *ExclusiveHandler) executeInternal(ctx context.Context, node *structs.ReadNode) error {
	// Parse config
	c, err := h.parseConfig(node)
	if err != nil {
		return err
	}

	// Get process variables
	process, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: node.ProcessID,
	})
	if err != nil {
		return err
	}

	// Evaluate conditions and get next node
	nextNode, err := h.evaluateConditions(ctx, node, c, process.Variables)
	if err != nil {
		if c.FailureMode == "continue" && c.DefaultPath != "" {
			// Use default path on evaluation failure
			nextNode = c.DefaultPath
		} else {
			return err
		}
	}

	// Execute next node
	return h.executeNextNode(ctx, node.ProcessID, nextNode)
}

// completeInternal completes the exclusive gateway node
func (h *ExclusiveHandler) completeInternal(ctx context.Context, node *structs.ReadNode, req *structs.CompleteTaskRequest) error {
	// Exclusive gateway completes immediately after execution
	return nil
}

// validateInternal validates the exclusive gateway node
func (h *ExclusiveHandler) validateInternal(node *structs.ReadNode) error {
	// Parse config
	c, err := h.parseConfig(node)
	if err != nil {
		return err
	}

	// Validate conditions
	for _, condition := range c.Conditions {
		// Validate expression syntax
		if err := h.validateExpression(condition.Expression); err != nil {
			return fmt.Errorf("invalid condition expression: %w", err)
		}

		// Validate next node exists
		if err := h.validateNodeExists(node.ProcessID, condition.NextNode); err != nil {
			return fmt.Errorf("invalid next node: %w", err)
		}
	}

	// Validate default path if specified
	if c.DefaultPath != "" {
		if err := h.validateNodeExists(node.ProcessID, c.DefaultPath); err != nil {
			return fmt.Errorf("invalid default path: %w", err)
		}
	}

	return nil
}

// parseConfig parses the exclusive gateway configuration
func (h *ExclusiveHandler) parseConfig(node *structs.ReadNode) (*config.ExclusiveHandlerConfig, error) {
	cfg, ok := node.Properties["gatewayConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing gateway configuration", nil)
	}

	result := &config.ExclusiveHandlerConfig{
		FailureMode: "error", // Default failure mode
	}

	// Parse conditions
	if conditions, ok := cfg["conditions"].([]any); ok {
		result.Conditions = make([]config.Condition, 0, len(conditions))
		for _, c := range conditions {
			condition, ok := c.(map[string]any)
			if !ok {
				continue
			}

			expr, ok := condition["expression"].(string)
			if !ok {
				continue
			}

			nextNode, ok := condition["next_node"].(string)
			if !ok {
				continue
			}

			priority := 0
			if p, ok := condition["priority"].(int); ok {
				priority = p
			}

			result.Conditions = append(result.Conditions, config.Condition{
				Expression: expr,
				NextNode:   nextNode,
				Priority:   priority,
			})
		}
	}

	// Parse default path
	if defaultPath, ok := cfg["default_path"].(string); ok {
		result.DefaultPath = defaultPath
	}

	// Parse failure mode
	if failureMode, ok := cfg["failure_mode"].(string); ok {
		result.FailureMode = failureMode
	}

	return result, nil
}

// evaluateConditions evaluates the conditions and returns the next node
func (h *ExclusiveHandler) evaluateConditions(
	ctx context.Context,
	node *structs.ReadNode,
	cfg *config.ExclusiveHandlerConfig,
	variables map[string]any,
) (string, error) {
	if len(cfg.Conditions) == 0 {
		if cfg.DefaultPath == "" {
			return "", types.NewError(types.ErrValidation, "no conditions and no default path", nil)
		}
		return cfg.DefaultPath, nil
	}

	// Sort conditions by priority
	conditions := make([]config.Condition, len(cfg.Conditions))
	copy(conditions, cfg.Conditions)
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i].Priority > conditions[j].Priority
	})

	// Evaluate conditions in priority order
	for _, condition := range conditions {
		result, err := h.evaluateCondition(ctx, condition.Expression, variables)
		if err != nil {
			continue
		}

		if result {
			return condition.NextNode, nil
		}
	}

	// No condition matched, use default path
	if cfg.DefaultPath != "" {
		return cfg.DefaultPath, nil
	}

	return "", types.NewError(types.ErrValidation, "no matching condition and no default path", nil)
}

// evaluateCondition evaluates a condition
func (h *ExclusiveHandler) evaluateCondition(
	ctx context.Context,
	expression string,
	variables map[string]any,
) (bool, error) {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Evaluate expression
	result, err := h.expression.Evaluate(timeoutCtx, expression, variables)
	if err != nil {
		return false, fmt.Errorf("expression evaluation failed: %w", err)
	}

	// Convert result to boolean
	switch v := result.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return v != "", nil
	default:
		return false, fmt.Errorf("expression must evaluate to boolean")
	}
}

// validateExpression validates the syntax of the expression
func (h *ExclusiveHandler) validateExpression(expression string) error {
	if expression == "" {
		return types.NewError(types.ErrValidation, "empty expression", nil)
	}
	return h.expression.ValidateSyntax(expression)
}

func (h *ExclusiveHandler) validateNodeExists(processID string, nodeKey string) error {
	_, err := h.services.Node.Get(context.Background(), &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   nodeKey,
	})
	return err
}

// executeNextNode executes the next node
func (h *ExclusiveHandler) executeNextNode(ctx context.Context, processID string, nodeKey string) error {
	// Get next node
	_, err := h.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   nodeKey,
	})
	if err != nil {
		return fmt.Errorf("failed to get next node: %w", err)
	}

	// Update process current node
	_, err = h.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: processID,
		ProcessBody: structs.ProcessBody{
			CurrentNode: nodeKey,
			ActiveNodes: []string{nodeKey},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update process: %w", err)
	}

	// TODO: Execute node
	// Execute next node
	// req := &structs.NodeBody{
	// 	ProcessID: processID,
	// 	NodeKey:   nodeKey,
	// }
	// _, err = h.services.Node.Execute(ctx, req)
	return err
}
