package repository

import (
	"context"
	"fmt"
	"ncobase/core/workflow/data"
	"ncobase/core/workflow/data/ent"
	taskEnt "ncobase/core/workflow/data/ent/task"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

type TaskRepositoryInterface interface {
	Create(context.Context, *structs.TaskBody) (*ent.Task, error)
	Get(context.Context, *structs.FindTaskParams) (*ent.Task, error)
	Update(context.Context, *structs.UpdateTaskBody) (*ent.Task, error)
	Delete(context.Context, *structs.FindTaskParams) error
	List(context.Context, *structs.ListTaskParams) ([]*ent.Task, error)
	CountX(context.Context, *structs.ListTaskParams) int

	ListByCondition(ctx context.Context, cond *structs.TaskCondition) ([]*ent.Task, error)
	CountByStatus(ctx context.Context, processID string) (map[string]int, error)
	GetTaskChain(ctx context.Context, taskID string) ([]*ent.Task, error)

	AddChildTask(ctx context.Context, parentID string, childID string) error
	RemoveChildTask(ctx context.Context, parentID string, childID string) error

	UpdateStatus(context.Context, string, string) error
	UpdateAssignee(context.Context, string, []string) error
	IncreaseUrgeCount(context.Context, string) error
	MarkTimeout(context.Context, string) error
}

type taskRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Task]
}

func NewTaskRepository(d *data.Data) TaskRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &taskRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Task](rc, "workflow_task", false),
	}
}

// Create creates a new task
func (r *taskRepository) Create(ctx context.Context, body *structs.TaskBody) (*ent.Task, error) {
	builder := r.ec.Task.Create()

	if body.Name != "" {
		builder.SetName(body.Name)
	}
	if body.Description != "" {
		builder.SetDescription(body.Description)
	}

	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}

	if body.TaskKey != "" {
		builder.SetTaskKey(body.TaskKey)
	}
	if body.ProcessID != "" {
		builder.SetProcessID(body.ProcessID)
	}
	if body.NodeKey != "" {
		builder.SetNodeKey(body.NodeKey)
	}
	if body.NodeType != "" {
		builder.SetNodeType(body.NodeType)
	}
	if validator.IsEmpty(body.Assignees) {
		builder.SetAssignees(body.Assignees)
	}
	if validator.IsEmpty(body.Candidates) {
		builder.SetCandidates(body.Candidates)
	}

	builder.SetPriority(body.Priority)
	builder.SetIsUrged(body.IsUrged)
	builder.SetUrgeCount(body.UrgeCount)
	builder.SetIsTimeout(body.IsTimeout)

	if body.FormData != nil {
		builder.SetFormData(body.FormData)
	}
	if body.Variables != nil {
		builder.SetVariables(body.Variables)
	}
	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "taskRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("tasks", row); err != nil {
		logger.Errorf(ctx, "taskRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific task
func (r *taskRepository) Get(ctx context.Context, params *structs.FindTaskParams) (*ent.Task, error) {
	builder := r.ec.Task.Query()

	if params.ProcessID != "" {
		builder.Where(taskEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.NodeKey != "" {
		builder.Where(taskEnt.NodeKeyEQ(params.NodeKey))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(taskEnt.StatusEQ(params.Status))
	}

	row, err := builder.First(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// Update updates an existing task
func (r *taskRepository) Update(ctx context.Context, body *structs.UpdateTaskBody) (*ent.Task, error) {
	builder := r.ec.Task.UpdateOneID(body.ID)

	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.Action != "" {
		builder.SetAction(body.Action)
	}
	if body.Comment != "" {
		builder.SetComment(body.Comment)
	}
	if body.FormData != nil {
		builder.SetFormData(body.FormData)
	}
	if body.Variables != nil {
		builder.SetVariables(body.Variables)
	}

	builder.SetIsUrged(body.IsUrged)
	builder.SetUrgeCount(body.UrgeCount)
	builder.SetIsTimeout(body.IsTimeout)

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	return builder.Save(ctx)
}

// Delete deletes a task
func (r *taskRepository) Delete(ctx context.Context, params *structs.FindTaskParams) error {
	builder := r.ec.Task.Delete()

	if params.ProcessID != "" {
		builder.Where(taskEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.NodeKey != "" {
		builder.Where(taskEnt.NodeKeyEQ(params.NodeKey))
	}

	_, err := builder.Exec(ctx)
	return err
}

// List returns a list of tasks
func (r *taskRepository) List(ctx context.Context, params *structs.ListTaskParams) ([]*ent.Task, error) {
	builder := r.ec.Task.Query()

	if params.ProcessID != "" {
		builder.Where(taskEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.NodeType != "" {
		builder.Where(taskEnt.NodeTypeEQ(params.NodeType))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(taskEnt.StatusEQ(params.Status))
	}
	if params.IsTimeout != nil {
		builder.Where(taskEnt.IsTimeoutEQ(*params.IsTimeout))
	}
	if params.Priority != nil {
		builder.Where(taskEnt.PriorityEQ(*params.Priority))
	}

	// Add sorting
	switch params.SortBy {
	case structs.SortByCreatedAt:
		builder.Order(ent.Desc(taskEnt.FieldCreatedAt))
	case structs.SortByPriority:
		builder.Order(ent.Desc(taskEnt.FieldPriority))
	case structs.SortByDueTime:
		builder.Order(ent.Asc(taskEnt.FieldDueTime))
	default:
		builder.Order(ent.Desc(taskEnt.FieldCreatedAt))
	}

	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// CountX returns the total count of tasks
func (r *taskRepository) CountX(ctx context.Context, params *structs.ListTaskParams) int {
	builder := r.ec.Task.Query()

	if params.ProcessID != "" {
		builder.Where(taskEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.NodeType != "" {
		builder.Where(taskEnt.NodeTypeEQ(params.NodeType))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(taskEnt.StatusEQ(params.Status))
	}

	return builder.CountX(ctx)
}

// ListByCondition returns a list of tasks
func (r *taskRepository) ListByCondition(ctx context.Context, cond *structs.TaskCondition) ([]*ent.Task, error) {
	builder := r.ec.Task.Query()

	if len(cond.ProcessIDs) > 0 {
		builder.Where(taskEnt.ProcessIDIn(cond.ProcessIDs...))
	}
	if len(cond.NodeTypes) > 0 {
		builder.Where(taskEnt.NodeTypeIn(cond.NodeTypes...))
	}
	if len(cond.Statuses) > 0 {
		builder.Where(taskEnt.StatusIn(cond.Statuses...))
	}
	if cond.Priority != nil {
		builder.Where(taskEnt.PriorityEQ(*cond.Priority))
	}
	if cond.IsTimeout != nil {
		builder.Where(taskEnt.IsTimeoutEQ(*cond.IsTimeout))
	}
	if cond.IsUrged != nil {
		builder.Where(taskEnt.IsUrgedEQ(*cond.IsUrged))
	}

	// Time range
	if cond.StartTime != nil {
		builder.Where(taskEnt.CreatedAtGTE(*cond.StartTime))
	}
	if cond.EndTime != nil {
		builder.Where(taskEnt.CreatedAtLTE(*cond.EndTime))
	}

	// Assignees
	if len(cond.Assignees) > 0 {
		// builder.Where(taskEnt(
		// 	func(s *ent.TaskQuery) {
		// 		s.Where(taskEnt.IDIn(cond.Assignees...))
		// 	},
		// ))
	}

	// Sorting
	for i, field := range cond.SortBy {
		order := "asc"
		if i < len(cond.OrderBy) {
			order = cond.OrderBy[i]
		}
		switch field {
		case "created_at":
			if order == "desc" {
				builder.Order(ent.Desc(taskEnt.FieldCreatedAt))
			} else {
				builder.Order(ent.Asc(taskEnt.FieldCreatedAt))
			}
		case "priority":
			if order == "desc" {
				builder.Order(ent.Desc(taskEnt.FieldPriority))
			} else {
				builder.Order(ent.Asc(taskEnt.FieldPriority))
			}
		}
	}

	// Pagination
	if cond.Limit > 0 {
		builder.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		builder.Offset(cond.Offset)
	}

	return builder.All(ctx)
}

// CountByStatus counts the number of tasks by status
func (r *taskRepository) CountByStatus(ctx context.Context, processID string) (map[string]int, error) {
	result := make(map[string]int)
	// get all status
	statuses := []string{
		string(structs.StatusDraft),
		string(structs.StatusPending),
		string(structs.StatusProcessing),
		string(structs.StatusCompleted),
		string(structs.StatusRejected),
		string(structs.StatusCancelled),
		string(structs.StatusWithdrawn),
	}

	// get count by status
	for _, status := range statuses {
		count, err := r.ec.Task.Query().
			Where(
				taskEnt.ProcessIDEQ(processID),
				taskEnt.StatusEQ(status),
			).
			Count(ctx)

		if err != nil {
			logger.Warnf(ctx, "Failed to count tasks with status %s: %v", status, err)
			continue
		}

		if count > 0 {
			result[status] = count
		}
	}

	// get total
	total, err := r.ec.Task.Query().
		Where(taskEnt.ProcessIDEQ(processID)).
		Count(ctx)

	if err != nil {
		logger.Warnf(ctx, "Failed to count total tasks: %v", err)
	} else {
		result["total"] = total
	}

	return result, nil
}

// GetTaskChain gets the task chain
func (r *taskRepository) GetTaskChain(ctx context.Context, taskID string) ([]*ent.Task, error) {
	var tasks []*ent.Task

	// current task
	task, err := r.ec.Task.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// get parent tasks
	if task.ParentID != "" {
		parentChain, err := r.getParentChain(ctx, task.ParentID)
		if err != nil {
			logger.Warnf(ctx, "Failed to get parent chain: %v", err)
		} else {
			tasks = append(tasks, parentChain...)
		}
	}

	// add current task
	tasks = append(tasks, task)

	// get child tasks
	childChain, err := r.getChildChain(ctx, task.ChildIds)
	if err != nil {
		logger.Warnf(ctx, "Failed to get child chain: %v", err)
	} else {
		tasks = append(tasks, childChain...)
	}

	return tasks, nil
}

// getParentChain gets the parent task chain
func (r *taskRepository) getParentChain(ctx context.Context, parentID string) ([]*ent.Task, error) {
	var chain []*ent.Task

	for parentID != "" {
		parent, err := r.ec.Task.Get(ctx, parentID)
		if err != nil {
			if ent.IsNotFound(err) {
				break
			}
			return nil, err
		}

		chain = append([]*ent.Task{parent}, chain...) // prepend
		parentID = parent.ParentID
	}

	return chain, nil
}

// getChildChain gets the child task chain
func (r *taskRepository) getChildChain(ctx context.Context, childIDs []string) ([]*ent.Task, error) {
	if len(childIDs) == 0 {
		return nil, nil
	}

	var chain []*ent.Task

	// get child tasks
	children, err := r.ec.Task.Query().
		Where(taskEnt.IDIn(childIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	chain = append(chain, children...)

	// Recursively get child tasks
	for _, child := range children {
		childChain, err := r.getChildChain(ctx, child.ChildIds)
		if err != nil {
			return nil, err
		}
		chain = append(chain, childChain...)
	}

	return chain, nil
}

// AddChildTask Adds a child task
func (r *taskRepository) AddChildTask(ctx context.Context, parentID string, childID string) error {
	// Get parent task
	parent, err := r.ec.Task.Get(ctx, parentID)
	if err != nil {
		return err
	}

	// Update child task parent
	err = r.ec.Task.UpdateOneID(childID).
		SetParentID(parentID).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Update parent task
	childIDs := append(parent.ChildIds, childID)
	return r.ec.Task.UpdateOneID(parentID).
		SetChildIds(childIDs).
		Exec(ctx)
}

// RemoveChildTask removes a child task
func (r *taskRepository) RemoveChildTask(ctx context.Context, parentID string, childID string) error {
	tx, err := r.ec.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// get parent
	parent, err := tx.Task.Get(ctx, parentID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// remove child
	var newChildIDs []string
	for _, id := range parent.ChildIds {
		if id != childID {
			newChildIDs = append(newChildIDs, id)
		}
	}

	// update parent
	err = tx.Task.UpdateOneID(parentID).
		SetChildIds(newChildIDs).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// clear child task parent
	err = tx.Task.UpdateOneID(childID).
		ClearParentID().
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateStatus updates task status
func (r *taskRepository) UpdateStatus(ctx context.Context, taskID string, status string) error {
	return r.ec.Task.UpdateOneID(taskID).
		SetStatus(status).
		Exec(ctx)
}

// UpdateAssignee updates task assignee
func (r *taskRepository) UpdateAssignee(ctx context.Context, taskID string, assignees []string) error {
	builder := r.ec.Task.UpdateOneID(taskID)

	if len(assignees) > 0 {
		builder.SetAssignees(assignees)
	}

	return builder.Exec(ctx)
}

// IncreaseUrgeCount increases the urge count
func (r *taskRepository) IncreaseUrgeCount(ctx context.Context, taskID string) error {
	return r.ec.Task.UpdateOneID(taskID).
		AddUrgeCount(1).
		SetIsUrged(true).
		Exec(ctx)
}

// MarkTimeout marks a task as timeout
func (r *taskRepository) MarkTimeout(ctx context.Context, taskID string) error {
	return r.ec.Task.UpdateOneID(taskID).
		SetIsTimeout(true).
		Exec(ctx)
}
