package repository

import (
	"context"
	"fmt"
	"ncobase/payment/data"
	"ncobase/payment/data/ent"
	paymentLogEnt "ncobase/payment/data/ent/paymentlog"
	"ncobase/payment/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// LogRepositoryInterface defines the interface for log repository operations
type LogRepositoryInterface interface {
	Create(ctx context.Context, log *structs.CreateLogInput) (*structs.Log, error)
	GetByID(ctx context.Context, id string) (*structs.Log, error)
	List(ctx context.Context, query *structs.LogQuery) ([]*structs.Log, error)
	Count(ctx context.Context, query *structs.LogQuery) (int64, error)
}

// logRepository handles payment log persistence
type logRepository struct {
	data *data.Data
}

// NewLogRepository creates a new log repository
func NewLogRepository(d *data.Data) LogRepositoryInterface {
	return &logRepository{data: d}
}

// Create creates a new payment log
func (r *logRepository) Create(ctx context.Context, log *structs.CreateLogInput) (*structs.Log, error) {

	builder := r.data.EC.PaymentLog.Create().
		SetChannelID(log.ChannelID).
		SetType(string(log.Type))

	// Set optional fields
	if log.OrderID != "" {
		builder.SetOrderID(log.OrderID)
	}

	if log.StatusBefore != "" {
		builder.SetStatusBefore(string(log.StatusBefore))
	}

	if log.StatusAfter != "" {
		builder.SetStatusAfter(string(log.StatusAfter))
	}

	if log.RequestData != "" {
		builder.SetRequestData(log.RequestData)
	}

	if log.ResponseData != "" {
		builder.SetResponseData(log.ResponseData)
	}

	if log.IP != "" {
		builder.SetIP(log.IP)
	}

	if log.UserAgent != "" {
		builder.SetUserAgent(log.UserAgent)
	}

	if log.UserID != "" {
		builder.SetUserID(log.UserID)
	}

	if log.Error != "" {
		builder.SetError(log.Error)
	}

	if validator.IsNotEmpty(log.Metadata) {
		builder.SetExtras(log.Metadata)
	}

	// Create log
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment log: %w", err)
	}

	// Convert to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByID gets a payment log by ID
func (r *logRepository) GetByID(ctx context.Context, id string) (*structs.Log, error) {
	log, err := r.data.EC.PaymentLog.Query().
		Where(paymentLogEnt.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment log: %w", err)
	}

	return r.entToStruct(log)
}

// List lists payment logs
func (r *logRepository) List(ctx context.Context, query *structs.LogQuery) ([]*structs.Log, error) {
	// Build query
	q := r.data.EC.PaymentLog.Query()

	// Apply filters
	if query.OrderID != "" {
		q = q.Where(paymentLogEnt.OrderID(query.OrderID))
	}

	if query.ChannelID != "" {
		q = q.Where(paymentLogEnt.ChannelID(query.ChannelID))
	}

	if query.Type != "" {
		q = q.Where(paymentLogEnt.Type(string(query.Type)))
	}

	if query.HasError != nil {
		if *query.HasError {
			q = q.Where(paymentLogEnt.ErrorNEQ(""))
		} else {
			q = q.Where(paymentLogEnt.ErrorEQ(""))
		}
	}

	if query.StartDate > 0 {
		q = q.Where(paymentLogEnt.CreatedAtGTE(query.StartDate))
	}

	if query.EndDate > 0 {
		q = q.Where(paymentLogEnt.CreatedAtLTE(query.EndDate))
	}

	if query.UserID != "" {
		q = q.Where(paymentLogEnt.UserID(query.UserID))
	}

	// Apply cursor pagination
	if query.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(query.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		// Default direction is forward
		q.Where(
			paymentLogEnt.Or(
				paymentLogEnt.CreatedAtLT(timestamp),
				paymentLogEnt.And(
					paymentLogEnt.CreatedAtEQ(timestamp),
					paymentLogEnt.IDLT(id),
				),
			),
		)
	}

	// Set order - most recent first by default
	q.Order(ent.Desc(paymentLogEnt.FieldCreatedAt), ent.Desc(paymentLogEnt.FieldID))

	// Set limit
	if query.PageSize > 0 {
		q.Limit(query.PageSize)
	} else {
		q.Limit(20) // Default page size
	}

	// Execute query
	rows, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment logs: %w", err)
	}

	// Convert to structs
	var logs []*structs.Log
	for _, row := range rows {
		log, err := r.entToStruct(row)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// Count counts payment logs
func (r *logRepository) Count(ctx context.Context, query *structs.LogQuery) (int64, error) {
	// Build query without cursor, limit, and order
	q := r.data.EC.PaymentLog.Query()

	// Apply filters
	if query.OrderID != "" {
		q = q.Where(paymentLogEnt.OrderID(query.OrderID))
	}

	if query.ChannelID != "" {
		q = q.Where(paymentLogEnt.ChannelID(query.ChannelID))
	}

	if query.Type != "" {
		q = q.Where(paymentLogEnt.Type(string(query.Type)))
	}

	if query.HasError != nil {
		if *query.HasError {
			q = q.Where(paymentLogEnt.ErrorNEQ(""))
		} else {
			q = q.Where(paymentLogEnt.ErrorEQ(""))
		}
	}

	if query.StartDate > 0 {
		q = q.Where(paymentLogEnt.CreatedAtGTE(query.StartDate))
	}

	if query.EndDate > 0 {
		q = q.Where(paymentLogEnt.CreatedAtLTE(query.EndDate))
	}

	if query.UserID != "" {
		q = q.Where(paymentLogEnt.UserID(query.UserID))
	}

	// Execute count
	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count payment logs: %w", err)
	}

	return int64(count), nil
}

// entToStruct converts an Ent PaymentLog to a structs.ReadLog
func (r *logRepository) entToStruct(log *ent.PaymentLog) (*structs.Log, error) {
	return &structs.Log{
		ID:           log.ID,
		OrderID:      log.OrderID,
		ChannelID:    log.ChannelID,
		Type:         structs.LogType(log.Type),
		StatusBefore: structs.PaymentStatus(log.StatusBefore),
		StatusAfter:  structs.PaymentStatus(log.StatusAfter),
		RequestData:  log.RequestData,
		ResponseData: log.ResponseData,
		IP:           log.IP,
		UserAgent:    log.UserAgent,
		UserID:       log.UserID,
		Metadata:     log.Extras,
		Error:        log.Error,
		CreatedAt:    log.CreatedAt,
		UpdatedAt:    log.UpdatedAt,
	}, nil
}
