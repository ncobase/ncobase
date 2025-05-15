package handler

import (
	"ncobase/workflow/service"

	"github.com/gin-gonic/gin"
)

// Handler represents the workflow handler
type Handler struct {
	Process       ProcessHandlerInterface
	Task          TaskHandlerInterface
	Template      TemplateHandlerInterface
	Node          NodeHandlerInterface
	ProcessDesign ProcessDesignHandlerInterface
	Rule          RuleHandlerInterface
	Delegation    DelegationHandlerInterface
	History       HistoryHandlerInterface
}

// New creates new workflow handler
func New(svc *service.Service) *Handler {
	return &Handler{
		Process:       NewProcessHandler(svc),
		Task:          NewTaskHandler(svc),
		Template:      NewTemplateHandler(svc),
		Node:          NewNodeHandler(svc),
		ProcessDesign: NewProcessDesignHandler(svc),
		Rule:          NewRuleHandler(svc),
		Delegation:    NewDelegationHandler(svc),
		History:       NewHistoryHandler(svc),
	}
}

// RegisterRoutes registers all workflow routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Process endpoints
	processes := r.Group("/processes")
	{
		processes.POST("", h.Process.Create)
		processes.GET("", h.Process.List)
		processes.GET("/:id", h.Process.Get)
		processes.PUT("/:id", h.Process.Update)
		processes.DELETE("/:id", h.Process.Delete)
		processes.POST("/:id/start", h.Process.Start)
		processes.POST("/:id/complete", h.Process.Complete)
		processes.POST("/:id/terminate", h.Process.Terminate)
		processes.POST("/:id/suspend", h.Process.Suspend)
		processes.POST("/:id/resume", h.Process.Resume)
		processes.GET("/:id/histories", h.History.GetProcessHistory)
		processes.POST("/:id/evaluate-rules", h.Rule.EvaluateRules)
	}

	// Task endpoints
	tasks := r.Group("/tasks")
	{
		tasks.POST("", h.Task.Create)
		tasks.GET("", h.Task.List)
		tasks.GET("/:id", h.Task.Get)
		tasks.PUT("/:id", h.Task.Update)
		tasks.DELETE("/:id", h.Task.Delete)
		tasks.POST("/:id/complete", h.Task.Complete)
		tasks.POST("/:id/delegate", h.Task.Delegate)
		tasks.POST("/:id/transfer", h.Task.Transfer)
		tasks.POST("/:id/withdraw", h.Task.Withdraw)
		tasks.POST("/:id/urge", h.Task.Urge)
		tasks.POST("/:id/claim", h.Task.Claim)
		tasks.GET("/:id/check-delegation", h.Delegation.CheckDelegation)
		tasks.GET("/:id/histories", h.History.GetTaskHistory)
	}

	// Template endpoints
	templates := r.Group("/templates")
	{
		templates.POST("", h.Template.Create)
		templates.GET("", h.Template.List)
		templates.GET("/:id", h.Template.Get)
		templates.PUT("/:id", h.Template.Update)
		templates.DELETE("/:id", h.Template.Delete)
		templates.POST("/:id/versions", h.Template.CreateVersion)
		templates.PUT("/:id/versions/latest", h.Template.SetLatestVersion)
		templates.POST("/:id/enable", h.Template.Enable)
		templates.POST("/:id/disable", h.Template.Disable)
		templates.POST("/:id/validate", h.Template.ValidateTemplate)
		templates.GET("/:id/designs", h.Template.GetDesigns)
		templates.GET("/:id/rules", h.Template.GetRules)
		templates.POST("/:id/designs/import", h.ProcessDesign.ImportDesign)
	}

	// Node endpoints
	nodes := r.Group("/nodes")
	{
		nodes.POST("", h.Node.Create)
		nodes.GET("", h.Node.List)
		nodes.GET("/:id", h.Node.Get)
		nodes.PUT("/:id", h.Node.Update)
		nodes.DELETE("/:id", h.Node.Delete)
		nodes.PUT("/:id/status", h.Node.UpdateStatus)
		nodes.POST("/:id/validate", h.Node.ValidateNodeConfig)
		nodes.GET("/:id/tasks", h.Node.GetNodeTasks)
		nodes.GET("/:id/rules", h.Node.GetNodeRules)
	}

	// Process Design endpoints
	designs := r.Group("/process-designs")
	{
		designs.POST("", h.ProcessDesign.Create)
		designs.GET("", h.ProcessDesign.List)
		designs.GET("/:id", h.ProcessDesign.Get)
		designs.PUT("/:id", h.ProcessDesign.Update)
		designs.DELETE("/:id", h.ProcessDesign.Delete)
		designs.POST("/:id/drafts", h.ProcessDesign.SaveDraft)
		designs.POST("/:id/publish", h.ProcessDesign.PublishDraft)
		designs.POST("/validate", h.ProcessDesign.ValidateDesign)
		designs.GET("/:id/export", h.ProcessDesign.ExportDesign)
	}

	// Rule endpoints
	rules := r.Group("/rules")
	{
		rules.POST("", h.Rule.Create)
		rules.GET("", h.Rule.List)
		rules.GET("/active", h.Rule.GetActiveRules)
		rules.GET("/:id", h.Rule.Get)
		rules.PUT("/:id", h.Rule.Update)
		rules.DELETE("/:id", h.Rule.Delete)
		rules.POST("/:id/enable", h.Rule.Enable)
		rules.POST("/:id/disable", h.Rule.Disable)
		rules.POST("/:id/validate", h.Rule.ValidateRule)
	}

	// Delegation endpoints
	delegations := r.Group("/delegations")
	{
		delegations.POST("", h.Delegation.Create)
		delegations.GET("", h.Delegation.List)
		delegations.GET("/:id", h.Delegation.Get)
		delegations.PUT("/:id", h.Delegation.Update)
		delegations.DELETE("/:id", h.Delegation.Delete)
		delegations.POST("/:id/enable", h.Delegation.Enable)
		delegations.POST("/:id/disable", h.Delegation.Disable)
	}

	// History endpoints
	histories := r.Group("/histories")
	{
		histories.POST("", h.History.Create)
		histories.GET("", h.History.List)
		histories.GET("/:id", h.History.Get)
	}

	// User related endpoints
	users := r.Group("/users")
	{
		users.GET("/:id/delegations", h.Delegation.GetActiveDelegations)
	}

	// Operator related endpoints
	operators := r.Group("/operators")
	{
		operators.GET("/:id/histories", h.History.GetOperatorHistory)
	}
}
