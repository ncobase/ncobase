package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/proxy/data"
	"ncobase/proxy/data/ent"
	logEnt "ncobase/proxy/data/ent/logs"
	"ncobase/proxy/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"entgo.io/ent/dialect/sql"
)

// LogRepositoryInterface is the interface for the log repository.
type LogRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateLogBody) (*ent.Logs, error)
	GetByID(ctx context.Context, id string) (*ent.Logs, error)
	List(ctx context.Context, params *structs.ListLogParams) ([]*ent.Logs, error)
	CountX(ctx context.Context, params *structs.ListLogParams) int
	DeleteOlderThan(ctx context.Context, days int) (int, error)
}

// logRepository implements the LogRepositoryInterface.
type logRepository struct {
	ec *ent.Client
}

// NewLogRepository creates a new log repository.
func NewLogRepository(d *data.Data) LogRepositoryInterface {
	ec := d.GetMasterEntClient()
	return &logRepository{ec}
}

// Create creates a new proxy log.
func (r *logRepository) Create(ctx context.Context, body *structs.CreateLogBody) (*ent.Logs, error) {
	// Convert headers to string format for storage
	var requestHeadersStr, responseHeadersStr string

	if body.RequestHeaders != nil {
		requestHeadersBytes, _ := json.Marshal(body.RequestHeaders)
		requestHeadersStr = string(requestHeadersBytes)
	}

	if body.ResponseHeaders != nil {
		responseHeadersBytes, _ := json.Marshal(body.ResponseHeaders)
		responseHeadersStr = string(responseHeadersBytes)
	}

	// Create builder
	builder := r.ec.Logs.Create()

	// Set values
	builder.SetEndpointID(body.EndpointID)
	builder.SetRouteID(body.RouteID)
	builder.SetRequestMethod(body.RequestMethod)
	builder.SetRequestPath(body.RequestPath)
	builder.SetRequestHeaders(requestHeadersStr)
	builder.SetRequestBody(body.RequestBody)
	builder.SetStatusCode(body.StatusCode)
	builder.SetResponseHeaders(responseHeadersStr)
	builder.SetResponseBody(body.ResponseBody)
	builder.SetDuration(body.Duration)

	if body.Error != "" {
		builder.SetError(body.Error)
	}

	if body.ClientIP != "" {
		builder.SetClientIP(body.ClientIP)
	}

	if body.UserID != "" {
		builder.SetUserID(body.UserID)
	}

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "logRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a proxy log by ID.
func (r *logRepository) GetByID(ctx context.Context, id string) (*ent.Logs, error) {
	// Create builder
	builder := r.ec.Logs.Query()

	// Set conditions
	builder = builder.Where(logEnt.IDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "logRepo.GetByID error: %v", err)
		return nil, err
	}

	return row, nil
}

// List lists proxy logs with pagination.
func (r *logRepository) List(ctx context.Context, params *structs.ListLogParams) ([]*ent.Logs, error) {
	// Create builder for list
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, err
	}

	// Apply cursor pagination
	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				logEnt.Or(
					logEnt.CreatedAtGT(timestamp),
					logEnt.And(
						logEnt.CreatedAtEQ(timestamp),
						logEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				logEnt.Or(
					logEnt.CreatedAtLT(timestamp),
					logEnt.And(
						logEnt.CreatedAtEQ(timestamp),
						logEnt.IDLT(id),
					),
				),
			)
		}
	}

	// Set order
	if params.Direction == "backward" {
		builder.Order(ent.Asc(logEnt.FieldCreatedAt), ent.Asc(logEnt.FieldID))
	} else {
		builder.Order(ent.Desc(logEnt.FieldCreatedAt), ent.Desc(logEnt.FieldID))
	}

	// Set limit
	builder.Limit(params.Limit)

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "logRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// CountX counts proxy logs.
func (r *logRepository) CountX(ctx context.Context, params *structs.ListLogParams) int {
	// Create builder for count
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// DeleteOlderThan deletes logs older than a specified number of days.
func (r *logRepository) DeleteOlderThan(ctx context.Context, days int) (int, error) {
	cutoffTime := time.Now().AddDate(0, 0, -days)

	// Create delete builder
	result, err := r.ec.Logs.Delete().
		Where(logEnt.CreatedAtLT(cutoffTime.Unix())).
		Exec(ctx)

	if err != nil {
		logger.Errorf(ctx, "logRepo.DeleteOlderThan error: %v", err)
		return 0, err
	}

	return result, nil
}

// listBuilder creates a builder for listing proxy logs.
func (r *logRepository) listBuilder(ctx context.Context, params *structs.ListLogParams) (*ent.LogsQuery, error) {
	// Create builder
	builder := r.ec.Logs.Query()

	// Apply filters
	if params.EndpointID != "" {
		builder = builder.Where(logEnt.EndpointIDEQ(params.EndpointID))
	}

	if params.RouteID != "" {
		builder = builder.Where(logEnt.RouteIDEQ(params.RouteID))
	}

	if params.RequestMethod != "" {
		builder = builder.Where(logEnt.RequestMethodEQ(params.RequestMethod))
	}

	if params.StatusCode != nil {
		builder = builder.Where(logEnt.StatusCodeEQ(*params.StatusCode))
	}

	if params.Error != nil && *params.Error {
		// Filter for logs with errors (non-empty error field)
		builder = builder.Where(func(s *sql.Selector) {
			s.Where(sql.And(
				sql.NotNull(logEnt.FieldError),
				sql.NEQ(logEnt.FieldError, ""),
			))
		})
	}

	if params.UserID != "" {
		builder = builder.Where(logEnt.UserIDEQ(params.UserID))
	}

	// Time range filters
	if params.FromTime != nil {
		builder = builder.Where(logEnt.CreatedAtGTE(*params.FromTime))
	}

	if params.ToTime != nil {
		builder = builder.Where(logEnt.CreatedAtLTE(*params.ToTime))
	}

	return builder, nil
}
