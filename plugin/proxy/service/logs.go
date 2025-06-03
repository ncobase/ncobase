package service

import (
	"context"
	"errors"
	"ncobase/proxy/data"
	"ncobase/proxy/data/ent"
	"ncobase/proxy/data/repository"
	"ncobase/proxy/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

// LogServiceInterface is the interface for the log service.
type LogServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateLogBody) (*structs.ReadLog, error)
	GetByID(ctx context.Context, id string) (*structs.ReadLog, error)
	List(ctx context.Context, params *structs.ListLogParams) (paging.Result[*structs.ReadLog], error)
	DeleteOlderThan(ctx context.Context, days int) (int, error)
	Serialize(row *ent.Logs) *structs.ReadLog
	Serializes(rows []*ent.Logs) []*structs.ReadLog
}

// logService is the struct for the log service.
type logService struct {
	log repository.LogRepositoryInterface
}

// NewLogService creates a new log service.
func NewLogService(d *data.Data) LogServiceInterface {
	return &logService{
		log: repository.NewLogRepository(d),
	}
}

// Create creates a new proxy log.
func (s *logService) Create(ctx context.Context, body *structs.CreateLogBody) (*structs.ReadLog, error) {
	// Only log request/response bodies if they're not too large
	// This prevents storing extremely large payloads
	const maxBodySize = 1024 * 1024 // 1MB

	if len(body.RequestBody) > maxBodySize {
		logger.Warnf(ctx, "Request body too large for logging (%d bytes), truncating", len(body.RequestBody))
		body.RequestBody = body.RequestBody[:maxBodySize] + "... [truncated]"
	}

	if len(body.ResponseBody) > maxBodySize {
		logger.Warnf(ctx, "Response body too large for logging (%d bytes), truncating", len(body.ResponseBody))
		body.ResponseBody = body.ResponseBody[:maxBodySize] + "... [truncated]"
	}

	// Remove sensitive headers
	if body.RequestHeaders != nil {
		body.RequestHeaders.Del("Authorization")
		body.RequestHeaders.Del("Cookie")
		body.RequestHeaders.Del("Set-Cookie")
	}

	if body.ResponseHeaders != nil {
		body.ResponseHeaders.Del("Authorization")
		body.ResponseHeaders.Del("Cookie")
		body.ResponseHeaders.Del("Set-Cookie")
	}

	row, err := s.log.Create(ctx, body)
	if err := handleEntError(ctx, "Log", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByID retrieves a proxy log by ID.
func (s *logService) GetByID(ctx context.Context, id string) (*structs.ReadLog, error) {
	row, err := s.log.GetByID(ctx, id)
	if err := handleEntError(ctx, "Log", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// List lists all proxy logs.
func (s *logService) List(ctx context.Context, params *structs.ListLogParams) (paging.Result[*structs.ReadLog], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadLog, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.log.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing proxy logs: %v", err)
			return nil, 0, err
		}

		total := s.log.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// DeleteOlderThan deletes logs older than a specified number of days.
func (s *logService) DeleteOlderThan(ctx context.Context, days int) (int, error) {
	return s.log.DeleteOlderThan(ctx, days)
}

// Serializes serializes a list of proxy log entities to a response format.
func (s *logService) Serializes(rows []*ent.Logs) []*structs.ReadLog {
	rs := make([]*structs.ReadLog, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a proxy log entity to a response format.
func (s *logService) Serialize(row *ent.Logs) *structs.ReadLog {
	return &structs.ReadLog{
		ID:              row.ID,
		EndpointID:      row.EndpointID,
		RouteID:         row.RouteID,
		RequestMethod:   row.RequestMethod,
		RequestPath:     row.RequestPath,
		RequestHeaders:  nil, // Convert from DB format as needed
		RequestBody:     row.RequestBody,
		StatusCode:      row.StatusCode,
		ResponseHeaders: nil, // Convert from DB format as needed
		ResponseBody:    row.ResponseBody,
		Duration:        row.Duration,
		Error:           row.Error,
		ClientIP:        row.ClientIP,
		UserID:          row.UserID,
		CreatedAt:       &row.CreatedAt,
	}
}
