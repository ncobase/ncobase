package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	activityEnt "ncobase/access/data/ent/activity"
	"ncobase/access/structs"
	"strconv"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
)

// ActivityRepositoryInterface defines repository operations for activitys
type ActivityRepositoryInterface interface {
	Create(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*structs.ActivityDocument, error)
	GetByID(ctx context.Context, id string) (*structs.ActivityDocument, error)
	List(ctx context.Context, params *structs.ListActivityParams) (paging.Result[*structs.ActivityDocument], error)
	GetRecentByUserID(ctx context.Context, userID string, limit int) ([]*structs.ActivityDocument, error)
	Search(ctx context.Context, params *structs.SearchActivityParams) ([]*structs.ActivityDocument, int, error)
	CountX(ctx context.Context, params *structs.ListActivityParams) int
}

// activityRepository implements ActivityRepositoryInterface
type activityRepository struct {
	data        *data.Data
	searchIndex string
}

// NewActivityRepository creates a new activity repository
func NewActivityRepository(d *data.Data) ActivityRepositoryInterface {
	return &activityRepository{
		data:        d,
		searchIndex: "activities",
	}
}

// Create creates a new activity
func (r *activityRepository) Create(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*structs.ActivityDocument, error) {
	id := nanoid.PrimaryKey()()
	now := time.Now().UnixMilli()

	doc := &structs.ActivityDocument{
		ID:        id,
		UserID:    userID,
		Type:      log.Type,
		Details:   log.Details,
		Metadata:  log.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Check if search engine is available
	searchEngines := r.data.GetAvailableSearchEngines()
	searchAvailable := len(searchEngines) > 0

	if searchAvailable {
		// Primary storage: Search engine
		if err := r.indexToSearch(ctx, doc); err != nil {
			logger.Errorf(ctx, "Failed to index to search engine: %v", err)
			// Fallback to database if search fails
			if err := r.storeInDatabase(ctx, doc); err != nil {
				logger.Errorf(ctx, "Failed to store in database fallback: %v", err)
				return nil, err
			}
		} else {
			// Async backup to database for data durability
			go func() {
				backgroundCtx := context.Background()
				if err := r.storeInDatabase(backgroundCtx, doc); err != nil {
					logger.Warnf(backgroundCtx, "Failed to backup to database: %v", err)
				}
			}()
		}
	} else {
		// Primary storage: Database (search not available)
		if err := r.storeInDatabase(ctx, doc); err != nil {
			logger.Errorf(ctx, "Failed to store in database: %v", err)
			return nil, err
		}
	}

	return doc, nil
}

// GetByID retrieves an activity by ID
func (r *activityRepository) GetByID(ctx context.Context, id string) (*structs.ActivityDocument, error) {
	searchEngines := r.data.GetAvailableSearchEngines()

	if len(searchEngines) > 0 {
		if doc, err := r.getFromSearch(ctx, id); err == nil && doc != nil {
			return doc, nil
		}
		logger.Debugf(ctx, "Search engine lookup failed for ID %s, trying database", id)
	}

	return r.getFromDatabase(ctx, id)
}

// List retrieves a list of activities
func (r *activityRepository) List(ctx context.Context, params *structs.ListActivityParams) (paging.Result[*structs.ActivityDocument], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ActivityDocument, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		searchEngines := r.data.GetAvailableSearchEngines()

		if len(searchEngines) > 0 {
			if docs, total, err := r.listFromSearch(ctx, &lp); err == nil {
				return docs, total, nil
			}
			logger.Debugf(ctx, "Search engine list failed, using database")
		}

		return r.listFromDatabase(ctx, &lp)
	})
}

// GetRecentByUserID retrieves recent activities for a user
func (r *activityRepository) GetRecentByUserID(ctx context.Context, userID string, limit int) ([]*structs.ActivityDocument, error) {
	params := &structs.ListActivityParams{
		UserID:    userID,
		Limit:     limit,
		Direction: "forward",
	}

	result, err := r.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// Search searches for activities
func (r *activityRepository) Search(ctx context.Context, params *structs.SearchActivityParams) ([]*structs.ActivityDocument, int, error) {
	searchEngines := r.data.GetAvailableSearchEngines()
	if len(searchEngines) == 0 {
		logger.Warnf(ctx, "No search engines available, using database fallback")
		return r.searchFallback(ctx, params)
	}

	req := &search.Request{
		Index: r.searchIndex,
		Query: params.Query,
		From:  params.From,
		Size:  params.Size,
	}

	if params.UserID != "" || params.Type != "" || params.FromDate > 0 || params.ToDate > 0 {
		req.Filter = make(map[string]interface{})

		if params.UserID != "" {
			req.Filter["user_id"] = params.UserID
		}
		if params.Type != "" {
			req.Filter["type"] = params.Type
		}
		if params.FromDate > 0 || params.ToDate > 0 {
			dateRange := make(map[string]interface{})
			if params.FromDate > 0 {
				dateRange["gte"] = params.FromDate
			}
			if params.ToDate > 0 {
				dateRange["lte"] = params.ToDate
			}
			req.Filter["created_at"] = map[string]interface{}{"range": dateRange}
		}
	}

	resp, err := r.data.Search(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "Search engine query failed: %v", err)
		return r.searchFallback(ctx, params)
	}

	docs := make([]*structs.ActivityDocument, len(resp.Hits))
	for i, hit := range resp.Hits {
		doc, err := r.convertHitToDocument(hit.Source)
		if err != nil {
			logger.Warnf(ctx, "Failed to convert search hit: %v", err)
			continue
		}
		docs[i] = doc
	}

	return docs, int(resp.Total), nil
}

// CountX counts the number of activities
func (r *activityRepository) CountX(ctx context.Context, params *structs.ListActivityParams) int {
	searchEngines := r.data.GetAvailableSearchEngines()

	if len(searchEngines) > 0 {
		// Try search engine count
		req := &search.Request{
			Index: r.searchIndex,
			Query: "*",
			Size:  0, // Only get count
		}

		if resp, err := r.data.Search(ctx, req); err == nil {
			return int(resp.Total)
		}
	}

	// Fallback to database count
	builder := r.data.GetSlaveEntClient().Activity.Query()
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

	return builder.CountX(ctx)
}

// indexToSearch indexes an activity to the search engine
func (r *activityRepository) indexToSearch(ctx context.Context, doc *structs.ActivityDocument) error {
	indexDoc := map[string]interface{}{
		"id":         doc.ID,
		"user_id":    doc.UserID,
		"type":       doc.Type,
		"details":    doc.Details,
		"created_at": doc.CreatedAt,
		"updated_at": doc.UpdatedAt,
	}

	if doc.Metadata != nil {
		indexDoc["metadata"] = *doc.Metadata
	}

	req := &search.IndexRequest{
		Index:      r.searchIndex,
		DocumentID: doc.ID,
		Document:   indexDoc,
	}

	return r.data.IndexDocument(ctx, req)
}

// storeInDatabase stores an activity in the database
func (r *activityRepository) storeInDatabase(ctx context.Context, doc *structs.ActivityDocument) error {
	ec := r.data.GetMasterEntClient()

	builder := ec.Activity.Create()
	builder.SetID(doc.ID)
	builder.SetUserID(doc.UserID)
	builder.SetType(doc.Type)
	builder.SetDetails(doc.Details)
	builder.SetCreatedAt(doc.CreatedAt)
	builder.SetUpdatedAt(doc.UpdatedAt)

	if doc.Metadata != nil {
		builder.SetMetadata(*doc.Metadata)
	}

	_, err := builder.Save(ctx)
	return err
}

// getFromSearch retrieves an activity from the search engine
func (r *activityRepository) getFromSearch(ctx context.Context, id string) (*structs.ActivityDocument, error) {
	query := fmt.Sprintf(`{"term": {"id": "%s"}}`, id)
	req := &search.Request{
		Index: r.searchIndex,
		Query: query,
		Size:  1,
	}

	resp, err := r.data.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Hits) == 0 {
		return nil, fmt.Errorf("activity not found")
	}

	return r.convertHitToDocument(resp.Hits[0].Source)
}

// getFromDatabase retrieves an activity from the database
func (r *activityRepository) getFromDatabase(ctx context.Context, id string) (*structs.ActivityDocument, error) {
	ec := r.data.GetSlaveEntClient()

	row, err := ec.Activity.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.convertEntToDocument(row), nil
}

// listFromSearch retrieves a list of activities from the search engine
func (r *activityRepository) listFromSearch(ctx context.Context, params *structs.ListActivityParams) ([]*structs.ActivityDocument, int, error) {
	req := &search.Request{
		Index: r.searchIndex,
		Query: "*",
		From:  params.Offset,
		Size:  params.Limit,
	}

	if params.UserID != "" || params.Type != "" || params.FromDate > 0 || params.ToDate > 0 {
		filters := make(map[string]interface{})

		if params.UserID != "" {
			filters["user_id"] = params.UserID
		}
		if params.Type != "" {
			filters["type"] = params.Type
		}
		if params.FromDate > 0 || params.ToDate > 0 {
			dateRange := make(map[string]interface{})
			if params.FromDate > 0 {
				dateRange["gte"] = params.FromDate
			}
			if params.ToDate > 0 {
				dateRange["lte"] = params.ToDate
			}
			filters["created_at"] = map[string]interface{}{"range": dateRange}
		}

		req.Filter = filters
	}

	resp, err := r.data.Search(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	docs := make([]*structs.ActivityDocument, len(resp.Hits))
	for i, hit := range resp.Hits {
		doc, err := r.convertHitToDocument(hit.Source)
		if err != nil {
			logger.Warnf(ctx, "Failed to convert search hit: %v", err)
			continue
		}
		docs[i] = doc
	}

	return docs, int(resp.Total), nil
}

// listFromDatabase retrieves a list of activities from the database
func (r *activityRepository) listFromDatabase(ctx context.Context, params *structs.ListActivityParams) ([]*structs.ActivityDocument, int, error) {
	ec := r.data.GetSlaveEntClient()
	builder := ec.Activity.Query()

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

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, 0, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				activityEnt.Or(
					activityEnt.CreatedAtGT(timestamp),
					activityEnt.And(
						activityEnt.CreatedAtEQ(timestamp),
						activityEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				activityEnt.Or(
					activityEnt.CreatedAtLT(timestamp),
					activityEnt.And(
						activityEnt.CreatedAtEQ(timestamp),
						activityEnt.IDLT(id),
					),
				),
			)
		}
	}

	// Apply ordering
	if params.Direction == "backward" {
		builder.Order(ent.Asc(activityEnt.FieldCreatedAt), ent.Asc(activityEnt.FieldID))
	} else {
		builder.Order(ent.Desc(activityEnt.FieldCreatedAt), ent.Desc(activityEnt.FieldID))
	}

	// Get total count
	total, err := builder.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if params.Limit > 0 {
		builder = builder.Limit(params.Limit)
	}
	if params.Offset > 0 {
		builder = builder.Offset(params.Offset)
	}

	rows, err := builder.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	docs := make([]*structs.ActivityDocument, len(rows))
	for i, row := range rows {
		docs[i] = r.convertEntToDocument(row)
	}

	return docs, total, nil
}

// searchFallback performs full-text search on activities
func (r *activityRepository) searchFallback(ctx context.Context, params *structs.SearchActivityParams) ([]*structs.ActivityDocument, int, error) {
	ec := r.data.GetSlaveEntClient()
	builder := ec.Activity.Query()

	if params.Query != "" {
		builder = builder.Where(activityEnt.DetailsContains(params.Query))
	}

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

	total, err := builder.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	builder = builder.Order(ent.Desc(activityEnt.FieldCreatedAt))

	if params.Size > 0 {
		builder = builder.Limit(params.Size)
	}
	if params.From > 0 {
		builder = builder.Offset(params.From)
	}

	rows, err := builder.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	docs := make([]*structs.ActivityDocument, len(rows))
	for i, row := range rows {
		docs[i] = r.convertEntToDocument(row)
	}

	return docs, total, nil
}

func (r *activityRepository) convertHitToDocument(source map[string]interface{}) (*structs.ActivityDocument, error) {
	doc := &structs.ActivityDocument{}

	if id, ok := source["id"].(string); ok {
		doc.ID = id
	}
	if userID, ok := source["user_id"].(string); ok {
		doc.UserID = userID
	}
	if actType, ok := source["type"].(string); ok {
		doc.Type = actType
	}
	if details, ok := source["details"].(string); ok {
		doc.Details = details
	}

	if createdAt, ok := source["created_at"]; ok {
		switch v := createdAt.(type) {
		case float64:
			doc.CreatedAt = int64(v)
		case string:
			if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
				doc.CreatedAt = ts
			}
		}
	}

	if updatedAt, ok := source["updated_at"]; ok {
		switch v := updatedAt.(type) {
		case float64:
			doc.UpdatedAt = int64(v)
		case string:
			if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
				doc.UpdatedAt = ts
			}
		}
	}

	if metadata, ok := source["metadata"]; ok {
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			var metadataJSON types.JSON
			if err := json.Unmarshal(metadataBytes, &metadataJSON); err == nil {
				doc.Metadata = &metadataJSON
			}
		}
	}

	return doc, nil
}

// convertEntToDocument converts an activity entity to document
func (r *activityRepository) convertEntToDocument(row *ent.Activity) *structs.ActivityDocument {
	return &structs.ActivityDocument{
		ID:        row.ID,
		UserID:    row.UserID,
		Type:      row.Type,
		Details:   row.Details,
		Metadata:  &row.Metadata,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
