package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/extension"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/core/workflow/data/ent"
	"ncobase/core/workflow/data/repository"
	"ncobase/core/workflow/structs"
)

type TaskServiceInterface interface {
	Create(ctx context.Context, body *structs.TaskBody) (*structs.ReadTask, error)
	Get(ctx context.Context, params *structs.FindTaskParams) (*structs.ReadTask, error)
	Update(ctx context.Context, body *structs.UpdateTaskBody) (*structs.ReadTask, error)
	Delete(ctx context.Context, params *structs.FindTaskParams) error
	List(ctx context.Context, params *structs.ListTaskParams) (paging.Result[*structs.ReadTask], error)

	Complete(ctx context.Context, req *structs.CompleteTaskRequest) (*structs.CompleteTaskResponse, error)
	Delegate(ctx context.Context, req *structs.DelegateTaskRequest) error
	Transfer(ctx context.Context, req *structs.TransferTaskRequest) error
	Withdraw(ctx context.Context, req *structs.WithdrawTaskRequest) error
	Urge(ctx context.Context, req *structs.UrgeTaskRequest) error
	Claim(ctx context.Context, taskID string, assignees *types.StringArray) error
}

type taskService struct {
	taskRepo    repository.TaskRepositoryInterface
	historyRepo repository.HistoryRepositoryInterface
	nodeRepo    repository.NodeRepositoryInterface
	em          *extension.Manager
}

func NewTaskService(repo repository.Repository, em *extension.Manager) TaskServiceInterface {
	return &taskService{
		taskRepo:    repo.GetTask(),
		historyRepo: repo.GetHistory(),
		nodeRepo:    repo.GetNode(),
		em:          em,
	}
}

// Create creates a new task
func (s *taskService) Create(ctx context.Context, body *structs.TaskBody) (*structs.ReadTask, error) {
	if body.ProcessID == "" {
		return nil, errors.New(ecode.FieldIsRequired("process_id"))
	}
	if body.NodeKey == "" {
		return nil, errors.New(ecode.FieldIsRequired("node_key"))
	}

	task, err := s.taskRepo.Create(ctx, body)
	if err := handleEntError(ctx, "Task", err); err != nil {
		return nil, err
	}

	// Publish event
	s.em.PublishEvent(structs.EventTaskCreated, &structs.EventData{
		TaskID:    task.ID,
		TaskName:  task.Name,
		ProcessID: task.ProcessID,
		NodeID:    task.NodeKey,
		NodeType:  structs.NodeType(task.NodeType),
		Assignees: task.Assignees,
	})

	return s.serialize(task), nil
}

// Complete completes a task
func (s *taskService) Complete(ctx context.Context, req *structs.CompleteTaskRequest) (*structs.CompleteTaskResponse, error) {
	// Get task
	task, err := s.taskRepo.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err := handleEntError(ctx, "Task", err); err != nil {
		return nil, err
	}

	// Verify operator
	for _, assignee := range task.Assignees {
		if assignee == req.Operator {
			break
		}
		return nil, errors.New("task assignee mismatch")
	}

	// Update task
	updatedTask, err := s.taskRepo.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Status:  string(structs.StatusCompleted),
			Action:  string(req.Action),
			Comment: req.Comment,
		},
	})
	if err != nil {
		return nil, err
	}

	// Create history record
	_, err = s.historyRepo.Create(ctx, &structs.HistoryBody{
		Type:      "task",
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  task.NodeType,
		Operator:  req.Operator,
		Action:    string(req.Action),
		Comment:   req.Comment,
		Variables: req.Variables,
		FormData:  req.FormData,
	})
	if err != nil {
		log.Errorf(ctx, "Error creating task completion history: %v", err)
	}

	// Get next nodes
	node, err := s.nodeRepo.Get(ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeKey,
	})
	if err != nil {
		return nil, err
	}

	return &structs.CompleteTaskResponse{
		TaskID:    task.ID,
		ProcessID: task.ProcessID,
		Action:    req.Action,
		EndTime:   updatedTask.EndTime,
		NextNodes: node.NextNodes,
	}, nil
}

// Delegate delegates a task to another user
func (s *taskService) Delegate(ctx context.Context, req *structs.DelegateTaskRequest) error {
	// Get task
	task, err := s.taskRepo.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err := handleEntError(ctx, "Task", err); err != nil {
		return err
	}

	// Verify delegator
	for _, assignee := range task.Assignees {
		if assignee == req.Delegator {
			break
		}
		return errors.New("task delegator mismatch")
	}

	// Update task
	_, err = s.taskRepo.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Assignees:     []string{req.Delegate},
			IsDelegated:   true,
			DelegatedFrom: types.JSON{"delegate": req.Delegator},
		},
	})
	if err != nil {
		return err
	}

	// Create history record
	_, err = s.historyRepo.Create(ctx, &structs.HistoryBody{
		Type:      "task",
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  task.NodeType,
		Operator:  req.Delegator,
		Action:    string(structs.ActionDelegate),
		Comment:   req.Comment,
		Details: map[string]any{
			"delegate_to": req.Delegate,
			"reason":      req.Reason,
		},
	})
	if err != nil {
		log.Errorf(ctx, "Error creating task delegation history: %v", err)
	}

	// Publish event
	s.em.PublishEvent(string(structs.EventTaskDelegated), &structs.EventData{
		TaskID:    task.ID,
		TaskName:  task.Name,
		ProcessID: task.ProcessID,
		NodeID:    task.NodeKey,
		Operator:  req.Delegator,
		Action:    structs.ActionDelegate,
	})

	return nil
}

// Transfer transfers a task to another user
func (s *taskService) Transfer(ctx context.Context, req *structs.TransferTaskRequest) error {
	// Get task
	task, err := s.taskRepo.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err := handleEntError(ctx, "Task", err); err != nil {
		return err
	}

	// Verify transferor
	for _, assignee := range task.Assignees {
		if assignee == req.Transferor {
			break
		}
		return errors.New("task transferor mismatch")
	}

	// Update task
	_, err = s.taskRepo.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Assignees:     types.StringArray{req.Transferee},
			IsTransferred: true,
		},
	})
	if err != nil {
		return err
	}

	// Create history record
	_, err = s.historyRepo.Create(ctx, &structs.HistoryBody{
		Type:      "task",
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  task.NodeType,
		Operator:  req.Transferor,
		Action:    string(structs.ActionTransfer),
		Comment:   req.Comment,
		Details: map[string]any{
			"transfer_to": req.Transferee,
			"reason":      req.Reason,
		},
	})
	if err != nil {
		log.Errorf(ctx, "Error creating task transfer history: %v", err)
	}

	// Publish event
	s.em.PublishEvent(string(structs.EventTaskTransferred), &structs.EventData{
		TaskID:    task.ID,
		TaskName:  task.Name,
		ProcessID: task.ProcessID,
		NodeID:    task.NodeKey,
		Operator:  req.Transferor,
		Action:    structs.ActionTransfer,
	})

	return nil
}

// Withdraw withdraws a task
func (s *taskService) Withdraw(ctx context.Context, req *structs.WithdrawTaskRequest) error {
	// Get task
	task, err := s.taskRepo.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err := handleEntError(ctx, "Task", err); err != nil {
		return err
	}

	// Update task status
	_, err = s.taskRepo.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Status:     string(structs.StatusPending),
			IsResubmit: true,
		},
	})
	if err != nil {
		return err
	}

	// Create history record
	_, err = s.historyRepo.Create(ctx, &structs.HistoryBody{
		Type:      "task",
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  task.NodeType,
		Operator:  req.Operator,
		Action:    string(structs.ActionWithdraw),
		Comment:   req.Comment,
		Details: map[string]any{
			"reason": req.Reason,
		},
	})
	if err != nil {
		log.Errorf(ctx, "Error creating task withdrawal history: %v", err)
	}

	return nil
}

// Urge urges a task
func (s *taskService) Urge(ctx context.Context, req *structs.UrgeTaskRequest) error {
	// Get task
	task, err := s.taskRepo.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err := handleEntError(ctx, "Task", err); err != nil {
		return err
	}

	// Update urge count
	err = s.taskRepo.IncreaseUrgeCount(ctx, task.ID)
	if err != nil {
		return err
	}

	// Create history record
	_, err = s.historyRepo.Create(ctx, &structs.HistoryBody{
		Type:      "task",
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  task.NodeType,
		Operator:  req.Operator,
		Action:    string(structs.ActionUrge),
		Comment:   req.Comment,
		Variables: req.Variables,
	})
	if err != nil {
		log.Errorf(ctx, "Error creating task urge history: %v", err)
	}

	// Publish event
	s.em.PublishEvent(string(structs.EventTaskUrged), &structs.EventData{
		TaskID:    task.ID,
		TaskName:  task.Name,
		ProcessID: task.ProcessID,
		NodeID:    task.NodeKey,
		Operator:  req.Operator,
		Action:    structs.ActionUrge,
	})

	return nil
}

// Claim claims a task
func (s *taskService) Claim(ctx context.Context, taskID string, assignees *types.StringArray) error {
	// Get task
	task, err := s.taskRepo.Get(ctx, &structs.FindTaskParams{
		ProcessID: taskID,
	})
	if err := handleEntError(ctx, "Task", err); err != nil {
		return err
	}

	// Verify task is unassigned
	if len(task.Assignees) > 0 {

		return errors.New("task already assigned")
	}

	// Update task assignee
	_, err = s.taskRepo.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Assignees: *assignees,
			ClaimTime: task.ClaimTime,
		},
	})

	return err
}

// Get retrieves a specific task
func (s *taskService) Get(ctx context.Context, params *structs.FindTaskParams) (*structs.ReadTask, error) {
	task, err := s.taskRepo.Get(ctx, params)
	if err := handleEntError(ctx, "Task", err); err != nil {
		return nil, err
	}

	return s.serialize(task), nil
}

// Update updates an existing task
func (s *taskService) Update(ctx context.Context, body *structs.UpdateTaskBody) (*structs.ReadTask, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	task, err := s.taskRepo.Update(ctx, body)
	if err := handleEntError(ctx, "Task", err); err != nil {
		return nil, err
	}

	return s.serialize(task), nil
}

// Delete deletes a task
func (s *taskService) Delete(ctx context.Context, params *structs.FindTaskParams) error {
	return handleEntError(ctx, "Task", s.taskRepo.Delete(ctx, params))
}

// List returns a list of tasks
func (s *taskService) List(ctx context.Context, params *structs.ListTaskParams) (paging.Result[*structs.ReadTask], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTask, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		tasks, err := s.taskRepo.List(ctx, &lp)
		if err != nil {
			log.Errorf(ctx, "Error listing tasks: %v", err)
			return nil, 0, err
		}

		total := s.taskRepo.CountX(ctx, params)

		return s.serializes(tasks), total, nil
	})
}

// Serialization helpers
func (s *taskService) serialize(task *ent.Task) *structs.ReadTask {
	if task == nil {
		return nil
	}

	return &structs.ReadTask{
		ID:          task.ID,
		Name:        task.Name,
		Description: task.Description,
		Status:      task.Status,
		TaskKey:     task.TaskKey,
		ProcessID:   task.ProcessID,
		NodeKey:     task.NodeKey,
		NodeType:    task.NodeType,
		Assignees:   task.Assignees,
		Candidates:  task.Candidates,
		Action:      task.Action,
		Comment:     task.Comment,
		FormData:    task.FormData,
		Variables:   task.Variables,
		Priority:    task.Priority,
		IsUrged:     task.IsUrged,
		UrgeCount:   task.UrgeCount,
		IsTimeout:   task.IsTimeout,
		Extras:      task.Extras,
		TenantID:    task.TenantID,
		StartTime:   &task.StartTime,
		ClaimTime:   task.ClaimTime,
		EndTime:     task.EndTime,
		DueTime:     task.DueTime,
		Duration:    &task.Duration,
		CreatedBy:   &task.CreatedBy,
		CreatedAt:   &task.CreatedAt,
		UpdatedBy:   &task.UpdatedBy,
		UpdatedAt:   &task.UpdatedAt,
	}
}

func (s *taskService) serializes(tasks []*ent.Task) []*structs.ReadTask {
	result := make([]*structs.ReadTask, len(tasks))
	for i, task := range tasks {
		result[i] = s.serialize(task)
	}
	return result
}
