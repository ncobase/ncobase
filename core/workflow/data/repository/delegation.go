package repository

import (
	"context"
	"ncobase/workflow/data"
	"ncobase/workflow/data/ent"
	delegationEnt "ncobase/workflow/data/ent/delegation"
	"ncobase/workflow/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

type DelegationRepositoryInterface interface {
	Create(context.Context, *structs.DelegationBody) (*ent.Delegation, error)
	Get(context.Context, *structs.FindDelegationParams) (*ent.Delegation, error)
	Update(context.Context, *structs.UpdateDelegationBody) (*ent.Delegation, error)
	List(context.Context, *structs.ListDelegationParams) ([]*ent.Delegation, error)
	Delete(context.Context, string) error

	EnableDelegation(ctx context.Context, id string) error
	DisableDelegation(ctx context.Context, id string) error
	GetActiveDelegations(ctx context.Context, delegatorID string) ([]*ent.Delegation, error)
}

type delegationRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Delegation]
}

func NewDelegationRepository(d *data.Data) DelegationRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &delegationRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Delegation](rc, "workflow_delegation", false),
	}
}

// Create creates a new delegation
func (r *delegationRepository) Create(ctx context.Context, body *structs.DelegationBody) (*ent.Delegation, error) {
	builder := r.ec.Delegation.Create()

	if body.DelegatorID != "" {
		builder.SetDelegatorID(body.DelegatorID)
	}
	if body.DelegateeID != "" {
		builder.SetDelegateeID(body.DelegateeID)
	}
	if body.TemplateID != "" {
		builder.SetTemplateID(body.TemplateID)
	}
	if body.NodeType != "" {
		builder.SetNodeType(body.NodeType)
	}
	if body.Conditions != nil {
		builder.SetConditions(body.Conditions)
	}

	builder.SetStartTime(body.StartTime)
	builder.SetEndTime(body.EndTime)
	builder.SetIsEnabled(body.IsEnabled)

	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "delegationRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("delegations", row); err != nil {
		logger.Errorf(ctx, "delegationRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific delegation
func (r *delegationRepository) Get(ctx context.Context, params *structs.FindDelegationParams) (*ent.Delegation, error) {
	builder := r.ec.Delegation.Query()

	if params.DelegatorID != "" {
		builder.Where(delegationEnt.DelegatorIDEQ(params.DelegatorID))
	}
	if params.DelegateeID != "" {
		builder.Where(delegationEnt.DelegateeIDEQ(params.DelegateeID))
	}
	if params.TemplateID != "" {
		builder.Where(delegationEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.NodeType != "" {
		builder.Where(delegationEnt.NodeTypeEQ(params.NodeType))
	}
	if params.IsEnabled != nil {
		builder.Where(delegationEnt.IsEnabledEQ(*params.IsEnabled))
	}
	if params.Status != "" {
		builder.Where(delegationEnt.StatusEQ(params.Status))
	}
	if params.StartTime != nil {
		builder.Where(delegationEnt.StartTimeLTE(*params.StartTime))
	}
	if params.EndTime != nil {
		builder.Where(delegationEnt.EndTimeGTE(*params.EndTime))
	}

	return builder.First(ctx)
}

// Update updates a delegation
func (r *delegationRepository) Update(ctx context.Context, body *structs.UpdateDelegationBody) (*ent.Delegation, error) {
	builder := r.ec.Delegation.UpdateOneID(body.ID)

	if body.DelegateeID != "" {
		builder.SetDelegateeID(body.DelegateeID)
	}
	if body.TemplateID != "" {
		builder.SetTemplateID(body.TemplateID)
	}
	if body.NodeType != "" {
		builder.SetNodeType(body.NodeType)
	}
	if body.Conditions != nil {
		builder.SetConditions(body.Conditions)
	}

	builder.SetStartTime(body.StartTime)
	builder.SetEndTime(body.EndTime)
	builder.SetIsEnabled(body.IsEnabled)

	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	return builder.Save(ctx)
}

// List returns a list of delegations
func (r *delegationRepository) List(ctx context.Context, params *structs.ListDelegationParams) ([]*ent.Delegation, error) {
	builder := r.ec.Delegation.Query()

	if params.DelegatorID != "" {
		builder.Where(delegationEnt.DelegatorIDEQ(params.DelegatorID))
	}
	if params.TemplateID != "" {
		builder.Where(delegationEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.NodeType != "" {
		builder.Where(delegationEnt.NodeTypeEQ(params.NodeType))
	}
	if params.IsEnabled != nil {
		builder.Where(delegationEnt.IsEnabledEQ(*params.IsEnabled))
	}
	if params.Status != "" {
		builder.Where(delegationEnt.StatusEQ(params.Status))
	}

	builder.Order(ent.Desc(delegationEnt.FieldCreatedAt))
	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// Delete deletes a delegation
func (r *delegationRepository) Delete(ctx context.Context, id string) error {
	return r.ec.Delegation.DeleteOneID(id).Exec(ctx)
}

// EnableDelegation enables a delegation
func (r *delegationRepository) EnableDelegation(ctx context.Context, id string) error {
	return r.ec.Delegation.UpdateOneID(id).
		SetIsEnabled(true).
		Exec(ctx)
}

// DisableDelegation disables a delegation
func (r *delegationRepository) DisableDelegation(ctx context.Context, id string) error {
	return r.ec.Delegation.UpdateOneID(id).
		SetIsEnabled(false).
		Exec(ctx)
}

// GetActiveDelegations gets active delegations for a delegator
func (r *delegationRepository) GetActiveDelegations(ctx context.Context, delegatorID string) ([]*ent.Delegation, error) {
	return r.ec.Delegation.Query().
		Where(
			delegationEnt.DelegatorIDEQ(delegatorID),
			delegationEnt.IsEnabledEQ(true),
			delegationEnt.StartTimeLTE(time.Now().UnixMilli()),
			delegationEnt.EndTimeGTE(time.Now().UnixMilli()),
		).
		Order(ent.Desc(delegationEnt.FieldCreatedAt)).
		All(ctx)
}
