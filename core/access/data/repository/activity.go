package repository

import (
	"context"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	activityEnt "ncobase/access/data/ent/activity"
	"ncobase/access/structs"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
)

// ActivityRepositoryInterface defines repository operations for activity logs
type ActivityRepositoryInterface interface {
	Create(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*ent.Activity, error)
	GetByID(ctx context.Context, id string) (*ent.Activity, error)
	List(ctx context.Context, params *structs.ListActivityParams) ([]*ent.Activity, int, error)
	GetRecentByUserID(ctx context.Context, userID string, limit int) ([]*ent.Activity, error)
}

// activityRepository implements ActivityRepositoryInterface
type activityRepository struct {
	ec *ent.Client
}

// NewActivityRepository creates a new activity log repository
func NewActivityRepository(d *data.Data) ActivityRepositoryInterface {
	ec := d.GetMasterEntClient()
	return &activityRepository{ec}
}

// Create adds a new activity log
func (r *activityRepository) Create(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*ent.Activity, error) {
	builder := r.ec.Activity.Create()

	builder.SetID(nanoid.PrimaryKey()())
	builder.SetUserID(userID)
	builder.SetType(string(log.Type))
	builder.SetDetails(log.Details)

	if log.Metadata != nil {
		builder.SetMetadata(*log.Metadata)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "activityRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID retrieves an activity log by ID
func (r *activityRepository) GetByID(ctx context.Context, id string) (*ent.Activity, error) {
	row, err := r.ec.Activity.Get(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "activityRepo.GetByID error: %v", err)
		return nil, err
	}

	return row, nil
}

// List retrieves activity logs with pagination and filters
func (r *activityRepository) List(ctx context.Context, params *structs.ListActivityParams) ([]*ent.Activity, int, error) {
	builder := r.ec.Activity.Query()

	// Apply filters
	if params.UserID != "" {
		builder = builder.Where(activityEnt.UserIDEQ(params.UserID))
	}

	if params.Type != "" {
		builder = builder.Where(activityEnt.TypeEQ(params.Type))
	}

	if params.FromDate > 0 {
		builder = builder.Where(activityEnt.CreatedAtGTE(params.FromDate))
	}

	if params.ToDate > 0 {
		builder = builder.Where(activityEnt.CreatedAtLTE(params.ToDate))
	}

	// Get total count
	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "activityRepo.List count error: %v", err)
		return nil, 0, err
	}

	// Order by created_at desc
	builder = builder.Order(ent.Desc(activityEnt.FieldCreatedAt))

	// Apply pagination
	if params.Limit > 0 {
		builder = builder.Limit(params.Limit)
	} else {
		builder = builder.Limit(20) // Default limit
	}

	if params.Offset > 0 {
		builder = builder.Offset(params.Offset)
	}

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "activityRepo.List error: %v", err)
		return nil, 0, err
	}

	return rows, total, nil
}

// GetRecentByUserID retrieves recent activity logs for a user
func (r *activityRepository) GetRecentByUserID(ctx context.Context, userID string, limit int) ([]*ent.Activity, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	rows, err := r.ec.Activity.
		Query().
		Where(activityEnt.UserIDEQ(userID)).
		Order(ent.Desc(activityEnt.FieldCreatedAt)).
		Limit(limit).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "activityRepo.GetRecentByUserID error: %v", err)
		return nil, err
	}

	return rows, nil
}
