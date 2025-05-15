package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/workflow/data/ent"
	"ncobase/workflow/data/repository"
	"ncobase/workflow/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

type ProcessServiceInterface interface {
	Create(ctx context.Context, body *structs.ProcessBody) (*structs.ReadProcess, error)
	Get(ctx context.Context, params *structs.FindProcessParams) (*structs.ReadProcess, error)
	Update(ctx context.Context, body *structs.UpdateProcessBody) (*structs.ReadProcess, error)
	Delete(ctx context.Context, params *structs.FindProcessParams) error
	List(ctx context.Context, params *structs.ListProcessParams) (paging.Result[*structs.ReadProcess], error)

	// Process specific operations

	Start(ctx context.Context, req *structs.StartProcessRequest) (*structs.StartProcessResponse, error)
	Complete(ctx context.Context, processID string) error
	Terminate(ctx context.Context, req *structs.TerminateProcessRequest) error
	Suspend(ctx context.Context, processID string, reason string) error
	Resume(ctx context.Context, processID string) error
}

type processService struct {
	processRepo  repository.ProcessRepositoryInterface
	templateRepo repository.TemplateRepositoryInterface
	historyRepo  repository.HistoryRepositoryInterface
	em           ext.ManagerInterface
}

func NewProcessService(repo repository.Repository, em ext.ManagerInterface) ProcessServiceInterface {
	return &processService{
		processRepo:  repo.GetProcess(),
		templateRepo: repo.GetTemplate(),
		historyRepo:  repo.GetHistory(),
		em:           em,
	}
}

// Create creates a new process
func (s *processService) Create(ctx context.Context, body *structs.ProcessBody) (*structs.ReadProcess, error) {
	if body.TemplateID == "" {
		return nil, errors.New(ecode.FieldIsRequired("template_id"))
	}

	// Get process template
	template, err := s.templateRepo.Get(ctx, &structs.FindTemplateParams{
		Code: body.TemplateID,
	})
	if err := handleEntError(ctx, "Template", err); err != nil {
		return nil, err
	}

	// Generate process key
	body.ProcessKey = fmt.Sprintf("%s-%s", template.Code, body.BusinessKey)

	process, err := s.processRepo.Create(ctx, body)
	if err := handleEntError(ctx, "Process", err); err != nil {
		return nil, err
	}

	// Publish event
	s.em.PublishEvent(string(structs.EventProcessStarted), &structs.EventData{
		ProcessID:    process.ID,
		ProcessName:  body.ProcessCode,
		TemplateID:   template.ID,
		TemplateName: template.Name,
		ModuleCode:   body.ModuleCode,
		FormCode:     body.FormCode,
	})

	return s.serialize(process), nil
}

// Start starts a new process instance
func (s *processService) Start(ctx context.Context, req *structs.StartProcessRequest) (*structs.StartProcessResponse, error) {
	// Validate request
	if req.TemplateID == "" {
		return nil, errors.New(ecode.FieldIsRequired("template_id"))
	}
	if req.BusinessKey == "" {
		return nil, errors.New(ecode.FieldIsRequired("business_key"))
	}

	// Create process instance
	process, err := s.Create(ctx, &structs.ProcessBody{
		TemplateID:    req.TemplateID,
		BusinessKey:   req.BusinessKey,
		ModuleCode:    req.ModuleCode,
		FormCode:      req.FormCode,
		Status:        string(structs.StatusActive),
		Initiator:     req.Initiator,
		InitiatorDept: req.InitiatorDept,
		Variables:     req.Variables,
	})
	if err != nil {
		return nil, err
	}

	return &structs.StartProcessResponse{
		ProcessID: process.ID,
		Status:    structs.Status(process.Status),
		StartTime: process.StartTime,
		Variables: process.Variables,
	}, nil
}

// Get retrieves a specific process
func (s *processService) Get(ctx context.Context, params *structs.FindProcessParams) (*structs.ReadProcess, error) {
	process, err := s.processRepo.Get(ctx, params)
	if err := handleEntError(ctx, "Process", err); err != nil {
		return nil, err
	}

	return s.serialize(process), nil
}

// Update updates an existing process
func (s *processService) Update(ctx context.Context, body *structs.UpdateProcessBody) (*structs.ReadProcess, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	process, err := s.processRepo.Update(ctx, body)
	if err := handleEntError(ctx, "Process", err); err != nil {
		return nil, err
	}

	return s.serialize(process), nil
}

// Delete deletes a process
func (s *processService) Delete(ctx context.Context, params *structs.FindProcessParams) error {
	return handleEntError(ctx, "Process", s.processRepo.Delete(ctx, params))
}

// List returns a list of processes
func (s *processService) List(ctx context.Context, params *structs.ListProcessParams) (paging.Result[*structs.ReadProcess], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadProcess, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		processes, err := s.processRepo.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing processes: %v", err)
			return nil, 0, err
		}

		total := s.processRepo.CountX(ctx, params)

		return s.serializes(processes), total, nil
	})
}

// Complete completes a process
func (s *processService) Complete(ctx context.Context, processID string) error {
	process, err := s.processRepo.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err := handleEntError(ctx, "Process", err); err != nil {
		return err
	}

	// Update process status
	_, err = s.processRepo.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status: string(structs.StatusCompleted),
		},
	})
	if err != nil {
		return err
	}

	// Publish event
	s.em.PublishEvent(string(structs.EventProcessCompleted), &structs.EventData{
		ProcessID: process.ID,
	})

	return nil
}

// Terminate terminates a process
func (s *processService) Terminate(ctx context.Context, req *structs.TerminateProcessRequest) error {
	process, err := s.processRepo.Get(ctx, &structs.FindProcessParams{
		ProcessKey: req.ProcessID,
	})
	if err := handleEntError(ctx, "Process", err); err != nil {
		return err
	}

	// Update process status
	_, err = s.processRepo.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status: string(structs.StatusTerminated),
		},
	})
	if err != nil {
		return err
	}

	// Create history record
	_, err = s.historyRepo.Create(ctx, &structs.HistoryBody{
		Type:      "process",
		ProcessID: process.ID,
		Operator:  req.Operator,
		Action:    string(structs.ActionTerminate),
		Comment:   req.Comment,
	})
	if err != nil {
		logger.Errorf(ctx, "Error creating termination history: %v", err)
	}

	return nil
}

// Suspend suspends a process
func (s *processService) Suspend(ctx context.Context, processID string, reason string) error {
	process, err := s.processRepo.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err := handleEntError(ctx, "Process", err); err != nil {
		return err
	}

	_, err = s.processRepo.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			IsSuspended:   true,
			SuspendReason: reason,
		},
	})
	return err
}

// Resume resumes a suspended process
func (s *processService) Resume(ctx context.Context, processID string) error {
	process, err := s.processRepo.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err := handleEntError(ctx, "Process", err); err != nil {
		return err
	}

	_, err = s.processRepo.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			IsSuspended:   false,
			SuspendReason: "",
		},
	})
	return err
}

// Serialization helpers
func (s *processService) serialize(process *ent.Process) *structs.ReadProcess {
	if process == nil {
		return nil
	}

	return &structs.ReadProcess{
		ID:            process.ID,
		ProcessKey:    process.ProcessKey,
		Status:        process.Status,
		TemplateID:    process.TemplateID,
		BusinessKey:   process.BusinessKey,
		ModuleCode:    process.ModuleCode,
		FormCode:      process.FormCode,
		Initiator:     process.Initiator,
		InitiatorDept: process.InitiatorDept,
		ProcessCode:   process.ProcessCode,
		Variables:     process.Variables,
		CurrentNode:   process.CurrentNode,
		ActiveNodes:   process.ActiveNodes,
		FlowStatus:    process.FlowStatus,
		Priority:      process.Priority,
		IsSuspended:   process.IsSuspended,
		SuspendReason: process.SuspendReason,
		AllowCancel:   process.AllowCancel,
		AllowUrge:     process.AllowUrge,
		UrgeCount:     process.UrgeCount,
		Extras:        process.Extras,
		TenantID:      process.TenantID,
		StartTime:     &process.StartTime,
		EndTime:       process.EndTime,
		DueDate:       process.DueTime,
		Duration:      &process.Duration,
		CreatedBy:     &process.CreatedBy,
		CreatedAt:     &process.CreatedAt,
		UpdatedBy:     &process.UpdatedBy,
		UpdatedAt:     &process.UpdatedAt,
	}
}

func (s *processService) serializes(processes []*ent.Process) []*structs.ReadProcess {
	result := make([]*structs.ReadProcess, len(processes))
	for i, process := range processes {
		result[i] = s.serialize(process)
	}
	return result
}
