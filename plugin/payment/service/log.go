package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/plugin/payment/data/repository"
	"ncobase/plugin/payment/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

// LogServiceInterface defines the interface for log service operations
type LogServiceInterface interface {
	GetByID(ctx context.Context, id string) (*structs.Log, error)
	GetByOrderID(ctx context.Context, orderID string, page, pageSize int) (paging.Result[*structs.Log], error)
	List(ctx context.Context, query *structs.LogQuery) (paging.Result[*structs.Log], error)
	Create(ctx context.Context, log *structs.CreateLogInput) (*structs.Log, error)
	Serialize(log *structs.Log) *structs.Log
	Serializes(logs []*structs.Log) []*structs.Log
}

// logService provides operations for payment logs
type logService struct {
	repo repository.LogRepositoryInterface
}

// NewLogService creates a new log service
func NewLogService(repo repository.LogRepositoryInterface) LogServiceInterface {
	return &logService{
		repo: repo,
	}
}

// GetByID gets a payment log by ID
func (s *logService) GetByID(ctx context.Context, id string) (*structs.Log, error) {
	if id == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	log, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, handleEntError(ctx, "Log", err)
	}

	return s.Serialize(log), nil
}

// GetByOrderID gets payment logs for an order
func (s *logService) GetByOrderID(ctx context.Context, orderID string, page, pageSize int) (paging.Result[*structs.Log], error) {
	if orderID == "" {
		return paging.Result[*structs.Log]{}, errors.New(ecode.FieldIsRequired("order_id"))
	}

	// Create a log query
	query := &structs.LogQuery{
		OrderID: orderID,
		PaginationQuery: structs.PaginationQuery{
			PageSize: pageSize,
		},
	}

	return s.List(ctx, query)
}

// List lists payment logs
func (s *logService) List(ctx context.Context, query *structs.LogQuery) (paging.Result[*structs.Log], error) {
	pp := paging.Params{
		Cursor:    query.Cursor,
		Limit:     query.PageSize,
		Direction: "forward", // Default direction
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.Log, int, error) {
		lq := *query
		lq.Cursor = cursor
		lq.PageSize = limit

		logs, err := s.repo.List(ctx, &lq)
		if err != nil {
			logger.Errorf(ctx, "Error listing logs: %v", err)
			return nil, 0, err
		}

		total, err := s.repo.Count(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Error counting logs: %v", err)
			return nil, 0, err
		}

		return s.Serializes(logs), int(total), nil
	})
}

// Create creates a new payment log
func (s *logService) Create(ctx context.Context, log *structs.CreateLogInput) (*structs.Log, error) {
	if log.ChannelID == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	if log.Type == "" {
		return nil, fmt.Errorf("log type is required")
	}

	return s.repo.Create(ctx, log)
}

// Serialize serializes a log entity to a response format
func (s *logService) Serialize(log *structs.Log) *structs.Log {
	if log == nil {
		return nil
	}

	return &structs.Log{
		ID:           log.ID,
		OrderID:      log.OrderID,
		ChannelID:    log.ChannelID,
		Type:         log.Type,
		StatusBefore: log.StatusBefore,
		StatusAfter:  log.StatusAfter,
		RequestData:  log.RequestData,
		ResponseData: log.ResponseData,
		IP:           log.IP,
		UserAgent:    log.UserAgent,
		UserID:       log.UserID,
		Error:        log.Error,
		Metadata:     log.Metadata,
		CreatedAt:    log.CreatedAt,
		UpdatedAt:    log.UpdatedAt,
	}
}

// Serializes serializes multiple log entities to response format
func (s *logService) Serializes(logs []*structs.Log) []*structs.Log {
	result := make([]*structs.Log, len(logs))
	for i, log := range logs {
		result[i] = s.Serialize(log)
	}
	return result
}
