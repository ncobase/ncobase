package repository

import (
	"context"
	"ncobase/core/workflow/data"
	"ncobase/core/workflow/data/ent"
	processDesignEnt "ncobase/core/workflow/data/ent/processdesign"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/pkg/data/cache"
	"github.com/ncobase/ncore/pkg/data/meili"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/types"

	"github.com/redis/go-redis/v9"
)

type ProcessDesignRepositoryInterface interface {
	Create(context.Context, *structs.ProcessDesignBody) (*ent.ProcessDesign, error)
	Get(context.Context, *structs.FindProcessDesignParams) (*ent.ProcessDesign, error)
	Update(context.Context, *structs.UpdateProcessDesignBody) (*ent.ProcessDesign, error)
	List(context.Context, *structs.ListProcessDesignParams) ([]*ent.ProcessDesign, error)
	Delete(context.Context, string) error

	SaveDraft(ctx context.Context, id string, data types.JSON) error
	PublishDraft(ctx context.Context, id string) error
}

type processDesignRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.ProcessDesign]
}

func NewProcessDesignRepository(d *data.Data) ProcessDesignRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &processDesignRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.ProcessDesign](rc, "workflow_process_design", false),
	}
}

// Create creates a new process design
func (r *processDesignRepository) Create(ctx context.Context, body *structs.ProcessDesignBody) (*ent.ProcessDesign, error) {
	builder := r.ec.ProcessDesign.Create()

	if body.TemplateID != "" {
		builder.SetTemplateID(body.TemplateID)
	}
	if body.GraphData != nil {
		builder.SetGraphData(body.GraphData)
	}
	if body.NodeLayouts != nil {
		builder.SetNodeLayouts(body.NodeLayouts)
	}
	if body.Properties != nil {
		builder.SetProperties(body.Properties)
	}
	if body.ValidationRules != nil {
		builder.SetValidationRules(body.ValidationRules)
	}

	builder.SetIsDraft(body.IsDraft)

	if body.Version != "" {
		builder.SetVersion(body.Version)
	}
	if body.SourceVersion != "" {
		builder.SetSourceVersion(body.SourceVersion)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "processDesignRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("process_designs", row); err != nil {
		logger.Errorf(ctx, "processDesignRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific process design
func (r *processDesignRepository) Get(ctx context.Context, params *structs.FindProcessDesignParams) (*ent.ProcessDesign, error) {
	builder := r.ec.ProcessDesign.Query()

	if params.TemplateID != "" {
		builder.Where(processDesignEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.Version != "" {
		builder.Where(processDesignEnt.VersionEQ(params.Version))
	}
	if params.IsDraft != nil {
		builder.Where(processDesignEnt.IsDraftEQ(*params.IsDraft))
	}

	row, err := builder.First(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// Update updates an existing process design
func (r *processDesignRepository) Update(ctx context.Context, body *structs.UpdateProcessDesignBody) (*ent.ProcessDesign, error) {
	builder := r.ec.ProcessDesign.UpdateOneID(body.ID)

	if body.GraphData != nil {
		builder.SetGraphData(body.GraphData)
	}
	if body.NodeLayouts != nil {
		builder.SetNodeLayouts(body.NodeLayouts)
	}
	if body.Properties != nil {
		builder.SetProperties(body.Properties)
	}
	if body.ValidationRules != nil {
		builder.SetValidationRules(body.ValidationRules)
	}

	builder.SetIsDraft(body.IsDraft)

	if body.Version != "" {
		builder.SetVersion(body.Version)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	return builder.Save(ctx)
}

// List returns a list of process designs
func (r *processDesignRepository) List(ctx context.Context, params *structs.ListProcessDesignParams) ([]*ent.ProcessDesign, error) {
	builder := r.ec.ProcessDesign.Query()

	if params.TemplateID != "" {
		builder.Where(processDesignEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.IsDraft != nil {
		builder.Where(processDesignEnt.IsDraftEQ(*params.IsDraft))
	}

	builder.Order(ent.Desc(processDesignEnt.FieldCreatedAt))
	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// Delete deletes a process design
func (r *processDesignRepository) Delete(ctx context.Context, id string) error {
	return r.ec.ProcessDesign.DeleteOneID(id).Exec(ctx)
}

// SaveDraft saves process design as draft
func (r *processDesignRepository) SaveDraft(ctx context.Context, id string, data types.JSON) error {
	return r.ec.ProcessDesign.UpdateOneID(id).
		SetIsDraft(true).
		SetGraphData(data).
		Exec(ctx)
}

// PublishDraft publishes a draft process design
func (r *processDesignRepository) PublishDraft(ctx context.Context, id string) error {
	return r.ec.ProcessDesign.UpdateOneID(id).
		SetIsDraft(false).
		Exec(ctx)
}
