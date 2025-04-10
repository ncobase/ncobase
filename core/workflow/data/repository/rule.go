package repository

import (
	"context"
	"ncobase/core/workflow/data"
	"ncobase/core/workflow/data/ent"
	ruleEnt "ncobase/core/workflow/data/ent/rule"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/pkg/data/cache"
	"github.com/ncobase/ncore/pkg/data/meili"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/validator"

	"github.com/redis/go-redis/v9"
)

type RuleRepositoryInterface interface {
	Create(context.Context, *structs.RuleBody) (*ent.Rule, error)
	Get(context.Context, *structs.FindRuleParams) (*ent.Rule, error)
	Update(context.Context, *structs.UpdateRuleBody) (*ent.Rule, error)
	List(context.Context, *structs.ListRuleParams) ([]*ent.Rule, error)
	Delete(context.Context, string) error

	EnableRule(ctx context.Context, id string) error
	DisableRule(ctx context.Context, id string) error
	GetTemplateRules(ctx context.Context, templateID string) ([]*ent.Rule, error)
	GetNodeRules(ctx context.Context, nodeKey string) ([]*ent.Rule, error)
}

type ruleRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Rule]
}

func NewRuleRepository(d *data.Data) RuleRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &ruleRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Rule](rc, "workflow_rule", false),
	}
}

// Create creates a new rule
func (r *ruleRepository) Create(ctx context.Context, body *structs.RuleBody) (*ent.Rule, error) {
	builder := r.ec.Rule.Create()

	if body.Name != "" {
		builder.SetName(body.Name)
	}
	if body.Code != "" {
		builder.SetCode(body.Code)
	}
	if body.Description != "" {
		builder.SetDescription(body.Description)
	}
	if body.Type != "" {
		builder.SetType(body.Type)
	}
	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.RuleKey != "" {
		builder.SetRuleKey(body.RuleKey)
	}
	if body.TemplateID != "" {
		builder.SetTemplateID(body.TemplateID)
	}
	if body.NodeKey != "" {
		builder.SetNodeKey(body.NodeKey)
	}
	if body.Conditions != nil {
		builder.SetConditions(body.Conditions)
	}
	if body.Actions != nil {
		builder.SetActions(body.Actions)
	}

	builder.SetPriority(body.Priority)
	builder.SetIsEnabled(body.IsEnabled)

	if body.EffectiveTime != nil {
		builder.SetEffectiveTime(*body.EffectiveTime)
	}
	if body.ExpireTime != nil {
		builder.SetExpireTime(*body.ExpireTime)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "ruleRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("rules", row); err != nil {
		logger.Errorf(ctx, "ruleRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific rule
func (r *ruleRepository) Get(ctx context.Context, params *structs.FindRuleParams) (*ent.Rule, error) {
	builder := r.ec.Rule.Query()

	if params.RuleKey != "" {
		builder.Where(ruleEnt.RuleKeyEQ(params.RuleKey))
	}
	if params.TemplateID != "" {
		builder.Where(ruleEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.NodeKey != "" {
		builder.Where(ruleEnt.NodeKeyEQ(params.NodeKey))
	}
	if params.Type != "" {
		builder.Where(ruleEnt.TypeEQ(params.Type))
	}
	if params.Status != "" {
		builder.Where(ruleEnt.StatusEQ(params.Status))
	}
	if params.IsEnabled != nil {
		builder.Where(ruleEnt.IsEnabledEQ(*params.IsEnabled))
	}

	return builder.First(ctx)
}

// Update updates a rule
func (r *ruleRepository) Update(ctx context.Context, body *structs.UpdateRuleBody) (*ent.Rule, error) {
	builder := r.ec.Rule.UpdateOneID(body.ID)

	if body.Name != "" {
		builder.SetName(body.Name)
	}
	if body.Description != "" {
		builder.SetDescription(body.Description)
	}
	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.Conditions != nil {
		builder.SetConditions(body.Conditions)
	}
	if body.Actions != nil {
		builder.SetActions(body.Actions)
	}

	builder.SetPriority(body.Priority)
	builder.SetIsEnabled(body.IsEnabled)

	if body.EffectiveTime != nil {
		builder.SetEffectiveTime(*body.EffectiveTime)
	}
	if body.ExpireTime != nil {
		builder.SetExpireTime(*body.ExpireTime)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	return builder.Save(ctx)
}

// List returns a list of rules
func (r *ruleRepository) List(ctx context.Context, params *structs.ListRuleParams) ([]*ent.Rule, error) {
	builder := r.ec.Rule.Query()

	if params.Type != "" {
		builder.Where(ruleEnt.TypeEQ(params.Type))
	}
	if params.Status != "" {
		builder.Where(ruleEnt.StatusEQ(params.Status))
	}
	if params.TemplateID != "" {
		builder.Where(ruleEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.NodeKey != "" {
		builder.Where(ruleEnt.NodeKeyEQ(params.NodeKey))
	}
	if params.IsEnabled != nil {
		builder.Where(ruleEnt.IsEnabledEQ(*params.IsEnabled))
	}

	builder.Order(ent.Desc(ruleEnt.FieldPriority))
	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// Delete deletes a rule
func (r *ruleRepository) Delete(ctx context.Context, id string) error {
	return r.ec.Rule.DeleteOneID(id).Exec(ctx)
}

// EnableRule enables a rule
func (r *ruleRepository) EnableRule(ctx context.Context, id string) error {
	return r.ec.Rule.UpdateOneID(id).
		SetIsEnabled(true).
		Exec(ctx)
}

// DisableRule disables a rule
func (r *ruleRepository) DisableRule(ctx context.Context, id string) error {
	return r.ec.Rule.UpdateOneID(id).
		SetIsEnabled(false).
		Exec(ctx)
}

// GetTemplateRules gets rules for a template
func (r *ruleRepository) GetTemplateRules(ctx context.Context, templateID string) ([]*ent.Rule, error) {
	return r.ec.Rule.Query().
		Where(
			ruleEnt.TemplateIDEQ(templateID),
			ruleEnt.IsEnabledEQ(true),
		).
		Order(ent.Desc(ruleEnt.FieldPriority)).
		All(ctx)
}

// GetNodeRules gets rules for a node
func (r *ruleRepository) GetNodeRules(ctx context.Context, nodeKey string) ([]*ent.Rule, error) {
	return r.ec.Rule.Query().
		Where(
			ruleEnt.NodeKeyEQ(nodeKey),
			ruleEnt.IsEnabledEQ(true),
		).
		Order(ent.Desc(ruleEnt.FieldPriority)).
		All(ctx)
}
