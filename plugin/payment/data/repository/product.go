package repository

import (
	"context"
	"fmt"
	"ncobase/payment/data"
	"ncobase/payment/data/ent"
	paymentOrderEnt "ncobase/payment/data/ent/paymentorder"
	paymentProductEnt "ncobase/payment/data/ent/paymentproduct"
	paymentSubscriptionEnt "ncobase/payment/data/ent/paymentsubscription"
	"ncobase/payment/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
)

// ProductRepositoryInterface defines the interface for product repository operations
type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *structs.CreateProductInput) (*structs.Product, error)
	GetByID(ctx context.Context, id string) (*structs.Product, error)
	Update(ctx context.Context, product *structs.UpdateProductInput) (*structs.Product, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, query *structs.ProductQuery) ([]*structs.Product, error)
	Count(ctx context.Context, query *structs.ProductQuery) (int64, error)
	IsInUse(ctx context.Context, id string) (bool, error)
}

// productRepository handles product persistence
type productRepository struct {
	data *data.Data
}

// NewProductRepository creates a new product repository
func NewProductRepository(d *data.Data) ProductRepositoryInterface {
	return &productRepository{data: d}
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *structs.CreateProductInput) (*structs.Product, error) {
	builder := r.data.EC.PaymentProduct.Create().
		SetName(product.Name).
		SetDescription(product.Description).
		SetStatus(string(product.Status)).
		SetPricingType(string(product.PricingType)).
		SetPrice(product.Price).
		SetCurrency(string(product.Currency)).
		SetTrialDays(product.TrialDays)

	// Set optional fields
	if product.BillingInterval != "" {
		builder.SetBillingInterval(string(product.BillingInterval))
	}

	if validator.IsNotEmpty(product.Features) {
		builder.SetFeatures(product.Features)
	}

	if product.TenantID != "" {
		builder.SetTenantID(product.TenantID)
	}

	if validator.IsNotEmpty(product.Metadata) {
		builder.SetExtras(product.Metadata)
	}

	// Create product
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Convert to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByID gets a product by ID
func (r *productRepository) GetByID(ctx context.Context, id string) (*structs.Product, error) {
	product, err := r.data.EC.PaymentProduct.Query().
		Where(paymentProductEnt.IDEQ(id)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return r.entToStruct(product)
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *structs.UpdateProductInput) (*structs.Product, error) {
	builder := r.data.EC.PaymentProduct.UpdateOneID(product.ID).
		SetName(product.Name).
		SetDescription(product.Description).
		SetStatus(string(product.Status)).
		SetNillablePrice(product.Price).
		SetNillableTrialDays(product.TrialDays)

	// Set optional fields
	if product.BillingInterval != "" {
		builder.SetBillingInterval(string(product.BillingInterval))
	}

	if validator.IsNotEmpty(product.Features) {
		builder.SetFeatures(product.Features)
	}

	if validator.IsNotEmpty(product.Metadata) {
		builder.SetExtras(product.Metadata)
	}

	// Update product
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Convert to struct
	result, err := r.entToStruct(row)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Delete deletes a product
func (r *productRepository) Delete(ctx context.Context, id string) error {
	return r.data.EC.PaymentProduct.DeleteOneID(id).Exec(ctx)
}

// List lists products with pagination
func (r *productRepository) List(ctx context.Context, query *structs.ProductQuery) ([]*structs.Product, error) {
	// Build query
	q := r.data.EC.PaymentProduct.Query()

	// Apply filters
	if query.Status != "" {
		q = q.Where(paymentProductEnt.Status(string(query.Status)))
	}

	if query.PricingType != "" {
		q = q.Where(paymentProductEnt.PricingType(string(query.PricingType)))
	}

	if query.TenantID != "" {
		q = q.Where(paymentProductEnt.TenantID(query.TenantID))
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
			paymentProductEnt.Or(
				paymentProductEnt.CreatedAtLT(timestamp),
				paymentProductEnt.And(
					paymentProductEnt.CreatedAtEQ(timestamp),
					paymentProductEnt.IDLT(id),
				),
			),
		)
	}

	// Set order - most recent first by default
	q.Order(ent.Desc(paymentProductEnt.FieldCreatedAt), ent.Desc(paymentProductEnt.FieldID))

	// Set limit
	if query.PageSize > 0 {
		q.Limit(query.PageSize)
	} else {
		q.Limit(20) // Default page size
	}

	// Execute query
	rows, err := q.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to structs
	var products []*structs.Product
	for _, row := range rows {
		product, err := r.entToStruct(row)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

// Count counts products
func (r *productRepository) Count(ctx context.Context, query *structs.ProductQuery) (int64, error) {
	// Build query without cursor, limit, and order
	q := r.data.EC.PaymentProduct.Query()

	// Apply filters
	if query.Status != "" {
		q = q.Where(paymentProductEnt.Status(string(query.Status)))
	}

	if query.PricingType != "" {
		q = q.Where(paymentProductEnt.PricingType(string(query.PricingType)))
	}

	if query.TenantID != "" {
		q = q.Where(paymentProductEnt.TenantID(query.TenantID))
	}

	// Execute count
	count, err := q.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return int64(count), nil
}

// IsInUse checks if a product is in use
func (r *productRepository) IsInUse(ctx context.Context, id string) (bool, error) {
	// Check if there are any orders using this product
	orderCount, err := r.data.EC.PaymentOrder.Query().
		Where(paymentOrderEnt.ProductIDEQ(id)).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if product is in use by orders: %w", err)
	}

	// Check if there are any subscriptions using this product
	subscriptionCount, err := r.data.EC.PaymentSubscription.Query().
		Where(paymentSubscriptionEnt.ProductIDEQ(id)).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if product is in use by subscriptions: %w", err)
	}

	return orderCount > 0 || subscriptionCount > 0, nil
}

// entToStruct converts an Ent PaymentProduct to a structs.Product
func (r *productRepository) entToStruct(product *ent.PaymentProduct) (*structs.Product, error) {
	return &structs.Product{
		ID:              product.ID,
		Name:            product.Name,
		Description:     product.Description,
		Status:          structs.ProductStatus(product.Status),
		PricingType:     structs.PricingType(product.PricingType),
		Price:           product.Price,
		Currency:        structs.CurrencyCode(product.Currency),
		BillingInterval: structs.BillingInterval(product.BillingInterval),
		TrialDays:       product.TrialDays,
		Features:        product.Features,
		TenantID:        product.TenantID,
		Metadata:        product.Extras,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}, nil
}
