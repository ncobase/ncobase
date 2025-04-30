package service

import (
	"context"
	"fmt"
	"ncobase/core/payment/data/repository"
	"ncobase/core/payment/event"
	"ncobase/core/payment/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

// ProductServiceInterface defines the interface for product service operations
type ProductServiceInterface interface {
	Create(ctx context.Context, input *structs.CreateProductInput) (*structs.Product, error)
	GetByID(ctx context.Context, id string) (*structs.Product, error)
	Update(ctx context.Context, id string, input *structs.UpdateProductInput) (*structs.Product, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, query *structs.ProductQuery) (paging.Result[*structs.Product], error)
	Serialize(product *structs.Product) *structs.Product
	Serializes(products []*structs.Product) []*structs.Product
}

// productService provides operations for products
type productService struct {
	repo      repository.ProductRepositoryInterface
	publisher event.PublisherInterface
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepositoryInterface, publisher event.PublisherInterface) ProductServiceInterface {
	return &productService{
		repo:      repo,
		publisher: publisher,
	}
}

// Create creates a new product
func (s *productService) Create(ctx context.Context, input *structs.CreateProductInput) (*structs.Product, error) {
	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("name"))
	}

	if input.Description == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("description"))
	}

	if input.Price < 0 {
		return nil, fmt.Errorf(ecode.FieldIsInvalid("price must be non-negative"))
	}

	// Set default currency if not provided
	if input.Currency == "" {
		input.Currency = structs.CurrencyUSD
	}

	// Create product entity
	product := &structs.CreateProductInput{
		Name:            input.Name,
		Description:     input.Description,
		Status:          input.Status,
		PricingType:     input.PricingType,
		Price:           input.Price,
		Currency:        input.Currency,
		BillingInterval: input.BillingInterval,
		TrialDays:       input.TrialDays,
		Features:        input.Features,
		TenantID:        input.TenantID,
		Metadata:        input.Metadata,
	}

	// Save to database
	created, err := s.repo.Create(ctx, product)
	if err != nil {
		return nil, handleEntError(ctx, "Product", err)
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.ProductEventData{
			ProductID:       created.ID,
			Name:            created.Name,
			Status:          created.Status,
			PricingType:     created.PricingType,
			Price:           created.Price,
			Currency:        created.Currency,
			BillingInterval: created.BillingInterval,
			TenantID:        created.TenantID,
			Metadata:        created.Metadata,
		}

		s.publisher.PublishProductCreated(ctx, eventData)
	}

	return s.Serialize(created), nil
}

// GetByID gets a product by ID
func (s *productService) GetByID(ctx context.Context, id string) (*structs.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	return s.repo.GetByID(ctx, id)
}

// Update updates a product
func (s *productService) Update(ctx context.Context, id string, input *structs.UpdateProductInput) (*structs.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	// Save to database
	updated, err := s.repo.Update(ctx, input)
	if err != nil {
		return nil, err
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.ProductEventData{
			ProductID:       updated.ID,
			Name:            updated.Name,
			Status:          updated.Status,
			PricingType:     updated.PricingType,
			Price:           updated.Price,
			Currency:        updated.Currency,
			BillingInterval: updated.BillingInterval,
			TenantID:        updated.TenantID,
			Metadata:        updated.Metadata,
		}

		s.publisher.PublishProductUpdated(ctx, eventData)
	}

	return updated, nil
}

// Delete deletes a product
func (s *productService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("product ID is required")
	}

	// Check if the product is in use
	inUse, err := s.repo.IsInUse(ctx, id)
	if err != nil {
		return err
	}

	if inUse {
		return fmt.Errorf("product is in use and cannot be deleted")
	}

	// Get existing product for event data
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.ProductEventData{
			ProductID:       existing.ID,
			Name:            existing.Name,
			Status:          existing.Status,
			PricingType:     existing.PricingType,
			Price:           existing.Price,
			Currency:        existing.Currency,
			BillingInterval: existing.BillingInterval,
			TenantID:        existing.TenantID,
			Metadata:        existing.Metadata,
		}

		s.publisher.PublishProductDeleted(ctx, eventData)
	}

	return nil
}

// List lists products with pagination
func (s *productService) List(ctx context.Context, query *structs.ProductQuery) (paging.Result[*structs.Product], error) {
	pp := paging.Params{
		Cursor:    query.Cursor,
		Limit:     query.PageSize,
		Direction: "forward", // Default direction
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.Product, int, error) {
		lq := *query
		lq.Cursor = cursor
		lq.PageSize = limit

		products, err := s.repo.List(ctx, &lq)
		if err != nil {
			logger.Errorf(ctx, "Error listing products: %v", err)
			return nil, 0, err
		}

		total, err := s.repo.Count(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Error counting products: %v", err)
			return nil, 0, err
		}

		return s.Serializes(products), int(total), nil
	})
}

// Serialize serializes a product entity to a response format
func (s *productService) Serialize(product *structs.Product) *structs.Product {
	if product == nil {
		return nil
	}

	return &structs.Product{
		ID:              product.ID,
		Name:            product.Name,
		Description:     product.Description,
		Status:          product.Status,
		PricingType:     product.PricingType,
		Price:           product.Price,
		Currency:        product.Currency,
		BillingInterval: product.BillingInterval,
		TrialDays:       product.TrialDays,
		Features:        product.Features,
		TenantID:        product.TenantID,
		Metadata:        product.Metadata,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}
}

// Serializes serializes multiple product entities to response format
func (s *productService) Serializes(products []*structs.Product) []*structs.Product {
	result := make([]*structs.Product, len(products))
	for i, product := range products {
		result[i] = s.Serialize(product)
	}
	return result
}
