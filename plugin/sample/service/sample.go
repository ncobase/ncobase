package service

import (
	"context"
	"ncobase/plugin/sample/data/repository"
	"ncobase/plugin/sample/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// SampleServiceInterface defines the interface for sample service operations
type SampleServiceInterface interface {
	List(ctx context.Context) ([]*structs.Sample, error)
	GetByID(ctx context.Context, id string) (*structs.Sample, error)
	Create(ctx context.Context, input *structs.CreateSampleInput) (*structs.Sample, error)
	Update(ctx context.Context, id string, input *structs.UpdateSampleInput) (*structs.Sample, error)
	Delete(ctx context.Context, id string) error
}

// sampleService implements SampleServiceInterface
type sampleService struct {
	repo repository.SampleRepositoryInterface
}

// NewSampleService creates a new sample service
func NewSampleService(repo repository.SampleRepositoryInterface) SampleServiceInterface {
	return &sampleService{
		repo: repo,
	}
}

// List retrieves all samples
func (s *sampleService) List(ctx context.Context) ([]*structs.Sample, error) {
	logger.Debugf(ctx, "Listing all samples")

	samples, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	return samples, nil
}

// GetByID retrieves a sample by ID
func (s *sampleService) GetByID(ctx context.Context, id string) (*structs.Sample, error) {
	logger.Debugf(ctx, "Getting sample by ID: %s", id)

	sample, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return sample, nil
}

// Create creates a new sample
func (s *sampleService) Create(ctx context.Context, input *structs.CreateSampleInput) (*structs.Sample, error) {
	logger.Infof(ctx, "Creating sample: %s", input.Name)

	// Business logic validation
	if input.Name == "" {
		return nil, structs.ErrInvalidInput
	}

	sample, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "Sample created successfully: %s", sample.ID)
	return sample, nil
}

// Update updates an existing sample
func (s *sampleService) Update(ctx context.Context, id string, input *structs.UpdateSampleInput) (*structs.Sample, error) {
	logger.Infof(ctx, "Updating sample: %s", id)

	// Check if sample exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Business logic validation
	if input.Name != nil && *input.Name == "" {
		return nil, structs.ErrInvalidInput
	}

	sample, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "Sample updated successfully: %s (was: %s)", sample.Name, existing.Name)
	return sample, nil
}

// Delete deletes a sample
func (s *sampleService) Delete(ctx context.Context, id string) error {
	logger.Infof(ctx, "Deleting sample: %s", id)

	// Check if sample exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	logger.Infof(ctx, "Sample deleted successfully: %s", id)
	return nil
}
