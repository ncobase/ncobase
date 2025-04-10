package repository

import (
	"context"
	"ncobase/core/workflow/data"
	"ncobase/core/workflow/data/ent"
	businessEnt "ncobase/core/workflow/data/ent/business"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/pkg/data/cache"
	"github.com/ncobase/ncore/pkg/data/meili"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/types"
	"github.com/ncobase/ncore/pkg/validator"

	"github.com/redis/go-redis/v9"
)

type BusinessRepositoryInterface interface {
	Create(context.Context, *structs.BusinessBody) (*ent.Business, error)
	Get(context.Context, *structs.FindBusinessParams) (*ent.Business, error)
	Update(context.Context, *structs.UpdateBusinessBody) (*ent.Business, error)
	Delete(context.Context, *structs.FindBusinessParams) error
	List(context.Context, *structs.ListBusinessParams) ([]*ent.Business, error)
	CountX(context.Context, *structs.ListBusinessParams) int
	UpdateFlowStatus(context.Context, string, string) error
	UpdateBusinessData(context.Context, string, types.JSON) error
}

type businessRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Business]
}

func NewBusinessRepository(d *data.Data) BusinessRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &businessRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Business](rc, "workflow_business", false),
	}
}

// Create creates a new business record
func (r *businessRepository) Create(ctx context.Context, body *structs.BusinessBody) (*ent.Business, error) {
	builder := r.ec.Business.Create()

	if body.Code != "" {
		builder.SetCode(body.Code)
	}
	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.ModuleCode != "" {
		builder.SetModuleCode(body.ModuleCode)
	}
	if body.FormCode != "" {
		builder.SetFormCode(body.FormCode)
	}
	if body.ProcessID != "" {
		builder.SetProcessID(body.ProcessID)
	}
	if body.TemplateID != "" {
		builder.SetTemplateID(body.TemplateID)
	}
	if body.FlowStatus != "" {
		builder.SetFlowStatus(body.FlowStatus)
	}

	builder.SetIsDraft(body.IsDraft)

	if body.OriginData != nil {
		builder.SetOriginData(body.OriginData)
	}
	if body.CurrentData != nil {
		builder.SetCurrentData(body.CurrentData)
	}
	if body.Variables != nil {
		builder.SetFlowVariables(body.Variables)
	}
	if len(body.BusinessTags) > 0 {
		builder.SetBusinessTags(body.BusinessTags)
	}
	if len(body.Viewers) > 0 {
		builder.SetViewers(body.Viewers)
	}
	if len(body.Editors) > 0 {
		builder.SetEditors(body.Editors)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "businessRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("businesses", row); err != nil {
		logger.Errorf(ctx, "businessRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific business record
func (r *businessRepository) Get(ctx context.Context, params *structs.FindBusinessParams) (*ent.Business, error) {
	builder := r.ec.Business.Query()

	if params.ProcessID != "" {
		builder.Where(businessEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.ModuleCode != "" {
		builder.Where(businessEnt.ModuleCodeEQ(params.ModuleCode))
	}
	if params.FormCode != "" {
		builder.Where(businessEnt.FormCodeEQ(params.FormCode))
	}
	if params.Code != "" {
		builder.Where(businessEnt.CodeEQ(params.Code))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(businessEnt.StatusEQ(params.Status))
	}
	if params.FlowStatus != "" {
		builder.Where(businessEnt.FlowStatusEQ(params.FlowStatus))
	}
	if params.IsDraft != nil {
		builder.Where(businessEnt.IsDraftEQ(*params.IsDraft))
	}

	row, err := builder.First(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// Update updates an existing business record
func (r *businessRepository) Update(ctx context.Context, body *structs.UpdateBusinessBody) (*ent.Business, error) {
	builder := r.ec.Business.UpdateOneID(body.ID)

	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.FlowStatus != "" {
		builder.SetFlowStatus(body.FlowStatus)
	}

	if body.CurrentData != nil {
		builder.SetCurrentData(body.CurrentData)
	}
	if body.Variables != nil {
		builder.SetFlowVariables(body.Variables)
	}

	builder.SetIsDraft(body.IsDraft)

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	return builder.Save(ctx)
}

// Delete deletes a business record
func (r *businessRepository) Delete(ctx context.Context, params *structs.FindBusinessParams) error {
	builder := r.ec.Business.Delete()

	if params.ProcessID != "" {
		builder.Where(businessEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.Code != "" {
		builder.Where(businessEnt.CodeEQ(params.Code))
	}

	_, err := builder.Exec(ctx)
	return err
}

// List returns a list of business records
func (r *businessRepository) List(ctx context.Context, params *structs.ListBusinessParams) ([]*ent.Business, error) {
	builder := r.ec.Business.Query()

	if params.ModuleCode != "" {
		builder.Where(businessEnt.ModuleCodeEQ(params.ModuleCode))
	}
	if params.FormCode != "" {
		builder.Where(businessEnt.FormCodeEQ(params.FormCode))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(businessEnt.StatusEQ(params.Status))
	}
	if params.IsDraft != nil {
		builder.Where(businessEnt.IsDraftEQ(*params.IsDraft))
	}

	// Add sorting
	switch params.SortBy {
	default:
		builder.Order(ent.Desc(businessEnt.FieldCreatedAt))
	}

	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// CountX returns the total count of business records
func (r *businessRepository) CountX(ctx context.Context, params *structs.ListBusinessParams) int {
	builder := r.ec.Business.Query()

	if params.ModuleCode != "" {
		builder.Where(businessEnt.ModuleCodeEQ(params.ModuleCode))
	}
	if params.FormCode != "" {
		builder.Where(businessEnt.FormCodeEQ(params.FormCode))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(businessEnt.StatusEQ(params.Status))
	}
	if params.IsDraft != nil {
		builder.Where(businessEnt.IsDraftEQ(*params.IsDraft))
	}

	return builder.CountX(ctx)
}

// UpdateFlowStatus updates the flow status of a business record
func (r *businessRepository) UpdateFlowStatus(ctx context.Context, businessID string, status string) error {
	return r.ec.Business.UpdateOneID(businessID).
		SetFlowStatus(status).
		Exec(ctx)
}

// UpdateBusinessData updates the business data
func (r *businessRepository) UpdateBusinessData(ctx context.Context, businessID string, data types.JSON) error {
	return r.ec.Business.UpdateOneID(businessID).
		SetCurrentData(data).
		Exec(ctx)
}
