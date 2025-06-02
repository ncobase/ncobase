package handler

import (
	"ncobase/workflow/service"
	"ncobase/workflow/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// NodeHandlerInterface defines node handler interface
type NodeHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	UpdateStatus(c *gin.Context)
	GetProcessNodes(c *gin.Context)
	ValidateNodeConfig(c *gin.Context)
	GetNodeTasks(c *gin.Context)
	GetNodeRules(c *gin.Context)
}

// NodeHandler implements node operations
type NodeHandler struct {
	nodeService service.NodeServiceInterface
	taskService service.TaskServiceInterface
	ruleService service.RuleServiceInterface
}

// NewNodeHandler creates new node handler
func NewNodeHandler(svc *service.Service) NodeHandlerInterface {
	return &NodeHandler{
		nodeService: svc.GetNode(),
		taskService: svc.GetTask(),
		ruleService: svc.GetRule(),
	}
}

// Create handles node creation
// @Summary Create node
// @Description Create a new workflow node
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.NodeBody true "Node creation body"
// @Success 200 {object} structs.ReadNode
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes [post]
// @Security Bearer
func (h *NodeHandler) Create(c *gin.Context) {
	body := &structs.NodeBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.nodeService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting node details
// @Summary Get node
// @Description Get workflow node details
// @Tags flow
// @Produce json
// @Param id path string true "Node Key"
// @Success 200 {object} structs.ReadNode
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes/{id} [get]
// @Security Bearer
func (h *NodeHandler) Get(c *gin.Context) {
	params := &structs.FindNodeParams{
		NodeKey: c.Param("id"),
	}

	result, err := h.nodeService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles node update
// @Summary Update node
// @Description Update workflow node details
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Node ID"
// @Param body body structs.UpdateNodeBody true "Node update body"
// @Success 200 {object} structs.ReadNode
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes/{id} [put]
// @Security Bearer
func (h *NodeHandler) Update(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	body := &structs.UpdateNodeBody{ID: nodeID}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.nodeService.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles node deletion
// @Summary Delete node
// @Description Delete a workflow node
// @Tags flow
// @Produce json
// @Param id path string true "Node Key"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes/{id} [delete]
// @Security Bearer
func (h *NodeHandler) Delete(c *gin.Context) {
	params := &structs.FindNodeParams{
		NodeKey: c.Param("id"),
	}

	if err := h.nodeService.Delete(c.Request.Context(), params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles node listing
// @Summary List nodes
// @Description List workflow nodes
// @Tags flow
// @Produce json
// @Param params query structs.ListNodeParams true "Node list parameters"
// @Success 200 {array} structs.ReadNode
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes [get]
// @Security Bearer
func (h *NodeHandler) List(c *gin.Context) {
	params := &structs.ListNodeParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.nodeService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdateStatus handles node status update
// @Summary Update node status
// @Description Update workflow node status
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Node ID"
// @Param status query string true "New Status"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes/{id}/status [put]
// @Security Bearer
func (h *NodeHandler) UpdateStatus(c *gin.Context) {
	nodeID := c.Param("id")
	status := c.Query("status")
	if nodeID == "" || status == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	if err := h.nodeService.UpdateStatus(c.Request.Context(), nodeID, status); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// GetProcessNodes handles getting nodes of a process
// @Summary Get process nodes
// @Description Get all nodes of a process instance
// @Tags flow
// @Produce json
// @Param process_id path string true "Process ID"
// @Success 200 {array} structs.ReadNode
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{process_id}/nodes [get]
// @Security Bearer
func (h *NodeHandler) GetProcessNodes(c *gin.Context) {
	processID := c.Param("process_id")
	if processID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("process_id")))
		return
	}

	result, err := h.nodeService.GetProcessNodes(c.Request.Context(), processID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ValidateNodeConfig handles node configuration validation
// @Summary Validate node config
// @Description Validate configuration of a workflow node
// @Tags flow
// @Produce json
// @Param id path string true "Node ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes/{id}/validate [post]
// @Security Bearer
func (h *NodeHandler) ValidateNodeConfig(c *gin.Context) {
	nodeID := c.Param("id")
	if nodeID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.nodeService.ValidateNodeConfig(c.Request.Context(), nodeID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// GetNodeTasks handles getting tasks of a node
// @Summary Get node tasks
// @Description Get all tasks associated with a node
// @Tags flow
// @Produce json
// @Param id path string true "Node Key"
// @Success 200 {array} structs.ReadTask
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes/{id}/tasks [get]
// @Security Bearer
func (h *NodeHandler) GetNodeTasks(c *gin.Context) {
	params := &structs.ListTaskParams{
		NodeKey: c.Param("id"),
	}

	result, err := h.taskService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetNodeRules handles getting rules of a node
// @Summary Get node rules
// @Description Get all rules associated with a node
// @Tags flow
// @Produce json
// @Param id path string true "Node Key"
// @Success 200 {array} structs.ReadRule
// @Failure 400 {object} resp.Exception
// @Router /flow/nodes/{id}/rules [get]
// @Security Bearer
func (h *NodeHandler) GetNodeRules(c *gin.Context) {
	params := &structs.ListRuleParams{
		NodeKey: c.Param("id"),
	}

	result, err := h.ruleService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
