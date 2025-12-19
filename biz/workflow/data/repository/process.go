package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/workflow/data"
	"ncobase/workflow/data/ent"
	businessEnt "ncobase/workflow/data/ent/business"
	nodeEnt "ncobase/workflow/data/ent/node"
	processEnt "ncobase/workflow/data/ent/process"
	taskEnt "ncobase/workflow/data/ent/task"
	"ncobase/workflow/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/search"
)

type ProcessRepositoryInterface interface {
	Create(context.Context, *structs.ProcessBody) (*ent.Process, error)
	Get(context.Context, *structs.FindProcessParams) (*ent.Process, error)
	Update(context.Context, *structs.UpdateProcessBody) (*ent.Process, error)
	Delete(context.Context, *structs.FindProcessParams) error
	List(context.Context, *structs.ListProcessParams) ([]*ent.Process, error)

	CreateSnapshot(ctx context.Context, processID string) error
	RestoreSnapshot(ctx context.Context, snapshotID string) error
	CleanupProcesses(ctx context.Context, before int64) error

	CountX(context.Context, *structs.ListProcessParams) int
}

type processRepository struct {
	data *data.Data
	ec   *ent.Client
	rc   *redis.Client
	c    *cache.Cache[ent.Process]
}

func NewProcessRepository(d *data.Data) ProcessRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &processRepository{
		data: d,
		ec:   ec,
		rc:   rc,
		c:    cache.NewCache[ent.Process](rc, "workflow_process", false),
	}
}

// Create creates a new process
func (r *processRepository) Create(ctx context.Context, body *structs.ProcessBody) (*ent.Process, error) {
	builder := r.ec.Process.Create()

	// Set process fields
	if body.ProcessKey != "" {
		builder.SetProcessKey(body.ProcessKey)
	}

	builder.SetStatus(body.Status)

	if body.TemplateID != "" {
		builder.SetTemplateID(body.TemplateID)
	}
	if body.BusinessKey != "" {
		builder.SetBusinessKey(body.BusinessKey)
	}
	if body.ModuleCode != "" {
		builder.SetModuleCode(body.ModuleCode)
	}
	if body.FormCode != "" {
		builder.SetFormCode(body.FormCode)
	}
	if body.Initiator != "" {
		builder.SetInitiator(body.Initiator)
	}
	if body.InitiatorDept != "" {
		builder.SetInitiatorDept(body.InitiatorDept)
	}
	if body.ProcessCode != "" {
		builder.SetProcessCode(body.ProcessCode)
	}

	if body.Variables != nil {
		builder.SetVariables(body.Variables)
	}
	if body.CurrentNode != "" {
		builder.SetCurrentNode(body.CurrentNode)
	}
	if len(body.ActiveNodes) > 0 {
		builder.SetActiveNodes(body.ActiveNodes)
	}
	if body.FlowStatus != "" {
		builder.SetFlowStatus(body.FlowStatus)
	}

	builder.SetPriority(body.Priority)
	builder.SetIsSuspended(body.IsSuspended)
	builder.SetSuspendReason(body.SuspendReason)
	builder.SetAllowCancel(body.AllowCancel)
	builder.SetAllowUrge(body.AllowUrge)
	builder.SetUrgeCount(body.UrgeCount)

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "processRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "processes", Document: row}); err != nil {
		logger.Errorf(ctx, "processRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific process
func (r *processRepository) Get(ctx context.Context, params *structs.FindProcessParams) (*ent.Process, error) {
	builder := r.ec.Process.Query()

	if params.ProcessKey != "" {
		builder.Where(processEnt.ProcessKeyEQ(params.ProcessKey))
	}
	if params.TemplateID != "" {
		builder.Where(processEnt.TemplateIDEQ(params.TemplateID))
	}
	if params.BusinessKey != "" {
		builder.Where(processEnt.BusinessKeyEQ(params.BusinessKey))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(processEnt.StatusEQ(params.Status))
	}
	if params.Initiator != "" {
		builder.Where(processEnt.InitiatorEQ(params.Initiator))
	}

	row, err := builder.First(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// Update updates an existing process
func (r *processRepository) Update(ctx context.Context, body *structs.UpdateProcessBody) (*ent.Process, error) {
	builder := r.ec.Process.UpdateOneID(body.ID)

	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.CurrentNode != "" {
		builder.SetCurrentNode(body.CurrentNode)
	}
	if len(body.ActiveNodes) > 0 {
		builder.SetActiveNodes(body.ActiveNodes)
	}
	if body.FlowStatus != "" {
		builder.SetFlowStatus(body.FlowStatus)
	}
	if body.Variables != nil {
		builder.SetVariables(body.Variables)
	}

	builder.SetPriority(body.Priority)
	builder.SetIsSuspended(body.IsSuspended)
	builder.SetSuspendReason(body.SuspendReason)
	builder.SetUrgeCount(body.UrgeCount)

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "processRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "processes", Document: row, DocumentID: row.ID}); err != nil {
		logger.Errorf(ctx, "processRepo.Update error updating Meilisearch index: %v", err)
	}

	return row, nil
}

// Delete deletes a process
func (r *processRepository) Delete(ctx context.Context, params *structs.FindProcessParams) error {
	builder := r.ec.Process.Delete()

	var targetID string
	if params.ProcessKey != "" {
		if row, err := r.ec.Process.Query().Where(processEnt.ProcessKeyEQ(params.ProcessKey)).First(ctx); err == nil && row != nil {
			targetID = row.ID
		}
		builder.Where(processEnt.ProcessKeyEQ(params.ProcessKey))
	}

	_, err := builder.Exec(ctx)
	if err != nil {
		return err
	}

	// Delete from Meilisearch
	if targetID != "" {
		if err = r.data.DeleteDocument(ctx, "processes", targetID); err != nil {
			logger.Errorf(ctx, "processRepo.Delete error deleting Meilisearch index: %v", err)
		}
	}

	return nil
}

// List returns a list of processes
func (r *processRepository) List(ctx context.Context, params *structs.ListProcessParams) ([]*ent.Process, error) {
	builder := r.ec.Process.Query()

	if validator.IsNotEmpty(params.Status) {
		builder.Where(processEnt.StatusEQ(params.Status))
	}
	if params.Initiator != "" {
		builder.Where(processEnt.InitiatorEQ(params.Initiator))
	}
	if params.Priority != nil {
		builder.Where(processEnt.PriorityEQ(*params.Priority))
	}

	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// CreateSnapshot creates a snapshot
func (r *processRepository) CreateSnapshot(ctx context.Context, processID string) error {
	// get process
	process, err := r.ec.Process.Get(ctx, processID)
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	// get nodes
	nodes, err := r.ec.Node.Query().
		Where(nodeEnt.ProcessIDEQ(processID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}

	// get tasks
	tasks, err := r.ec.Task.Query().
		Where(taskEnt.ProcessIDEQ(processID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	// get business
	business, err := r.ec.Business.Query().
		Where(businessEnt.ProcessIDEQ(processID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return fmt.Errorf("failed to get business data: %w", err)
	}

	// marshal
	processData, _ := json.Marshal(process)
	nodeData, _ := json.Marshal(nodes)
	taskData, _ := json.Marshal(tasks)
	var businessData []byte
	if business != nil {
		businessData, _ = json.Marshal(business)
	}

	// create snapshot
	snapshot := &structs.ProcessSnapshot{
		ProcessID:    processID,
		ProcessData:  processData,
		NodeData:     nodeData,
		TaskData:     taskData,
		BusinessData: businessData,
		CreatedAt:    time.Now().UnixMilli(),
		CreatedBy:    "system", // TODO:
	}

	// save snapshot
	err = r.saveSnapshot(ctx, snapshot)
	if err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	logger.Infof(ctx, "Created snapshot for process %s", processID)
	return nil
}

// RestoreSnapshot restores a snapshot
func (r *processRepository) RestoreSnapshot(ctx context.Context, snapshotID string) error {
	// get snapshot
	snapshot, err := r.getSnapshot(ctx, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// start transaction
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

	// remove process
	_, err = tx.Task.Delete().Where(taskEnt.ProcessIDEQ(snapshot.ProcessID)).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete tasks: %w", err)
	}

	_, err = tx.Node.Delete().Where(nodeEnt.ProcessIDEQ(snapshot.ProcessID)).Exec(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete nodes: %w", err)
	}

	// restore process
	var process ent.Process
	if err := json.Unmarshal(snapshot.ProcessData, &process); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unmarshal process data: %w", err)
	}

	// restore nodes
	var nodes []*ent.Node
	if err := json.Unmarshal(snapshot.NodeData, &nodes); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unmarshal node data: %w", err)
	}

	// restore tasks
	var tasks []*ent.Task
	if err := json.Unmarshal(snapshot.TaskData, &tasks); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	// restore business
	if len(snapshot.BusinessData) > 0 {
		var business ent.Business
		if err := json.Unmarshal(snapshot.BusinessData, &business); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to unmarshal business data: %w", err)
		}
	}

	// submit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Infof(ctx, "Restored snapshot %s for process %s", snapshotID, snapshot.ProcessID)
	return nil
}

// CleanupProcesses cleans up processes
func (r *processRepository) CleanupProcesses(ctx context.Context, before int64) error {
	// start transaction
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

	// find needed processes
	processes, err := tx.Process.Query().
		Where(
			processEnt.CreatedAtLT(before),
			processEnt.StatusIn(
				string(structs.StatusCompleted),
				string(structs.StatusTerminated),
				string(structs.StatusCancelled),
				string(structs.StatusRejected),
			),
		).All(ctx)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to query processes: %w", err)
	}

	for _, process := range processes {
		// create snapshot
		if err := r.CreateSnapshot(ctx, process.ID); err != nil {
			logger.Warnf(ctx, "Failed to create snapshot for process %s: %v", process.ID, err)
			continue
		}

		// remove tasks
		_, err = tx.Task.Delete().
			Where(taskEnt.ProcessIDEQ(process.ID)).
			Exec(ctx)
		if err != nil {
			logger.Warnf(ctx, "Failed to delete tasks for process %s: %v", process.ID, err)
		}

		// remove nodes
		_, err = tx.Node.Delete().
			Where(nodeEnt.ProcessIDEQ(process.ID)).
			Exec(ctx)
		if err != nil {
			logger.Warnf(ctx, "Failed to delete nodes for process %s: %v", process.ID, err)
		}

		// remove process itself
		err = tx.Process.DeleteOne(process).Exec(ctx)
		if err != nil {
			logger.Warnf(ctx, "Failed to delete process %s: %v", process.ID, err)
		}
	}

	// submit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	logger.Infof(ctx, "Cleaned up %d processes before %v", len(processes), before)
	return nil
}

// saveSnapshot saves a snapshot
func (r *processRepository) saveSnapshot(ctx context.Context, snapshot *structs.ProcessSnapshot) error {
	// TODO: save snapshot, e.g. to redis / meilisearch / elasticsearch / mongodb / etc
	return nil
}

// getSnapshot gets a snapshot
func (r *processRepository) getSnapshot(ctx context.Context, snapshotID string) (*structs.ProcessSnapshot, error) {
	// TODO: get snapshot
	return nil, nil
}

// CountX returns the total count of processes
func (r *processRepository) CountX(ctx context.Context, params *structs.ListProcessParams) int {
	builder := r.ec.Process.Query()

	if validator.IsNotEmpty(params.Status) {
		builder.Where(processEnt.StatusEQ(params.Status))
	}
	if params.Initiator != "" {
		builder.Where(processEnt.InitiatorEQ(params.Initiator))
	}
	if params.Priority != nil {
		builder.Where(processEnt.PriorityEQ(*params.Priority))
	}

	return builder.CountX(ctx)
}
