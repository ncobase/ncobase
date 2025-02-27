package handler

import (
	"context"
	"ncobase/common/extension"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"time"
)

// SubprocessHandler handles subprocess nodes
type SubprocessHandler struct {
	*BaseHandler

	// Configuation
	config *config.SubProcessHandlerConfig
}

// NewSubprocessHandler creates a new subprocess handler
func NewSubprocessHandler(svc *service.Service, em *extension.Manager, cfg *config.Config) *SubprocessHandler {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &SubprocessHandler{
		BaseHandler: NewBaseHandler("subprocess", "Subprocess Handler", svc, em, cfg.Handlers.Base),
		config:      cfg.Handlers.SubProcess,
	}
}

// Type returns handler type
func (h *SubprocessHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *SubprocessHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *SubprocessHandler) Priority() int { return h.priority }

// executeInternal executes the subprocess node
func (h *SubprocessHandler) executeInternal(ctx context.Context, node *structs.ReadNode) error {
	// Parse config
	c, err := h.parseSubprocessConfig(node)
	if err != nil {
		return err
	}

	// Create subprocess instance
	subprocess, err := h.services.Process.Create(ctx, &structs.ProcessBody{
		TemplateID:  c.TemplateID,
		ProcessCode: c.ProcessCode,
		ParentID:    &node.ProcessID,
		Variables:   c.Variables,
	})
	if err != nil {
		return err
	}

	// Update node with subprocess info
	_, err = h.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusActive),
			Properties: map[string]any{
				"subprocessId": subprocess.ID,
			},
		},
	})
	if err != nil {
		return err
	}

	// For synchronous execution, wait for completion
	if c.WaitComplete {
		return h.waitForCompletion(ctx, subprocess.ID, c.Timeout)
	}

	return nil
}

// completeInternal completes the subprocess node
func (h *SubprocessHandler) completeInternal(ctx context.Context, node *structs.ReadNode, req *structs.CompleteTaskRequest) error {
	// Get subprocess id
	subprocessID, ok := node.Properties["subprocessId"].(string)
	if !ok {
		return types.NewError(types.ErrValidation, "subprocess id not found", nil)
	}

	// Get subprocess instance
	subprocess, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: subprocessID,
	})
	if err != nil {
		return err
	}

	// Update parent process variables
	if err := h.updateParentVariables(ctx, node.ProcessID, subprocess.Variables); err != nil {
		return err
	}

	return nil
}

// parseSubprocessConfig parses subprocess configuration
func (h *SubprocessHandler) parseSubprocessConfig(node *structs.ReadNode) (*config.SubProcessHandlerConfig, error) {
	c, ok := node.Properties["subprocessConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing subprocess configuration", nil)
	}

	result := config.DefaultSubProcessHandlerConfig()

	if templateID, ok := c["template_id"].(string); ok {
		result.TemplateID = templateID
	}
	if processCode, ok := c["process_code"].(string); ok {
		result.ProcessCode = processCode
	}
	if variables, ok := c["variables"].(map[string]any); ok {
		result.Variables = variables
	}
	if waitComplete, ok := c["wait_complete"].(bool); ok {
		result.WaitComplete = waitComplete
	}
	if timeout, ok := c["timeout"].(int); ok {
		result.Timeout = timeout
	}

	return result, nil
}

// waitForCompletion waits for subprocess completion
func (h *SubprocessHandler) waitForCompletion(ctx context.Context, processID string, timeout int) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return types.NewError(types.ErrTimeout, "subprocess execution timeout", nil)
		case <-ticker.C:
			process, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
				ProcessKey: processID,
			})
			if err != nil {
				return err
			}

			if process.Status != string(structs.StatusActive) {
				return nil
			}
		}
	}
}

// updateParentVariables updates parent process variables
func (h *SubprocessHandler) updateParentVariables(ctx context.Context, parentID string, variables map[string]any) error {
	parent, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: parentID,
	})
	if err != nil {
		return err
	}

	if parent.Variables == nil {
		parent.Variables = variables
	} else {
		for k, v := range variables {
			parent.Variables[k] = v
		}
	}

	_, err = h.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: parent.ID,
		ProcessBody: structs.ProcessBody{
			Variables: parent.Variables,
		},
	})

	return err
}

// executeAsync executes subprocess in background
func (h *SubprocessHandler) executeAsync(ctx context.Context, node *structs.ReadNode, config *config.SubProcessHandlerConfig) error {
	// Create subprocess in background
	go func() {
		subprocess, err := h.services.Process.Create(ctx, &structs.ProcessBody{
			TemplateID:  config.TemplateID,
			ProcessCode: config.ProcessCode,
			ParentID:    &node.ProcessID,
			Variables:   config.Variables,
		})

		if err != nil {
			h.logger.Error(ctx, "async subprocess creation failed", err)
			return
		}

		// Update node with subprocess info
		_, err = h.services.Node.Update(ctx, &structs.UpdateNodeBody{
			ID: node.ID,
			NodeBody: structs.NodeBody{
				Status: string(structs.StatusActive),
				Properties: map[string]any{
					"subprocessId": subprocess.ID,
					"async":        true,
				},
			},
		})

		if err != nil {
			h.logger.Error(ctx, "update node failed", err)
		}
	}()

	return nil
}
