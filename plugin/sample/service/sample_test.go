package service

import (
	"context"
	"ncobase/plugin/sample/data/repository"
	"ncobase/plugin/sample/structs"
	"testing"
)

func TestSampleService_Create(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *structs.CreateSampleInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &structs.CreateSampleInput{
				Name:        "Test Sample",
				Description: "Test Description",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: &structs.CreateSampleInput{
				Name:        "",
				Description: "Test Description",
			},
			wantErr: true,
		},
		{
			name: "valid without description",
			input: &structs.CreateSampleInput{
				Name: "Test Sample",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sample, err := service.Sample.Create(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if sample == nil {
					t.Error("Expected sample to be created")
					return
				}
				if sample.Name != tt.input.Name {
					t.Errorf("Expected name %s, got %s", tt.input.Name, sample.Name)
				}
			}
		})
	}
}

func TestSampleService_GetByID(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	// Create a sample first
	input := &structs.CreateSampleInput{
		Name:        "Test Sample",
		Description: "Test Description",
	}
	created, err := service.Sample.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test GetByID with valid ID
	found, err := service.Sample.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, found.ID)
	}

	// Test GetByID with empty ID
	_, err = service.Sample.GetByID(ctx, "")
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// Test GetByID with non-existing ID
	_, err = service.Sample.GetByID(ctx, "non-existing-id")
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestSampleService_List(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	// Initially empty
	samples, err := service.Sample.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(samples) != 0 {
		t.Errorf("Expected 0 samples, got %d", len(samples))
	}

	// Create samples
	for i := 0; i < 5; i++ {
		input := &structs.CreateSampleInput{
			Name:        "Test Sample",
			Description: "Test Description",
		}
		_, err := service.Sample.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List should return all samples
	samples, err = service.Sample.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(samples) != 5 {
		t.Errorf("Expected 5 samples, got %d", len(samples))
	}
}

func TestSampleService_Update(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	// Create a sample first
	input := &structs.CreateSampleInput{
		Name:        "Original Name",
		Description: "Original Description",
	}
	created, err := service.Sample.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test valid update
	newName := "Updated Name"
	updateInput := &structs.UpdateSampleInput{
		Name: &newName,
	}
	updated, err := service.Sample.Update(ctx, created.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, updated.Name)
	}

	// Test update with empty ID
	_, err = service.Sample.Update(ctx, "", updateInput)
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// Test update with non-existing ID
	_, err = service.Sample.Update(ctx, "non-existing-id", updateInput)
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}

	// Test update with empty name
	emptyName := ""
	invalidInput := &structs.UpdateSampleInput{
		Name: &emptyName,
	}
	_, err = service.Sample.Update(ctx, created.ID, invalidInput)
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

func TestSampleService_Delete(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	// Create a sample first
	input := &structs.CreateSampleInput{
		Name:        "Test Sample",
		Description: "Test Description",
	}
	created, err := service.Sample.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test valid delete
	err = service.Sample.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's deleted
	_, err = service.Sample.GetByID(ctx, created.ID)
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}

	// Test delete with empty ID
	err = service.Sample.Delete(ctx, "")
	if err == nil {
		t.Error("Expected error for empty ID")
	}

	// Test delete with non-existing ID
	err = service.Sample.Delete(ctx, "non-existing-id")
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestSampleService_ValidationRules(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *structs.CreateSampleInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid input",
			input: &structs.CreateSampleInput{
				Name:        "Valid Name",
				Description: "Valid Description",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			input: &structs.CreateSampleInput{
				Name:        "",
				Description: "Description",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "whitespace only name",
			input: &structs.CreateSampleInput{
				Name:        "   ",
				Description: "Description",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Sample.Create(ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSampleService_NilInputValidation(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	if _, err := service.Sample.Create(ctx, nil); err != structs.ErrInvalidInput {
		t.Fatalf("Create() expected ErrInvalidInput, got %v", err)
	}

	if _, err := service.Sample.Update(ctx, "some-id", nil); err != structs.ErrInvalidInput {
		t.Fatalf("Update() expected ErrInvalidInput, got %v", err)
	}
}

func TestSampleService_TrimNameOnWrite(t *testing.T) {
	repo := repository.NewSampleRepository()
	service := NewService(&repository.Repository{Sample: repo})
	ctx := context.Background()

	created, err := service.Sample.Create(ctx, &structs.CreateSampleInput{
		Name:        "  Trimmed Name  ",
		Description: "desc",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.Name != "Trimmed Name" {
		t.Fatalf("expected trimmed name, got %q", created.Name)
	}

	nextName := "  Updated Name  "
	updated, err := service.Sample.Update(ctx, created.ID, &structs.UpdateSampleInput{Name: &nextName})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Fatalf("expected trimmed updated name, got %q", updated.Name)
	}
}
