package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/extension"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/validator"
	"ncobase/core/workflow/data/ent"
	"ncobase/core/workflow/data/repository"
	"ncobase/core/workflow/structs"
)

type DelegationServiceInterface interface {
	Create(ctx context.Context, body *structs.DelegationBody) (*structs.ReadDelegation, error)
	Get(ctx context.Context, params *structs.FindDelegationParams) (*structs.ReadDelegation, error)
	Update(ctx context.Context, body *structs.UpdateDelegationBody) (*structs.ReadDelegation, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListDelegationParams) (paging.Result[*structs.ReadDelegation], error)

	// Delegation specific operations

	EnableDelegation(ctx context.Context, id string) error
	DisableDelegation(ctx context.Context, id string) error
	GetActiveDelegations(ctx context.Context, userID string) ([]*structs.ReadDelegation, error)
	CheckDelegation(ctx context.Context, taskID string, assigneeID string) (string, error)
}

type delegationService struct {
	taskRepo       repository.TaskRepositoryInterface
	delegationRepo repository.DelegationRepositoryInterface
	em             *extension.Manager
}

func NewDelegationService(repo repository.Repository, em *extension.Manager) DelegationServiceInterface {
	return &delegationService{
		taskRepo:       repo.GetTask(),
		delegationRepo: repo.GetDelegation(),
		em:             em,
	}
}

// Create creates a new delegation
func (s *delegationService) Create(ctx context.Context, body *structs.DelegationBody) (*structs.ReadDelegation, error) {
	if body.DelegatorID == "" {
		return nil, errors.New(ecode.FieldIsRequired("delegator_id"))
	}
	if body.DelegateeID == "" {
		return nil, errors.New(ecode.FieldIsRequired("delegatee_id"))
	}

	// Validate time range
	if body.StartTime > body.EndTime {
		return nil, errors.New("start time must be before end time")
	}

	delegation, err := s.delegationRepo.Create(ctx, body)
	if err := handleEntError(ctx, "Delegation", err); err != nil {
		return nil, err
	}

	return s.serialize(delegation), nil
}

// Get retrieves a specific delegation
func (s *delegationService) Get(ctx context.Context, params *structs.FindDelegationParams) (*structs.ReadDelegation, error) {
	delegation, err := s.delegationRepo.Get(ctx, params)
	if err := handleEntError(ctx, "Delegation", err); err != nil {
		return nil, err
	}

	return s.serialize(delegation), nil
}

// Update updates an existing delegation
func (s *delegationService) Update(ctx context.Context, body *structs.UpdateDelegationBody) (*structs.ReadDelegation, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate time range if provided
	if validator.IsNotEmpty(body.StartTime) && validator.IsNotEmpty(body.EndTime) {
		if body.StartTime > body.EndTime {
			return nil, errors.New("start time must be before end time")
		}
	}

	delegation, err := s.delegationRepo.Update(ctx, body)
	if err := handleEntError(ctx, "Delegation", err); err != nil {
		return nil, err
	}

	return s.serialize(delegation), nil
}

// Delete deletes a delegation
func (s *delegationService) Delete(ctx context.Context, id string) error {
	return s.delegationRepo.Delete(ctx, id)
}

// List returns a list of delegations
func (s *delegationService) List(ctx context.Context, params *structs.ListDelegationParams) (paging.Result[*structs.ReadDelegation], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadDelegation, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		delegations, err := s.delegationRepo.List(ctx, &lp)
		if err != nil {
			log.Errorf(ctx, "Error listing delegations: %v", err)
			return nil, 0, err
		}

		return s.serializes(delegations), len(delegations), nil
	})
}

// EnableDelegation enables a delegation
func (s *delegationService) EnableDelegation(ctx context.Context, id string) error {
	return s.delegationRepo.EnableDelegation(ctx, id)
}

// DisableDelegation disables a delegation
func (s *delegationService) DisableDelegation(ctx context.Context, id string) error {
	return s.delegationRepo.DisableDelegation(ctx, id)
}

// GetActiveDelegations gets active delegations for a user
func (s *delegationService) GetActiveDelegations(ctx context.Context, userID string) ([]*structs.ReadDelegation, error) {
	delegations, err := s.delegationRepo.GetActiveDelegations(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.serializes(delegations), nil
}

// CheckDelegation checks if there is valid delegation for a task
func (s *delegationService) CheckDelegation(ctx context.Context, taskID string, assigneeID string) (string, error) {
	// Get task details
	task, err := s.taskRepo.Get(ctx, &structs.FindTaskParams{
		ProcessID: taskID,
	})
	if err != nil {
		return "", err
	}

	// Get active delegations
	delegations, err := s.GetActiveDelegations(ctx, assigneeID)
	if err != nil {
		return "", err
	}

	// Check each delegation
	for _, delegation := range delegations {
		// Check template specific delegation
		if delegation.TemplateID != "" && delegation.TemplateID != task.ProcessID {
			continue
		}

		// Check node type specific delegation
		if delegation.NodeType != "" && delegation.NodeType != task.NodeType {
			continue
		}

		// Check conditions if any
		if delegation.Conditions != nil {
			// TODO: Evaluate delegation conditions
			continue
		}

		// Valid delegation found
		return delegation.DelegateeID, nil
	}

	return "", nil
}

// Serialization helpers
func (s *delegationService) serialize(delegation *ent.Delegation) *structs.ReadDelegation {
	if delegation == nil {
		return nil
	}

	return &structs.ReadDelegation{
		ID:          delegation.ID,
		DelegatorID: delegation.DelegatorID,
		DelegateeID: delegation.DelegateeID,
		TemplateID:  delegation.TemplateID,
		NodeType:    delegation.NodeType,
		Conditions:  delegation.Conditions,
		StartTime:   delegation.StartTime,
		EndTime:     delegation.EndTime,
		IsEnabled:   delegation.IsEnabled,
		Status:      delegation.Status,
		TenantID:    delegation.TenantID,
		Extras:      delegation.Extras,
		CreatedBy:   &delegation.CreatedBy,
		CreatedAt:   &delegation.CreatedAt,
		UpdatedBy:   &delegation.UpdatedBy,
		UpdatedAt:   &delegation.UpdatedAt,
	}
}

func (s *delegationService) serializes(delegations []*ent.Delegation) []*structs.ReadDelegation {
	result := make([]*structs.ReadDelegation, len(delegations))
	for i, delegation := range delegations {
		result[i] = s.serialize(delegation)
	}
	return result
}
