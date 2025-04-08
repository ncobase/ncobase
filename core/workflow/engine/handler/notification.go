package handler

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"ncore/extension"
	"sync"
	"sync/atomic"
	"text/template/parse"
	"time"
)

// NotificationHandler handles notification nodes
type NotificationHandler struct {
	*BaseHandler

	// Configuration
	config *config.NotificationHandlerConfig

	// Notification providers
	providers map[string]NotificationProvider
	templates *TemplateManager
	rateLimit *RateLimiter

	// Runtime tracking
	notifications sync.Map // ID -> *NotificationInfo

	// Metrics
	metrics *NotificationMetrics
}

// NotificationInfo tracks notification info
type NotificationInfo struct {
	ID         string
	Type       string
	Template   string
	Recipients []string
	Variables  map[string]any
	Status     string
	StartTime  time.Time
	EndTime    *time.Time
	Error      error
	RetryCount int
	mu         sync.RWMutex
}

// NotificationProvider defines the notification provider interface.
type NotificationProvider interface {
	// Basic info

	Type() string
	Name() string

	// Core operations

	Send(ctx context.Context, req *NotificationRequest) error
	ValidateTemplate(template string, variables map[string]any) error

	// Optional operations

	GetTemplateSchema() *TemplateSchema
	GetCapabilities() *NotificationCapabilities
}

// NotificationCapabilities defines provider capabilities
type NotificationCapabilities struct {
	SupportsTemplates bool
	SupportsScheduled bool
	SupportsBatch     bool
	MaxBatchSize      int
	RateLimit         int
	CooldownPeriod    time.Duration
}

// NotificationRequest represents notification request.
type NotificationRequest struct {
	Template   string            // Template name/ID
	Recipients []string          // Recipient list
	Variables  map[string]any    // Template variables
	Options    map[string]string // Additional options
	Schedule   *time.Time        // Optional schedule time
}

// TemplateManager manages notification templates
type TemplateManager struct {
	templates sync.Map // name -> *Template
	mu        sync.RWMutex
}

// Template represents notification template
type Template struct {
	Name        string
	Content     string
	Variables   []string
	Validators  []func(map[string]any) error
	LastUpdated time.Time
}

// TemplateSchema defines template structure
type TemplateSchema struct {
	Variables    map[string]string // Variable name -> type
	Required     []string          // Required variables
	MaxLength    int               // Maximum content length
	AllowedTypes []string          // Allowed content types
}

// RateLimiter implements rate limiting
type RateLimiter struct {
	limits   map[string]int      // type -> limit per interval
	counters map[string]*Counter // type -> counter
	interval time.Duration
	mu       sync.RWMutex
}

type Counter struct {
	count     atomic.Int32
	resetTime time.Time
}

// NotificationMetrics tracks metrics
type NotificationMetrics struct {
	SentCount      atomic.Int64
	FailureCount   atomic.Int64
	RetryCount     atomic.Int64
	ProcessingTime atomic.Int64
	RateLimitCount atomic.Int64
	BatchCount     atomic.Int64
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(svc *service.Service, em *extension.Manager, cfg *config.Config) *NotificationHandler {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	handler := &NotificationHandler{
		BaseHandler: NewBaseHandler("notification", "Notification Handler", svc, em, cfg.Handlers.Base),
		config:      cfg.Handlers.Notification,
		providers:   make(map[string]NotificationProvider),
		templates:   NewTemplateManager(),
		metrics:     &NotificationMetrics{},
	}

	return handler
}

// Type returns handler type
func (h *NotificationHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *NotificationHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *NotificationHandler) Priority() int { return h.priority }

// Start starts the handler
func (h *NotificationHandler) Start() error {
	if err := h.BaseHandler.Start(); err != nil {
		return err
	}

	// Initialize notification providers
	if err := h.registerBuiltinProviders(); err != nil {
		return err
	}

	// Start rate limiter
	h.rateLimit = NewRateLimiter()

	return nil
}

// Stop stops the handler
func (h *NotificationHandler) Stop() error {
	if err := h.BaseHandler.Stop(); err != nil {
		return err
	}

	return nil
}

func (h *NotificationHandler) Execute(ctx context.Context, node *structs.ReadNode) error {
	// Parse config
	cfg, err := h.parseNotificationConfig(node)
	if err != nil {
		return err
	}

	// Get provider
	provider, err := h.getProvider(cfg.Type)
	if err != nil {
		return err
	}

	// Validate template
	if err := provider.ValidateTemplate(cfg.Template, cfg.Variables); err != nil {
		return err
	}

	// Check rate limit
	if !h.rateLimit.Allow(cfg.Type) {
		h.metrics.RateLimitCount.Add(1)
		return types.NewError(types.ErrRateLimit, "rate limit exceeded", nil)
	}

	// Prepare request
	req, err := h.buildRequest(ctx, node, cfg)
	if err != nil {
		return err
	}

	// Track notification
	info := &NotificationInfo{
		ID:         node.ID,
		Type:       cfg.Type,
		Template:   cfg.Template,
		Recipients: cfg.Recipients,
		Variables:  cfg.Variables,
		StartTime:  time.Now(),
		Status:     string(types.ExecutionActive),
	}
	h.notifications.Store(node.ID, info)

	// Send notification
	startTime := time.Now()
	err = provider.Send(ctx, req)
	duration := time.Since(startTime)

	// Update metrics
	if err != nil {
		h.metrics.FailureCount.Add(1)
		info.Error = err
		info.Status = string(types.ExecutionError)
	} else {
		h.metrics.SentCount.Add(1)
		now := time.Now()
		info.EndTime = &now
		info.Status = string(types.ExecutionCompleted)
	}
	h.metrics.ProcessingTime.Add(duration.Nanoseconds())

	// Update notification status
	h.notifications.Store(node.ID, info)

	return err
}

func (h *NotificationHandler) Complete(ctx context.Context, node *structs.ReadNode, req *structs.CompleteTaskRequest) error {
	info, ok := h.notifications.Load(node.ID)
	if !ok {
		return types.NewError(types.ErrNotFound, "notification not found", nil)
	}
	notifInfo := info.(*NotificationInfo)

	// Update notification status
	notifInfo.mu.Lock()
	notifInfo.Status = string(types.ExecutionCompleted)
	now := time.Now()
	notifInfo.EndTime = &now
	notifInfo.mu.Unlock()

	h.notifications.Store(node.ID, notifInfo)

	return nil
}

func (h *NotificationHandler) Validate(node *structs.ReadNode) error {
	// Parse config
	cfg, err := h.parseNotificationConfig(node)
	if err != nil {
		return err
	}

	// Validate provider
	provider, err := h.getProvider(cfg.Type)
	if err != nil {
		return err
	}

	// Validate template
	if err := provider.ValidateTemplate(cfg.Template, node.Variables); err != nil {
		return err
	}

	// Validate recipients
	if len(cfg.Recipients) == 0 {
		return types.NewError(types.ErrValidation, "recipients required", nil)
	}

	return nil
}

// Template manager implementation

func NewTemplateManager() *TemplateManager {
	return &TemplateManager{}
}

func (m *TemplateManager) AddTemplate(name string, content string) error {
	if name == "" || content == "" {
		return types.NewError(types.ErrValidation, "name and content required", nil)
	}

	template := &Template{
		Name:        name,
		Content:     content,
		LastUpdated: time.Now(),
	}

	// Parse variables from content
	vars := parseTemplateVariables(content)
	template.Variables = vars

	// Store template
	m.templates.Store(name, template)
	return nil
}

func (m *TemplateManager) GetTemplate(name string) (*Template, error) {
	value, ok := m.templates.Load(name)
	if !ok {
		return nil, types.NewError(types.ErrNotFound, "template not found", nil)
	}
	return value.(*Template), nil
}

func (m *TemplateManager) RemoveTemplate(name string) {
	m.templates.Delete(name)
}

func (m *TemplateManager) ValidateTemplate(name string, variables map[string]any) error {
	tpl, err := m.GetTemplate(name)
	if err != nil {
		return err
	}

	// Check required variables
	for _, required := range tpl.Variables {
		if _, ok := variables[required]; !ok {
			return fmt.Errorf("missing required variable: %s", required)
		}
	}

	// Run validators
	for _, validator := range tpl.Validators {
		if err := validator(variables); err != nil {
			return err
		}
	}

	return nil
}

// Helper to parse variables from template content
func parseTemplateVariables(content string) []string {
	var vars []string
	tmpl := template.New("example")

	_, err := tmpl.Parse(content)
	if err != nil || tmpl.Tree == nil || tmpl.Tree.Root == nil {
		return vars // Return empty list if parsing fails
	}

	// Recursive function to extract variables from nodes
	var extractVariables func(node parse.Node)
	extractVariables = func(node parse.Node) {
		switch n := node.(type) {
		case *parse.ActionNode:
			// Handle action nodes (e.g., {{ .Var }})
			for _, cmd := range n.Pipe.Cmds {
				for _, arg := range cmd.Args {
					if ident, ok := arg.(*parse.IdentifierNode); ok {
						vars = append(vars, ident.Ident)
					}
				}
			}
		case *parse.ListNode:
			// Recurse into list nodes
			for _, child := range n.Nodes {
				extractVariables(child)
			}
		case *parse.IfNode:
			// Handle IfNode (e.g., {{ if .Condition }})
			if n.Pipe != nil {
				extractVariables(n.Pipe)
			}
			if n.List != nil {
				extractVariables(n.List)
			}
			if n.ElseList != nil {
				extractVariables(n.ElseList)
			}
		case *parse.RangeNode:
			// Handle RangeNode (e.g., {{ range .Items }})
			if n.Pipe != nil {
				extractVariables(n.Pipe)
			}
			if n.List != nil {
				extractVariables(n.List)
			}
			if n.ElseList != nil {
				extractVariables(n.ElseList)
			}
		case *parse.WithNode:
			// Handle WithNode (e.g., {{ with .Context }})
			if n.Pipe != nil {
				extractVariables(n.Pipe)
			}
			if n.List != nil {
				extractVariables(n.List)
			}
			if n.ElseList != nil {
				extractVariables(n.ElseList)
			}
		case *parse.PipeNode:
			// Handle pipelines
			for _, cmd := range n.Cmds {
				for _, arg := range cmd.Args {
					if ident, ok := arg.(*parse.IdentifierNode); ok {
						vars = append(vars, ident.Ident)
					}
				}
			}
		}
	}

	// Start recursion from the root node
	extractVariables(tmpl.Tree.Root)

	return vars
}

// Rate limiter implementation

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits:   make(map[string]int),
		counters: make(map[string]*Counter),
		interval: time.Minute,
	}
}

func (r *RateLimiter) SetLimit(notifType string, limit int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.limits[notifType] = limit
	r.counters[notifType] = &Counter{
		resetTime: time.Now().Add(r.interval),
	}
}

func (r *RateLimiter) Allow(notifType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	limit := r.limits[notifType]
	if limit == 0 {
		return true // No limit
	}

	counter := r.counters[notifType]
	if counter == nil {
		return true
	}

	// Reset counter if interval passed
	if time.Now().After(counter.resetTime) {
		counter.count.Store(0)
		counter.resetTime = time.Now().Add(r.interval)
	}

	// Check limit
	count := counter.count.Add(1)
	return count <= int32(limit)
}

// parseNotificationConfig parses the notification configuration
func (h *NotificationHandler) parseNotificationConfig(node *structs.ReadNode) (*config.NotificationHandlerConfig, error) {
	c, ok := node.Properties["notificationConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing notification config", nil)
	}

	result := config.DefaultNotificationHandlerConfig()

	// Parse required fields
	if notifType, ok := c["type"].(string); ok {
		result.Type = notifType
	} else {
		return nil, types.NewError(types.ErrValidation, "type required", nil)
	}

	if template, ok := c["template"].(string); ok {
		result.Template = template
	}

	if recipients, ok := c["recipients"].([]string); ok {
		result.Recipients = recipients
	}

	// Parse optional fields
	if variables, ok := c["variables"].(map[string]any); ok {
		result.Variables = variables
	}

	if options, ok := c["options"].(map[string]any); ok {
		result.Options = options
	}

	return result, nil
}

// buildRequest builds the notification request
func (h *NotificationHandler) buildRequest(ctx context.Context, node *structs.ReadNode, cfg *config.NotificationHandlerConfig) (*NotificationRequest, error) {
	// Get process for variable substitution
	process, err := h.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: node.ProcessID,
	})
	if err != nil {
		return nil, err
	}

	// Build variables
	vars := make(map[string]any)
	if cfg.Variables != nil {
		for k, v := range cfg.Variables {
			// Handle variable substitution
			if str, ok := v.(string); ok && len(str) > 0 && str[0] == '$' {
				varName := str[1:]
				if val, exists := process.Variables[varName]; exists {
					vars[k] = val
					continue
				}
			}
			vars[k] = v
		}
	}

	// Build options
	options := make(map[string]string)
	if cfg.Options != nil {
		for k, v := range cfg.Options {
			options[k] = fmt.Sprint(v)
		}
	}

	return &NotificationRequest{
		Template:   cfg.Template,
		Recipients: cfg.Recipients,
		Variables:  vars,
		Options:    options,
	}, nil
}

// getProvider returns the notification provider
func (h *NotificationHandler) getProvider(notifType string) (NotificationProvider, error) {
	provider, ok := h.providers[notifType]
	if !ok {
		return nil, types.NewError(types.ErrNotFound, "provider not found", nil)
	}
	return provider, nil
}

// Builtin providers registration
func (h *NotificationHandler) registerBuiltinProviders() error {
	h.providers = make(map[string]NotificationProvider)

	// Email provider
	emailProvider, err := NewEmailProvider(&EmailProviderConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "noreply@example.com",
	})
	if err != nil {
		log.Printf("Error initializing email provider: %v\n", err)
		return err
	}
	h.providers["email"] = emailProvider

	// SMS provider
	smsProvider, err := NewSMSProvider(&SMSProviderConfig{
		AccountSID: "account_sid",
		AuthToken:  "auth_token",
	})
	if err != nil {
		log.Printf("Error initializing SMS provider: %v\n", err)
		return err
	}
	h.providers["sms"] = smsProvider

	// Push notification provider
	pushProvider, err := NewPushProvider(&PushProviderConfig{
		ProjectID: "project_id",
		APIKey:    "api_key",
	})
	if err != nil {
		log.Printf("Error initializing push notification provider: %v\n", err)
		return err
	}
	h.providers["push"] = pushProvider

	// Webhook provider
	webhookProvider, err := NewWebhookProvider(&WebhookProviderConfig{
		Endpoint:  "https://example.com/webhook",
		AuthToken: "auth_token",
	})
	if err != nil {
		log.Printf("Error initializing webhook provider: %v\n", err)
		return err
	}
	h.providers["webhook"] = webhookProvider

	log.Println("All providers registered successfully")
	return nil
}

// Email provider implementation

// EmailProvider represents an email notification provider
type EmailProvider struct {
	config *EmailProviderConfig
}

// EmailProviderConfig represents email provider configuration
type EmailProviderConfig struct {
	SMTPHost string
	SMTPPort int
	From     string
}

// NewEmailProvider creates a new email provider
func NewEmailProvider(cfg *EmailProviderConfig) (*EmailProvider, error) {
	if cfg.SMTPHost == "" || cfg.SMTPPort == 0 || cfg.From == "" {
		return nil, errors.New("invalid email provider configuration")
	}
	return &EmailProvider{config: cfg}, nil
}

// Type returns the provider type
func (p *EmailProvider) Type() string {
	return "email"
}

// Name returns the provider name
func (p *EmailProvider) Name() string {
	return "EmailProvider"
}

// Send sends an email
func (p *EmailProvider) Send(ctx context.Context, req *NotificationRequest) error {
	if len(req.Recipients) == 0 {
		return errors.New("no recipients specified")
	}
	if req.Template == "" {
		return errors.New("template is required")
	}
	fmt.Printf("Sending email to %v using template '%s'\n", req.Recipients, req.Template)
	return nil
}

// ValidateTemplate validates a template
func (p *EmailProvider) ValidateTemplate(template string, variables map[string]any) error {
	if template == "" {
		return errors.New("template cannot be empty")
	}
	// Check for required variables
	schema := p.GetTemplateSchema()
	for _, required := range schema.Required {
		if _, ok := variables[required]; !ok {
			return fmt.Errorf("missing required variable: %s", required)
		}
	}
	// TODO: template validation logic
	return nil
}

// GetTemplateSchema returns the template schema
func (p *EmailProvider) GetTemplateSchema() *TemplateSchema {
	return &TemplateSchema{
		Variables: map[string]string{
			"RecipientName": "string",
			"Subject":       "string",
			"Body":          "string",
		},
		Required:     []string{"RecipientName", "Subject", "Body"},
		MaxLength:    2000,
		AllowedTypes: []string{"text/html", "text/plain"},
	}
}

// GetCapabilities returns the provider capabilities
func (p *EmailProvider) GetCapabilities() *NotificationCapabilities {
	return &NotificationCapabilities{
		SupportsTemplates: true,
		SupportsScheduled: true,
		SupportsBatch:     true,
		MaxBatchSize:      100,
		RateLimit:         50,
		CooldownPeriod:    2 * time.Second,
	}
}

// SMS provider implementation

// SMSProvider represents a SMS provider
type SMSProvider struct {
	config *SMSProviderConfig
}

// SMSProviderConfig defines the configuration for the SMS provider
type SMSProviderConfig struct {
	AccountSID string
	AuthToken  string
}

// NewSMSProvider creates a new SMS provider
func NewSMSProvider(cfg *SMSProviderConfig) (*SMSProvider, error) {
	if cfg.AccountSID == "" || cfg.AuthToken == "" {
		return nil, errors.New("invalid SMS provider configuration")
	}
	return &SMSProvider{config: cfg}, nil
}

// Type returns the type of the SMS provider
func (p *SMSProvider) Type() string {
	return "sms"
}

// Name returns the name of the SMS provider
func (p *SMSProvider) Name() string {
	return "SMSProvider"
}

// Send sends an SMS notification
func (p *SMSProvider) Send(ctx context.Context, req *NotificationRequest) error {
	if len(req.Recipients) == 0 {
		return errors.New("no recipients specified")
	}
	if req.Template == "" {
		return errors.New("template is required")
	}
	// TODO: sending SMS here.
	fmt.Printf("Sending SMS to %v using template '%s'\n", req.Recipients, req.Template)
	return nil
}

// ValidateTemplate validates the SMS template
func (p *SMSProvider) ValidateTemplate(template string, variables map[string]any) error {
	if template == "" {
		return errors.New("template cannot be empty")
	}
	// TODO: SMS template validation logic
	return nil
}

// GetTemplateSchema returns the schema for the SMS template
func (p *SMSProvider) GetTemplateSchema() *TemplateSchema {
	return &TemplateSchema{
		Variables: map[string]string{
			"RecipientName": "string",
			"Message":       "string",
		},
		Required:     []string{"RecipientName", "Message"},
		MaxLength:    500,
		AllowedTypes: []string{"text/plain"},
	}
}

// GetCapabilities returns the capabilities of the provider
func (p *SMSProvider) GetCapabilities() *NotificationCapabilities {
	return &NotificationCapabilities{
		SupportsTemplates: true,
		SupportsScheduled: false,
		SupportsBatch:     false,
		MaxBatchSize:      1,
		RateLimit:         100,
		CooldownPeriod:    1 * time.Second,
	}
}

// Push notification provider implementation

// PushProvider is a push notification provider
type PushProvider struct {
	config *PushProviderConfig
}

// PushProviderConfig is the configuration for the PushProvider
type PushProviderConfig struct {
	ProjectID string
	APIKey    string
}

// NewPushProvider creates a new PushProvider
func NewPushProvider(cfg *PushProviderConfig) (*PushProvider, error) {
	if cfg.ProjectID == "" || cfg.APIKey == "" {
		return nil, errors.New("invalid Push provider configuration")
	}
	return &PushProvider{config: cfg}, nil
}

// Type returns the type of the provider
func (p *PushProvider) Type() string {
	return "push"
}

// Name returns the name of the provider
func (p *PushProvider) Name() string {
	return "PushProvider"
}

// Send sends a push notification
func (p *PushProvider) Send(ctx context.Context, req *NotificationRequest) error {
	if len(req.Recipients) == 0 {
		return errors.New("no recipients specified")
	}
	if req.Template == "" {
		return errors.New("template is required")
	}
	// TODO: sending push notification here.
	fmt.Printf("Sending push notification to %v using template '%s'\n", req.Recipients, req.Template)
	return nil
}

// ValidateTemplate validates the template
func (p *PushProvider) ValidateTemplate(template string, variables map[string]any) error {
	if template == "" {
		return errors.New("template cannot be empty")
	}
	// TODO: Validate push-specific template requirements.
	return nil
}

// GetTemplateSchema returns the template schema
func (p *PushProvider) GetTemplateSchema() *TemplateSchema {
	return &TemplateSchema{
		Variables: map[string]string{
			"Title":   "string",
			"Message": "string",
		},
		Required:     []string{"Title", "Message"},
		MaxLength:    500,
		AllowedTypes: []string{"text/plain"},
	}
}

// GetCapabilities returns the capabilities of the provider
func (p *PushProvider) GetCapabilities() *NotificationCapabilities {
	return &NotificationCapabilities{
		SupportsTemplates: true,
		SupportsScheduled: true,
		SupportsBatch:     true,
		MaxBatchSize:      50,
		RateLimit:         100,
		CooldownPeriod:    1 * time.Second,
	}
}

// Webhook provider implementation

// WebhookProvider is a webhook provider
type WebhookProvider struct {
	config *WebhookProviderConfig
}

// WebhookProviderConfig is the configuration for the WebhookProvider
type WebhookProviderConfig struct {
	Endpoint  string
	AuthToken string
}

// NewWebhookProvider creates a new WebhookProvider
func NewWebhookProvider(cfg *WebhookProviderConfig) (*WebhookProvider, error) {
	if cfg.Endpoint == "" || cfg.AuthToken == "" {
		return nil, errors.New("invalid Webhook provider configuration")
	}
	return &WebhookProvider{config: cfg}, nil
}

// Type returns the type of the provider
func (p *WebhookProvider) Type() string {
	return "webhook"
}

// Name returns the name of the provider
func (p *WebhookProvider) Name() string {
	return "WebhookProvider"
}

// Send sends a webhook
func (p *WebhookProvider) Send(ctx context.Context, req *NotificationRequest) error {
	if req.Template == "" {
		return errors.New("template is required")
	}
	payload := map[string]any{
		"recipients": req.Recipients,
		"template":   req.Template,
		"variables":  req.Variables,
		"options":    req.Options,
	}
	// TODO: sending a webhook.
	fmt.Printf("Sending webhook to %s with payload: %v\n", p.config.Endpoint, payload)
	return nil
}

// ValidateTemplate validates the template
func (p *WebhookProvider) ValidateTemplate(template string, variables map[string]any) error {
	if template == "" {
		return errors.New("template cannot be empty")
	}
	// TODO: Validate webhook-specific requirements.
	return nil
}

// GetTemplateSchema returns the template schema
func (p *WebhookProvider) GetTemplateSchema() *TemplateSchema {
	return &TemplateSchema{
		Variables: map[string]string{
			"Event":   "string",
			"Payload": "json",
		},
		Required:     []string{"Event", "Payload"},
		MaxLength:    2000,
		AllowedTypes: []string{"application/json"},
	}
}

// GetCapabilities returns the capabilities of the provider
func (p *WebhookProvider) GetCapabilities() *NotificationCapabilities {
	return &NotificationCapabilities{
		SupportsTemplates: true,
		SupportsScheduled: false,
		SupportsBatch:     false,
		MaxBatchSize:      1,
		RateLimit:         10,
		CooldownPeriod:    5 * time.Second,
	}
}
