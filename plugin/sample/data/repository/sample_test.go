package repository

import (
	"context"
	"ncobase/plugin/sample/structs"
	"testing"
)

func TestSampleRepository_Create(t *testing.T) {
	ctx := context.Background()

	repo := NewSampleRepository()
	input := &structs.CreateSampleInput{
		Name:        "Test Sample",
		Description: "Test Description",
	}

	sample, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if sample.ID == "" {
		t.Error("Expected ID to be generated")
	}
	if sample.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, sample.Name)
	}
	if sample.Description != input.Description {
		t.Errorf("Expected description %s, got %s", input.Description, sample.Description)
	}
	if sample.Status != "active" {
		t.Errorf("Expected status 'active', got %s", sample.Status)
	}
	if sample.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if sample.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestSampleRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	repo := NewSampleRepository()
	// Create a sample first
	input := &structs.CreateSampleInput{
		Name:        "Test Sample",
		Description: "Test Description",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test GetByID with existing ID
	found, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, found.ID)
	}

	// Test GetByID with non-existing ID
	_, err = repo.GetByID(ctx, "non-existing-id")
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestSampleRepository_List(t *testing.T) {
	ctx := context.Background()

	repo := NewSampleRepository()
	// Initially empty
	samples, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(samples) != 0 {
		t.Errorf("Expected 0 samples, got %d", len(samples))
	}

	// Create multiple samples
	for i := 0; i < 3; i++ {
		input := &structs.CreateSampleInput{
			Name:        "Test Sample",
			Description: "Test Description",
		}
		_, err := repo.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List should return all samples
	samples, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(samples) != 3 {
		t.Errorf("Expected 3 samples, got %d", len(samples))
	}
}

func TestSampleRepository_Update(t *testing.T) {
	ctx := context.Background()

	repo := NewSampleRepository()
	// Create a sample first
	input := &structs.CreateSampleInput{
		Name:        "Original Name",
		Description: "Original Description",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update the sample
	newName := "Updated Name"
	newDesc := "Updated Description"
	newStatus := "inactive"
	updateInput := &structs.UpdateSampleInput{
		Name:        &newName,
		Description: &newDesc,
		Status:      &newStatus,
	}

	updated, err := repo.Update(ctx, created.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, updated.Name)
	}
	if updated.Description != newDesc {
		t.Errorf("Expected description %s, got %s", newDesc, updated.Description)
	}
	if updated.Status != newStatus {
		t.Errorf("Expected status %s, got %s", newStatus, updated.Status)
	}
	if updated.UpdatedAt.Before(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}

	// Test update with non-existing ID
	_, err = repo.Update(ctx, "non-existing-id", updateInput)
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestSampleRepository_Delete(t *testing.T) {
	ctx := context.Background()

	repo := NewSampleRepository()
	// Create a sample first
	input := &structs.CreateSampleInput{
		Name:        "Test Sample",
		Description: "Test Description",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the sample
	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's deleted
	_, err = repo.GetByID(ctx, created.ID)
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound after delete, got %v", err)
	}

	// Test delete with non-existing ID
	err = repo.Delete(ctx, "non-existing-id")
	if err != structs.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestSampleRepository_UpdatePartial(t *testing.T) {
	ctx := context.Background()

	repo := NewSampleRepository()
	// Create a sample
	input := &structs.CreateSampleInput{
		Name:        "Original Name",
		Description: "Original Description",
	}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update only name
	newName := "Updated Name Only"
	updateInput := &structs.UpdateSampleInput{
		Name: &newName,
	}

	updated, err := repo.Update(ctx, created.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, updated.Name)
	}
	if updated.Description != created.Description {
		t.Errorf("Expected description to remain %s, got %s", created.Description, updated.Description)
	}
	if updated.Status != created.Status {
		t.Errorf("Expected status to remain %s, got %s", created.Status, updated.Status)
	}
}

func TestSampleRepository_ReturnsDefensiveCopies(t *testing.T) {
	ctx := context.Background()

	repo := NewSampleRepository()
	created, err := repo.Create(ctx, &structs.CreateSampleInput{
		Name:        "Original Name",
		Description: "Original Description",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	created.Name = "Mutated Outside"
	fetched, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.Name != "Original Name" {
		t.Fatalf("expected stored name to remain unchanged, got %q", fetched.Name)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one sample, got %d", len(list))
	}

	list[0].Description = "Mutated In List"
	refetched, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if refetched.Description != "Original Description" {
		t.Fatalf("expected stored description to remain unchanged, got %q", refetched.Description)
	}
}
