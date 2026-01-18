package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/core/access/data"
	"ncobase/core/access/data/ent"
	activityEnt "ncobase/core/access/data/ent/activity"
	"ncobase/core/access/structs"
	"strconv"
	"time"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"

	"github.com/redis/go-redis/v9"
)

// ActivityRepositoryInterface defines repository operations for activities
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
	data          *data.Data
	searchClient  *search.Client
	activityCache cache.ICache[structs.ActivityDocument]
	userActCache  cache.ICache[[]structs.ActivityDocument] // Cache recent activities per user
	activityTTL   time.Duration
}

// NewActivityRepository creates a new activity repository
func NewActivityRepository(d *data.Data) ActivityRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &activityRepository{
		data:          d,
		activityCache: cache.NewCache[structs.ActivityDocument](redisClient, "ncse_activities"),
		userActCache:  cache.NewCache[[]structs.ActivityDocument](redisClient, "ncse_user_activities"),
		activityTTL:   time.Hour * 1, // 1 hour cache TTL
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
	var searchAvailable bool
	if r.searchClient != nil {
		searchEngines := r.searchClient.GetAvailableEngines()
		searchAvailable = len(searchEngines) > 0
	}

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

	// Cache the activity and invalidate user activities cache
	go func() {
		r.cacheActivity(context.Background(), doc)
		r.invalidateUserActivitiesCache(context.Background(), userID)
	}()

	return doc, nil
}

// GetByID retrieves an activity by ID
func (r *activityRepository) GetByID(ctx context.Context, id string) (*structs.ActivityDocument, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.activityCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	var doc *structs.ActivityDocument
	var err error

	if r.searchClient != nil {
		searchEngines := r.searchClient.GetAvailableEngines()
		if len(searchEngines) == 0 {
			goto databaseFallback
		}
		if doc, err = r.getFromSearch(ctx, id); err == nil && doc != nil {
			// Cache for future use
			go r.cacheActivity(context.Background(), doc)
			return doc, nil
		}
		logger.Debugf(ctx, "Search engine lookup failed for ID %s, trying database", id)
	}

databaseFallback:
	doc, err = r.getFromDatabase(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go r.cacheActivity(context.Background(), doc)

	return doc, nil
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

		if r.searchClient != nil {
			searchEngines := r.searchClient.GetAvailableEngines()
			if len(searchEngines) == 0 {
				goto listDatabaseFallback
			}
			if docs, total, err := r.listFromSearch(ctx, &lp); err == nil {
				// Cache activities in background
				go func() {
					for _, doc := range docs {
						r.cacheActivity(context.Background(), doc)
					}
				}()
				return docs, total, nil
			}
			logger.Debugf(ctx, "Search engine list failed, using database")
		}

	listDatabaseFallback:
		docs, total, err := r.listFromDatabase(ctx, &lp)
		if err != nil {
			return nil, 0, err
		}

		// Cache activities in background
		go func() {
			for _, doc := range docs {
				r.cacheActivity(context.Background(), doc)
			}
		}()

		return docs, total, nil
	})
}

// GetRecentByUserID retrieves recent activities for a user
func (r *activityRepository) GetRecentByUserID(ctx context.Context, userID string, limit int) ([]*structs.ActivityDocument, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user:%s:recent:%d", userID, limit)
	var cachedActivities []*structs.ActivityDocument
	if err := r.userActCache.GetArray(ctx, cacheKey, &cachedActivities); err == nil && len(cachedActivities) > 0 {
		return cachedActivities, nil
	}

	params := &structs.ListActivityParams{
		UserID:    userID,
		Limit:     limit,
		Direction: "forward",
	}

	result, err := r.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go func() {
		if err := r.userActCache.SetArray(context.Background(), cacheKey, result.Items, r.activityTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user activities %s: %v", userID, err)
		}
	}()

	return result.Items, nil
}

// Search searches for activities
func (r *activityRepository) Search(ctx context.Context, params *structs.SearchActivityParams) ([]*structs.ActivityDocument, int, error) {
	if r.searchClient == nil {
		logger.Warnf(ctx, "Search client not initialized, using database fallback")
		return r.searchFallback(ctx, params)
	}

	searchEngines := r.searchClient.GetAvailableEngines()
	if len(searchEngines) == 0 {
		logger.Warnf(ctx, "No search engines available, using database fallback")
		return r.searchFallback(ctx, params)
	}

	req := &search.Request{
		Index: "activities",
		Query: params.Query,
		From:  params.From,
		Size:  params.Size,
	}

	if params.UserID != "" || params.Type != "" || params.FromDate > 0 || params.ToDate > 0 {
		req.Filter = make(map[string]any)

		if params.UserID != "" {
			req.Filter["user_id"] = params.UserID
		}
		if params.Type != "" {
			req.Filter["type"] = params.Type
		}
		if params.FromDate > 0 || params.ToDate > 0 {
			dateRange := make(map[string]any)
			if params.FromDate > 0 {
				dateRange["gte"] = params.FromDate
			}
			if params.ToDate > 0 {
				dateRange["lte"] = params.ToDate
			}
			req.Filter["created_at"] = map[string]any{"range": dateRange}
		}
	}

	resp, err := r.searchClient.Search(ctx, req)
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

		// Cache activity in background
		go r.cacheActivity(context.Background(), doc)
	}

	return docs, int(resp.Total), nil
}

// CountX counts the number of activities
func (r *activityRepository) CountX(ctx context.Context, params *structs.ListActivityParams) int {
	if r.searchClient != nil {
		searchEngines := r.searchClient.GetAvailableEngines()
		if len(searchEngines) == 0 {
			goto countDatabaseFallback
		}
		// Try search engine count
		req := &search.Request{
			Index: "activities",
			Query: "*",
			Size:  0, // Only get count
		}

		if resp, err := r.searchClient.Search(ctx, req); err == nil {
			return int(resp.Total)
		}
	}

countDatabaseFallback:
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
	indexDoc := map[string]any{
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
		Index:      "activities",
		DocumentID: doc.ID,
		Document:   indexDoc,
	}

	return r.searchClient.Index(ctx, req)
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
		Index: "activities",
		Query: query,
		Size:  1,
	}

	resp, err := r.searchClient.Search(ctx, req)
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
		Index: "activities",
		Query: "*",
		From:  params.Offset,
		Size:  params.Limit,
	}

	if params.UserID != "" || params.Type != "" || params.FromDate > 0 || params.ToDate > 0 {
		filters := make(map[string]any)

		if params.UserID != "" {
			filters["user_id"] = params.UserID
		}
		if params.Type != "" {
			filters["type"] = params.Type
		}
		if params.FromDate > 0 || params.ToDate > 0 {
			dateRange := make(map[string]any)
			if params.FromDate > 0 {
				dateRange["gte"] = params.FromDate
			}
			if params.ToDate > 0 {
				dateRange["lte"] = params.ToDate
			}
			filters["created_at"] = map[string]any{"range": dateRange}
		}

		req.Filter = filters
	}

	resp, err := r.searchClient.Search(ctx, req)
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

	docs := make([]*structs.ActivityDocument, 0, len(rows))
	for _, row := range rows {
		docs = append(docs, r.convertEntToDocument(row))
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

	docs := make([]*structs.ActivityDocument, 0, len(rows))
	for _, row := range rows {
		docs = append(docs, r.convertEntToDocument(row))
	}

	return docs, total, nil
}

func (r *activityRepository) convertHitToDocument(source map[string]any) (*structs.ActivityDocument, error) {
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

// cacheActivity caches an activity
func (r *activityRepository) cacheActivity(ctx context.Context, doc *structs.ActivityDocument) {
	cacheKey := fmt.Sprintf("id:%s", doc.ID)
	if err := r.activityCache.Set(ctx, cacheKey, doc, r.activityTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache activity %s: %v", doc.ID, err)
	}
}

// invalidateUserActivitiesCache invalidates the cache for user activities
func (r *activityRepository) invalidateUserActivitiesCache(ctx context.Context, userID string) {
	// Clear all cached user activities for different limits
	for _, limit := range []int{5, 10, 20, 50} {
		cacheKey := fmt.Sprintf("user:%s:recent:%d", userID, limit)
		if err := r.userActCache.Delete(ctx, cacheKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate user activities cache %s: %v", userID, err)
		}
	}
}
