package repository

import (
	"context"
	"ncobase/plugin/sample/structs"
	"time"

	"github.com/google/uuid"
)

// SampleRepositoryInterface defines the interface for sample repository operations
type SampleRepositoryInterface interface {
	List(ctx context.Context) ([]*structs.Sample, error)
	GetByID(ctx context.Context, id string) (*structs.Sample, error)
	Create(ctx context.Context, input *structs.CreateSampleInput) (*structs.Sample, error)
	Update(ctx context.Context, id string, input *structs.UpdateSampleInput) (*structs.Sample, error)
	Delete(ctx context.Context, id string) error
}

// sampleRepository implements SampleRepositoryInterface
// This is a simple in-memory implementation for demonstration
// In production, you would use a database (Ent, GORM, etc.)
type sampleRepository struct {
	data map[string]*structs.Sample
}

// NewSampleRepository creates a new sample repository
func NewSampleRepository() SampleRepositoryInterface {
	return &sampleRepository{
		data: make(map[string]*structs.Sample),
	}
}

// List retrieves all samples
func (r *sampleRepository) List(ctx context.Context) ([]*structs.Sample, error) {
	samples := make([]*structs.Sample, 0, len(r.data))
	for _, sample := range r.data {
		samples = append(samples, sample)
	}
	return samples, nil
}

// GetByID retrieves a sample by ID
func (r *sampleRepository) GetByID(ctx context.Context, id string) (*structs.Sample, error) {
	sample, exists := r.data[id]
	if !exists {
		return nil, structs.ErrNotFound
	}
	return sample, nil
}

// Create creates a new sample
func (r *sampleRepository) Create(ctx context.Context, input *structs.CreateSampleInput) (*structs.Sample, error) {
	now := time.Now()
	sample := &structs.Sample{
		ID:          uuid.New().String(),
		Name:        input.Name,
		Description: input.Description,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.data[sample.ID] = sample
	return sample, nil
}

// Update updates an existing sample
func (r *sampleRepository) Update(ctx context.Context, id string, input *structs.UpdateSampleInput) (*structs.Sample, error) {
	sample, exists := r.data[id]
	if !exists {
		return nil, structs.ErrNotFound
	}

	if input.Name != nil {
		sample.Name = *input.Name
	}
	if input.Description != nil {
		sample.Description = *input.Description
	}
	if input.Status != nil {
		sample.Status = *input.Status
	}

	sample.UpdatedAt = time.Now()
	r.data[id] = sample

	return sample, nil
}

// Delete deletes a sample
func (r *sampleRepository) Delete(ctx context.Context, id string) error {
	if _, exists := r.data[id]; !exists {
		return structs.ErrNotFound
	}

	delete(r.data, id)
	return nil
}
