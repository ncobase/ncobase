package repository

import (
	"context"
	"fmt"
	"ncobase/payment/data"
	"ncobase/payment/data/ent"
	paymentChannelEnt "ncobase/payment/data/ent/paymentchannel"
	paymentOrderEnt "ncobase/payment/data/ent/paymentorder"
	"ncobase/payment/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
)

// ChannelRepositoryInterface defines the interface for channel repository operations
type ChannelRepositoryInterface interface {
	Create(ctx context.Context, channel *structs.CreateChannelInput) (*structs.Channel, error)
	GetByID(ctx context.Context, id string) (*structs.Channel, error)
	Update(ctx context.Context, channel *structs.UpdateChannelInput) (*structs.Channel, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, query *structs.ChannelQuery) ([]*structs.Channel, error)
	Count(ctx context.Context, query *structs.ChannelQuery) (int64, error)
	UnsetDefault(ctx context.Context, provider string, tenantID string) error
	GetDefault(ctx context.Context, provider string, tenantID string) (*structs.Channel, error)
	IsInUse(ctx context.Context, id string) (bool, error)
}

// channelRepository handles payment channel persistence
type channelRepository struct {
	data *data.Data
}

// NewChannelRepository creates a new channel repository
func NewChannelRepository(d *data.Data) ChannelRepositoryInterface {
	return &channelRepository{data: d}
}

// Create creates a new payment channel
func (r *channelRepository) Create(ctx context.Context, channel *structs.CreateChannelInput) (*structs.Channel, error) {
	// // Convert supported types to JSON
	// supportedTypesJSON, err := convert.ToJSON(channel.SupportedType)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to convert supported types to JSON: %w", err)
	// }

	// Begin a transaction
	tx, err := r.data.EC.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Errorf(ctx, "failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	// If channel is set as default, unset any existing default
	if channel.IsDefault {
		_, err = tx.PaymentChannel.Update().
			Where(
				paymentChannelEnt.Provider(string(channel.Provider)),
				paymentChannelEnt.IsDefault(true),
				paymentChannelEnt.TenantIDEQ(channel.TenantID),
			).
			SetIsDefault(false).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to unset existing default channels: %w", err)
		}
	}

	// Create channel
	row, err := tx.PaymentChannel.Create().
		SetName(channel.Name).
		SetProvider(string(channel.Provider)).
		SetStatus(string(channel.Status)).
		SetIsDefault(channel.IsDefault).
		SetSupportedTypes(convert.ToStringArray(channel.SupportedType)).
		SetConfig(channel.Config).
		SetExtras(channel.Metadata).
		SetTenantID(channel.TenantID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment channel: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert back to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByID gets a payment channel by ID
func (r *channelRepository) GetByID(ctx context.Context, id string) (*structs.Channel, error) {
	channel, err := r.data.EC.PaymentChannel.Query().
		Where(paymentChannelEnt.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment channel: %w", err)
	}

	return r.entToStruct(channel)
}

// Update updates a payment channel
func (r *channelRepository) Update(ctx context.Context, channel *structs.UpdateChannelInput) (*structs.Channel, error) {
	// Begin a transaction
	tx, err := r.data.EC.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Errorf(ctx, "failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	// If channel is being set as default, unset any existing default
	if convert.ToValue(channel.IsDefault) {
		// Get current channel
		existing, err := tx.PaymentChannel.Get(ctx, channel.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing channel: %w", err)
		}

		// If not already default, unset others
		if !existing.IsDefault {
			_, err = tx.PaymentChannel.Update().
				Where(
					paymentChannelEnt.Provider(string(channel.Provider)),
					paymentChannelEnt.IsDefault(true),
					paymentChannelEnt.IDNEQ(channel.ID),
					paymentChannelEnt.TenantIDEQ(channel.TenantID),
				).
				SetIsDefault(false).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to unset existing default channels: %w", err)
			}
		}
	}

	// Update channel
	row, err := tx.PaymentChannel.UpdateOneID(channel.ID).
		SetName(channel.Name).
		SetStatus(string(channel.Status)).
		SetIsDefault(convert.ToValue(channel.IsDefault)).
		SetSupportedTypes(convert.ToStringArray(channel.SupportedType)).
		SetConfig(channel.Config).
		SetExtras(channel.Metadata).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment channel: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Convert back to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete deletes a payment channel
func (r *channelRepository) Delete(ctx context.Context, id string) error {
	return r.data.EC.PaymentChannel.DeleteOneID(id).Exec(ctx)
}

// List lists payment channels with pagination
func (r *channelRepository) List(ctx context.Context, query *structs.ChannelQuery) ([]*structs.Channel, error) {
	// Build query
	q := r.data.EC.PaymentChannel.Query()

	// Apply filters
	if query.Provider != "" {
		q = q.Where(paymentChannelEnt.Provider(string(query.Provider)))
	}

	if query.Status != "" {
		q = q.Where(paymentChannelEnt.Status(string(query.Status)))
	}

	if query.TenantID != "" {
		q = q.Where(paymentChannelEnt.TenantID(query.TenantID))
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
			paymentChannelEnt.Or(
				paymentChannelEnt.CreatedAtLT(timestamp),
				paymentChannelEnt.And(
					paymentChannelEnt.CreatedAtEQ(timestamp),
					paymentChannelEnt.IDLT(id),
				),
			),
		)
	}

	// Set order - most recent first by default
	q.Order(ent.Desc(paymentChannelEnt.FieldCreatedAt), ent.Desc(paymentChannelEnt.FieldID))

	// Set limit
	if query.PageSize > 0 {
		q.Limit(query.PageSize)
	} else {
		q.Limit(20) // Default page size
	}

	// Execute query
	rows, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment channels: %w", err)
	}

	// Convert to structs
	var channels []*structs.Channel
	for _, row := range rows {
		channel, err := r.entToStruct(row)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

// Count counts payment channels
func (r *channelRepository) Count(ctx context.Context, query *structs.ChannelQuery) (int64, error) {
	// Build query without cursor, limit, and order
	q := r.data.EC.PaymentChannel.Query()

	// Apply filters
	if query.Provider != "" {
		q = q.Where(paymentChannelEnt.Provider(string(query.Provider)))
	}

	if query.Status != "" {
		q = q.Where(paymentChannelEnt.Status(string(query.Status)))
	}

	if query.TenantID != "" {
		q = q.Where(paymentChannelEnt.TenantID(query.TenantID))
	}

	// Execute count
	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count payment channels: %w", err)
	}

	return int64(count), nil
}

// UnsetDefault unsets the default channel for a provider
func (r *channelRepository) UnsetDefault(ctx context.Context, provider string, tenantID string) error {
	_, err := r.data.EC.PaymentChannel.Update().
		Where(
			paymentChannelEnt.Provider(provider),
			paymentChannelEnt.IsDefault(true),
			paymentChannelEnt.TenantIDEQ(tenantID),
		).
		SetIsDefault(false).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to unset default channel: %w", err)
	}

	return nil
}

// GetDefault gets the default channel for a provider
func (r *channelRepository) GetDefault(ctx context.Context, provider string, tenantID string) (*structs.Channel, error) {
	channel, err := r.data.EC.PaymentChannel.Query().
		Where(
			paymentChannelEnt.Provider(provider),
			paymentChannelEnt.IsDefault(true),
			paymentChannelEnt.TenantIDEQ(tenantID),
		).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get default channel: %w", err)
	}

	return r.entToStruct(channel)
}

// IsInUse checks if a channel is in use
func (r *channelRepository) IsInUse(ctx context.Context, id string) (bool, error) {
	// Check if there are any orders using this channel
	count, err := r.data.EC.PaymentOrder.Query().
		Where(paymentOrderEnt.ChannelID(id)).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if channel is in use: %w", err)
	}

	return count > 0, nil
}

// entToStruct converts an Ent PaymentChannel to a structs.ReadChannel
func (r *channelRepository) entToStruct(channel *ent.PaymentChannel) (*structs.Channel, error) {
	return &structs.Channel{
		ID:        channel.ID,
		CreatedAt: channel.CreatedAt,
		UpdatedAt: channel.UpdatedAt,
		Name:      channel.Name,
		Provider:  structs.PaymentProvider(channel.Provider),
		Status:    structs.ChannelStatus(channel.Status),
		IsDefault: channel.IsDefault,
		SupportedType: func() []structs.PaymentType {
			types, _ := convert.ToTypedArray[structs.PaymentType](channel.SupportedTypes)
			return types
		}(),
		Config:   channel.Config,
		Metadata: channel.Extras,
		TenantID: channel.TenantID,
	}, nil
}
