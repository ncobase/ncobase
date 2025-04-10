package repository

import (
	"context"
	"ncobase/core/workflow/data"
	"ncobase/core/workflow/data/ent"
	historyEnt "ncobase/core/workflow/data/ent/history"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/pkg/data/cache"
	"github.com/ncobase/ncore/pkg/data/meili"
	"github.com/ncobase/ncore/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type HistoryRepositoryInterface interface {
	Create(context.Context, *structs.HistoryBody) (*ent.History, error)
	Get(context.Context, *structs.FindHistoryParams) (*ent.History, error)
	List(context.Context, *structs.ListHistoryParams) ([]*ent.History, error)
	CountX(context.Context, *structs.ListHistoryParams) int
}

type historyRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.History]
}

func NewHistoryRepository(d *data.Data) HistoryRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &historyRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.History](rc, "workflow_history", false),
	}
}

// Create creates a new history record
func (r *historyRepository) Create(ctx context.Context, body *structs.HistoryBody) (*ent.History, error) {
	builder := r.ec.History.Create()

	if body.Type != "" {
		builder.SetType(body.Type)
	}
	if body.ProcessID != "" {
		builder.SetProcessID(body.ProcessID)
	}
	if body.TemplateID != "" {
		builder.SetTemplateID(body.TemplateID)
	}
	if body.NodeID != "" {
		builder.SetNodeKey(body.NodeID)
	}
	if body.NodeName != "" {
		builder.SetNodeName(body.NodeName)
	}
	if body.NodeType != "" {
		builder.SetNodeType(body.NodeType)
	}
	if body.TaskID != "" {
		builder.SetNodeKey(body.TaskID)
	}
	if body.Operator != "" {
		builder.SetOperator(body.Operator)
	}
	if body.OperatorDept != "" {
		builder.SetOperatorDept(body.OperatorDept)
	}
	if body.Action != "" {
		builder.SetAction(body.Action)
	}
	if body.Comment != "" {
		builder.SetComment(body.Comment)
	}
	if body.Details != nil {
		builder.SetDetails(body.Details)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "historyRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("histories", row); err != nil {
		logger.Errorf(ctx, "historyRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific history record
func (r *historyRepository) Get(ctx context.Context, params *structs.FindHistoryParams) (*ent.History, error) {
	builder := r.ec.History.Query()

	if params.ProcessID != "" {
		builder.Where(historyEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.TemplateID != "" {
		builder.Where(historyEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.NodeID != "" {
		builder.Where(historyEnt.NodeKeyEQ(params.NodeID))
	}
	if params.TaskID != "" {
		builder.Where(historyEnt.NodeKeyEQ(params.TaskID))
	}
	if params.Operator != "" {
		builder.Where(historyEnt.OperatorEQ(params.Operator))
	}
	if params.Action != "" {
		builder.Where(historyEnt.ActionEQ(params.Action))
	}
	if params.Type != "" {
		builder.Where(historyEnt.TypeEQ(params.Type))
	}

	row, err := builder.First(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// List returns a list of history records
func (r *historyRepository) List(ctx context.Context, params *structs.ListHistoryParams) ([]*ent.History, error) {
	builder := r.ec.History.Query()

	if params.ProcessID != "" {
		builder.Where(historyEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.TaskID != "" {
		builder.Where(historyEnt.NodeKeyEQ(params.TaskID))
	}
	if params.Operator != "" {
		builder.Where(historyEnt.OperatorEQ(params.Operator))
	}
	if params.Type != "" {
		builder.Where(historyEnt.TypeEQ(params.Type))
	}

	// Add sorting
	switch params.SortBy {
	default:
		builder.Order(ent.Desc(historyEnt.FieldCreatedAt))
	}

	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// CountX returns the total count of history records
func (r *historyRepository) CountX(ctx context.Context, params *structs.ListHistoryParams) int {
	builder := r.ec.History.Query()

	if params.ProcessID != "" {
		builder.Where(historyEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.TaskID != "" {
		builder.Where(historyEnt.NodeKeyEQ(params.TaskID))
	}
	if params.Operator != "" {
		builder.Where(historyEnt.OperatorEQ(params.Operator))
	}
	if params.Type != "" {
		builder.Where(historyEnt.TypeEQ(params.Type))
	}

	return builder.CountX(ctx)
}
