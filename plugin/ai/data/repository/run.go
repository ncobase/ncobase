package repository

import (
	"context"
	"fmt"
	"ncobase/plugin/ai/data"
	"ncobase/plugin/ai/data/ent"
	airunEnt "ncobase/plugin/ai/data/ent/airun"
	"ncobase/plugin/ai/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
)

type RunRepositoryInterface interface {
	Create(ctx context.Context, input *structs.CreateRunInput) (*structs.Run, error)
	Update(ctx context.Context, input *structs.UpdateRunInput) (*structs.Run, error)
	GetByID(ctx context.Context, id string) (*structs.Run, error)
	List(ctx context.Context, query *structs.RunQuery) ([]*structs.Run, error)
	Count(ctx context.Context, query *structs.RunQuery) (int64, error)
	Usage(ctx context.Context, query *structs.UsageQuery) (*structs.UsageSummary, error)
}

type runRepository struct {
	ec  *ent.Client
	ecr *ent.Client
}

func NewRunRepository(d *data.Data) RunRepositoryInterface {
	return &runRepository{
		ec:  d.GetMasterEntClient(),
		ecr: d.GetReadEntClient(),
	}
}

func (r *runRepository) Create(ctx context.Context, input *structs.CreateRunInput) (*structs.Run, error) {
	builder := r.ec.AIRun.Create().
		SetMode(string(input.Mode)).
		SetStatus(string(input.Status))

	if input.OperationID != "" {
		builder.SetOperationID(input.OperationID)
	}
	if input.Action != "" {
		builder.SetAction(input.Action)
	}
	if input.Provider != "" {
		builder.SetProvider(input.Provider)
	}
	if input.Model != "" {
		builder.SetModel(input.Model)
	}
	if input.RequestHash != "" {
		builder.SetRequestHash(input.RequestHash)
	}
	if len(input.Metadata) > 0 {
		builder.SetMetadata(input.Metadata)
	}
	if input.SpaceID != "" {
		builder.SetSpaceID(input.SpaceID)
	}
	if input.UserID != "" {
		builder.SetUserID(input.UserID)
	}
	if input.CreatedBy != "" {
		builder.SetCreatedBy(input.CreatedBy)
	}
	if input.UpdatedBy != "" {
		builder.SetUpdatedBy(input.UpdatedBy)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI run: %w", err)
	}
	return SerializeRun(row), nil
}

func (r *runRepository) Update(ctx context.Context, input *structs.UpdateRunInput) (*structs.Run, error) {
	if input.ID == "" {
		return nil, fmt.Errorf("run id is required")
	}

	builder := r.ec.AIRun.UpdateOneID(input.ID)
	if input.OperationID != "" {
		builder.SetOperationID(input.OperationID)
	}
	if input.Action != "" {
		builder.SetAction(input.Action)
	}
	if input.Mode != "" {
		builder.SetMode(string(input.Mode))
	}
	if input.Status != "" {
		builder.SetStatus(string(input.Status))
	}
	if input.Provider != "" {
		builder.SetProvider(input.Provider)
	}
	if input.Model != "" {
		builder.SetModel(input.Model)
	}
	if input.FallbackModel != "" {
		builder.SetFallbackModel(input.FallbackModel)
	}
	builder.SetInputTokens(input.InputTokens)
	builder.SetOutputTokens(input.OutputTokens)
	builder.SetTotalTokens(input.TotalTokens)
	builder.SetReasoningTokens(input.ReasoningTokens)
	builder.SetCacheCreatedTokens(input.CacheCreatedTokens)
	builder.SetCacheReadTokens(input.CacheReadTokens)
	if input.DurationMS > 0 {
		builder.SetDurationMs(input.DurationMS)
	}
	if input.ErrorCode != "" {
		builder.SetErrorCode(input.ErrorCode)
	}
	if input.ErrorMessage != "" {
		builder.SetErrorMessage(input.ErrorMessage)
	}
	if input.RequestHash != "" {
		builder.SetRequestHash(input.RequestHash)
	}
	if input.ResponseHash != "" {
		builder.SetResponseHash(input.ResponseHash)
	}
	if input.EstimatedCost > 0 {
		builder.SetEstimatedCost(input.EstimatedCost)
	}
	if input.Currency != "" {
		builder.SetCurrency(input.Currency)
	}
	if input.Metadata != nil {
		builder.SetMetadata(input.Metadata)
	}
	if input.UpdatedBy != "" {
		builder.SetUpdatedBy(input.UpdatedBy)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update AI run: %w", err)
	}
	return SerializeRun(row), nil
}

func (r *runRepository) GetByID(ctx context.Context, id string) (*structs.Run, error) {
	row, err := r.ecr.AIRun.Query().
		Where(airunEnt.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI run: %w", err)
	}
	return SerializeRun(row), nil
}

func (r *runRepository) List(ctx context.Context, query *structs.RunQuery) ([]*structs.Run, error) {
	q := r.applyRunFilters(r.ecr.AIRun.Query(), query)

	if query.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(query.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}
		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}
		q = q.Where(
			airunEnt.Or(
				airunEnt.CreatedAtLT(timestamp),
				airunEnt.And(
					airunEnt.CreatedAtEQ(timestamp),
					airunEnt.IDLT(id),
				),
			),
		)
	}

	q.Order(ent.Desc(airunEnt.FieldCreatedAt), ent.Desc(airunEnt.FieldID))
	if query.PageSize > 0 {
		q.Limit(query.PageSize)
	} else {
		q.Limit(20)
	}

	rows, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list AI runs: %w", err)
	}
	return SerializeRuns(rows), nil
}

func (r *runRepository) Count(ctx context.Context, query *structs.RunQuery) (int64, error) {
	count, err := r.applyRunFilters(r.ecr.AIRun.Query(), query).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count AI runs: %w", err)
	}
	return int64(count), nil
}

func (r *runRepository) Usage(ctx context.Context, query *structs.UsageQuery) (*structs.UsageSummary, error) {
	summary := &structs.UsageSummary{
		ByProvider: make(map[string]*structs.UsageBucket),
		ByAction:   make(map[string]*structs.UsageBucket),
		ByMode:     make(map[string]*structs.UsageBucket),
	}

	var totals []struct {
		Count              int     `json:"count"`
		InputTokens        int     `json:"input_tokens"`
		OutputTokens       int     `json:"output_tokens"`
		TotalTokens        int     `json:"total_tokens"`
		ReasoningTokens    int     `json:"reasoning_tokens"`
		CacheCreatedTokens int     `json:"cache_created_tokens"`
		CacheReadTokens    int     `json:"cache_read_tokens"`
		DurationMs         int64   `json:"duration_ms"`
		EstimatedCost      float64 `json:"estimated_cost"`
	}
	err := r.applyUsageFilters(r.ecr.AIRun.Query(), query).
		Aggregate(
			ent.Count(),
			ent.As(ent.Sum(airunEnt.FieldInputTokens), "input_tokens"),
			ent.As(ent.Sum(airunEnt.FieldOutputTokens), "output_tokens"),
			ent.As(ent.Sum(airunEnt.FieldTotalTokens), "total_tokens"),
			ent.As(ent.Sum(airunEnt.FieldReasoningTokens), "reasoning_tokens"),
			ent.As(ent.Sum(airunEnt.FieldCacheCreatedTokens), "cache_created_tokens"),
			ent.As(ent.Sum(airunEnt.FieldCacheReadTokens), "cache_read_tokens"),
			ent.As(ent.Sum(airunEnt.FieldDurationMs), "duration_ms"),
			ent.As(ent.Sum(airunEnt.FieldEstimatedCost), "estimated_cost"),
		).
		Scan(ctx, &totals)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate AI usage: %w", err)
	}
	if len(totals) > 0 {
		row := totals[0]
		summary.TotalRuns = int64(row.Count)
		summary.InputTokens = int64(row.InputTokens)
		summary.OutputTokens = int64(row.OutputTokens)
		summary.TotalTokens = int64(row.TotalTokens)
		summary.ReasoningTokens = int64(row.ReasoningTokens)
		summary.CacheCreatedTokens = int64(row.CacheCreatedTokens)
		summary.CacheReadTokens = int64(row.CacheReadTokens)
		summary.DurationMS = row.DurationMs
		summary.EstimatedCost = row.EstimatedCost
	}

	if err := r.fillStatusCounts(ctx, query, summary); err != nil {
		return nil, err
	}
	if err := r.fillUsageBuckets(ctx, query, airunEnt.FieldProvider, summary.ByProvider); err != nil {
		return nil, err
	}
	if err := r.fillUsageBuckets(ctx, query, airunEnt.FieldAction, summary.ByAction); err != nil {
		return nil, err
	}
	if err := r.fillUsageBuckets(ctx, query, airunEnt.FieldMode, summary.ByMode); err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *runRepository) fillStatusCounts(ctx context.Context, query *structs.UsageQuery, summary *structs.UsageSummary) error {
	var rows []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	if err := r.applyUsageFilters(r.ecr.AIRun.Query(), query).
		GroupBy(airunEnt.FieldStatus).
		Aggregate(ent.Count()).
		Scan(ctx, &rows); err != nil {
		return fmt.Errorf("failed to aggregate AI usage by status: %w", err)
	}
	for _, row := range rows {
		switch structs.RunStatus(row.Status) {
		case structs.RunStatusSucceeded:
			summary.SucceededRuns = int64(row.Count)
		case structs.RunStatusFailed:
			summary.FailedRuns = int64(row.Count)
		}
	}
	return nil
}

func (r *runRepository) fillUsageBuckets(ctx context.Context, query *structs.UsageQuery, field string, target map[string]*structs.UsageBucket) error {
	type bucketValues struct {
		Key           string
		Status        string
		Count         int
		TotalTokens   int
		EstimatedCost float64
	}
	values := []bucketValues{}
	aggregate := func(scanTarget any) error {
		return r.applyUsageFilters(r.ecr.AIRun.Query(), query).
			GroupBy(field, airunEnt.FieldStatus).
			Aggregate(
				ent.Count(),
				ent.As(ent.Sum(airunEnt.FieldTotalTokens), "total_tokens"),
				ent.As(ent.Sum(airunEnt.FieldEstimatedCost), "estimated_cost"),
			).
			Scan(ctx, scanTarget)
	}

	switch field {
	case airunEnt.FieldProvider:
		var rows []struct {
			Provider      string  `json:"provider"`
			Status        string  `json:"status"`
			Count         int     `json:"count"`
			TotalTokens   int     `json:"total_tokens"`
			EstimatedCost float64 `json:"estimated_cost"`
		}
		if err := aggregate(&rows); err != nil {
			logger.Errorf(ctx, "Failed to aggregate AI usage bucket %s: %v", field, err)
			return fmt.Errorf("failed to aggregate AI usage by %s: %w", field, err)
		}
		for _, row := range rows {
			values = append(values, bucketValues{Key: row.Provider, Status: row.Status, Count: row.Count, TotalTokens: row.TotalTokens, EstimatedCost: row.EstimatedCost})
		}
	case airunEnt.FieldAction:
		var rows []struct {
			Action        string  `json:"action"`
			Status        string  `json:"status"`
			Count         int     `json:"count"`
			TotalTokens   int     `json:"total_tokens"`
			EstimatedCost float64 `json:"estimated_cost"`
		}
		if err := aggregate(&rows); err != nil {
			logger.Errorf(ctx, "Failed to aggregate AI usage bucket %s: %v", field, err)
			return fmt.Errorf("failed to aggregate AI usage by %s: %w", field, err)
		}
		for _, row := range rows {
			values = append(values, bucketValues{Key: row.Action, Status: row.Status, Count: row.Count, TotalTokens: row.TotalTokens, EstimatedCost: row.EstimatedCost})
		}
	case airunEnt.FieldMode:
		var rows []struct {
			Mode          string  `json:"mode"`
			Status        string  `json:"status"`
			Count         int     `json:"count"`
			TotalTokens   int     `json:"total_tokens"`
			EstimatedCost float64 `json:"estimated_cost"`
		}
		if err := aggregate(&rows); err != nil {
			logger.Errorf(ctx, "Failed to aggregate AI usage bucket %s: %v", field, err)
			return fmt.Errorf("failed to aggregate AI usage by %s: %w", field, err)
		}
		for _, row := range rows {
			values = append(values, bucketValues{Key: row.Mode, Status: row.Status, Count: row.Count, TotalTokens: row.TotalTokens, EstimatedCost: row.EstimatedCost})
		}
	default:
		return fmt.Errorf("unsupported AI usage bucket field: %s", field)
	}

	for _, row := range values {
		key := row.Key
		if key == "" {
			key = "unknown"
		}
		bucket := target[key]
		if bucket == nil {
			bucket = &structs.UsageBucket{}
			target[key] = bucket
		}
		bucket.Runs += int64(row.Count)
		bucket.TotalTokens += int64(row.TotalTokens)
		bucket.EstimatedCost += row.EstimatedCost
		switch structs.RunStatus(row.Status) {
		case structs.RunStatusSucceeded:
			bucket.SucceededRuns += int64(row.Count)
		case structs.RunStatusFailed:
			bucket.FailedRuns += int64(row.Count)
		}
	}
	return nil
}

func (r *runRepository) applyRunFilters(q *ent.AIRunQuery, query *structs.RunQuery) *ent.AIRunQuery {
	if query == nil {
		return q
	}
	if query.ID != "" {
		q = q.Where(airunEnt.IDEQ(query.ID))
	}
	if query.Action != "" {
		q = q.Where(airunEnt.ActionEQ(query.Action))
	}
	if query.Mode != "" {
		q = q.Where(airunEnt.ModeEQ(string(query.Mode)))
	}
	if query.Status != "" {
		q = q.Where(airunEnt.StatusEQ(string(query.Status)))
	}
	if query.Provider != "" {
		q = q.Where(airunEnt.ProviderEQ(query.Provider))
	}
	if query.Model != "" {
		q = q.Where(airunEnt.ModelEQ(query.Model))
	}
	if query.UserID != "" {
		q = q.Where(airunEnt.UserIDEQ(query.UserID))
	}
	if query.SpaceID != "" {
		q = q.Where(airunEnt.SpaceIDEQ(query.SpaceID))
	}
	if query.StartDate > 0 {
		q = q.Where(airunEnt.CreatedAtGTE(query.StartDate))
	}
	if query.EndDate > 0 {
		q = q.Where(airunEnt.CreatedAtLTE(query.EndDate))
	}
	return q
}

func (r *runRepository) applyUsageFilters(q *ent.AIRunQuery, query *structs.UsageQuery) *ent.AIRunQuery {
	if query == nil {
		return q
	}
	if query.Action != "" {
		q = q.Where(airunEnt.ActionEQ(query.Action))
	}
	if query.Mode != "" {
		q = q.Where(airunEnt.ModeEQ(query.Mode))
	}
	if query.Status != "" {
		q = q.Where(airunEnt.StatusEQ(query.Status))
	}
	if query.Provider != "" {
		q = q.Where(airunEnt.ProviderEQ(query.Provider))
	}
	if query.Model != "" {
		q = q.Where(airunEnt.ModelEQ(query.Model))
	}
	if query.UserID != "" {
		q = q.Where(airunEnt.UserIDEQ(query.UserID))
	}
	if query.SpaceID != "" {
		q = q.Where(airunEnt.SpaceIDEQ(query.SpaceID))
	}
	if query.StartDate > 0 {
		q = q.Where(airunEnt.CreatedAtGTE(query.StartDate))
	}
	if query.EndDate > 0 {
		q = q.Where(airunEnt.CreatedAtLTE(query.EndDate))
	}
	return q
}
