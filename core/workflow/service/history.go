package service

import (
	"context"
	"ncobase/core/workflow/data/ent"
	"ncobase/core/workflow/data/repository"
	"ncobase/core/workflow/structs"
	"ncore/extension"
	"ncore/pkg/logger"
	"ncore/pkg/paging"
)

type HistoryServiceInterface interface {
	Create(ctx context.Context, body *structs.HistoryBody) (*structs.ReadHistory, error)
	Get(ctx context.Context, params *structs.FindHistoryParams) (*structs.ReadHistory, error)
	List(ctx context.Context, params *structs.ListHistoryParams) (paging.Result[*structs.ReadHistory], error)

	// History specific operations

	GetProcessHistory(ctx context.Context, processID string) ([]*structs.ReadHistory, error)
	GetTaskHistory(ctx context.Context, taskID string) ([]*structs.ReadHistory, error)
	GetOperatorHistory(ctx context.Context, operator string) ([]*structs.ReadHistory, error)
}

type historyService struct {
	historyRepo repository.HistoryRepositoryInterface
	em          *extension.Manager
}

func NewHistoryService(repo repository.Repository, em *extension.Manager) HistoryServiceInterface {
	return &historyService{
		historyRepo: repo.GetHistory(),
		em:          em,
	}
}

// Create creates a new history record
func (s *historyService) Create(ctx context.Context, body *structs.HistoryBody) (*structs.ReadHistory, error) {
	history, err := s.historyRepo.Create(ctx, body)
	if err := handleEntError(ctx, "History", err); err != nil {
		return nil, err
	}

	// 发布历史记录事件
	s.publishHistoryEvent(ctx, history)

	return s.serialize(history), nil
}

// Get retrieves a specific history record
func (s *historyService) Get(ctx context.Context, params *structs.FindHistoryParams) (*structs.ReadHistory, error) {
	history, err := s.historyRepo.Get(ctx, params)
	if err := handleEntError(ctx, "History", err); err != nil {
		return nil, err
	}

	return s.serialize(history), nil
}

// List returns a list of history records
func (s *historyService) List(ctx context.Context, params *structs.ListHistoryParams) (paging.Result[*structs.ReadHistory], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadHistory, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		histories, err := s.historyRepo.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing histories: %v", err)
			return nil, 0, err
		}

		total := s.historyRepo.CountX(ctx, params)

		return s.serializes(histories), total, nil
	})
}

// GetProcessHistory returns all history records for a process
func (s *historyService) GetProcessHistory(ctx context.Context, processID string) ([]*structs.ReadHistory, error) {
	histories, err := s.historyRepo.List(ctx, &structs.ListHistoryParams{
		ProcessID: processID,
		Limit:     1000, // 使用较大的限制以获取所有记录
	})
	if err != nil {
		return nil, err
	}

	return s.serializes(histories), nil
}

// GetTaskHistory returns all history records for a task
func (s *historyService) GetTaskHistory(ctx context.Context, taskID string) ([]*structs.ReadHistory, error) {
	histories, err := s.historyRepo.List(ctx, &structs.ListHistoryParams{
		TaskID: taskID,
		Limit:  1000,
	})
	if err != nil {
		return nil, err
	}

	return s.serializes(histories), nil
}

// GetOperatorHistory returns all history records for an operator
func (s *historyService) GetOperatorHistory(ctx context.Context, operator string) ([]*structs.ReadHistory, error) {
	histories, err := s.historyRepo.List(ctx, &structs.ListHistoryParams{
		Operator: operator,
		Limit:    1000,
	})
	if err != nil {
		return nil, err
	}

	return s.serializes(histories), nil
}

// Internal helper methods

// publishHistoryEvent publishes events based on history type
func (s *historyService) publishHistoryEvent(ctx context.Context, history *ent.History) {
	var eventType string

	switch history.Action {
	case string(structs.ActionSubmit):
		eventType = string(structs.EventProcessStarted)
	case string(structs.ActionApprove):
		eventType = string(structs.EventTaskCompleted)
	case string(structs.ActionReject):
		eventType = string(structs.EventTaskCompleted)
	case string(structs.ActionDelegate):
		eventType = string(structs.EventTaskDelegated)
	case string(structs.ActionTransfer):
		eventType = string(structs.EventTaskTransferred)
	case string(structs.ActionUrge):
		eventType = string(structs.EventTaskUrged)
	default:
		return
	}

	eventData := &structs.EventData{
		ProcessID: history.ProcessID,
		NodeID:    history.NodeKey,
		NodeName:  history.NodeName,
		NodeType:  structs.NodeType(history.NodeType),
		TaskID:    history.TaskID,
		Operator:  history.Operator,
		Action:    structs.ActionType(history.Action),
		Variables: history.Variables,
		Timestamp: history.CreatedAt,
	}

	s.em.PublishEvent(eventType, eventData)
}

// enrichHistoryData adds additional context to history records
func (s *historyService) enrichHistoryData(ctx context.Context, history *ent.History) error {
	// 可以根据需要添加额外的上下文信息
	// 例如：流程名称、节点标签等
	return nil
}

// Serialization helpers
func (s *historyService) serialize(history *ent.History) *structs.ReadHistory {
	if history == nil {
		return nil
	}

	return &structs.ReadHistory{
		ID:           history.ID,
		Type:         history.Type,
		ProcessID:    history.ProcessID,
		TemplateID:   history.TemplateID,
		NodeID:       history.NodeKey,
		NodeName:     history.NodeName,
		NodeType:     history.NodeType,
		TaskID:       history.TaskID,
		Operator:     history.Operator,
		OperatorDept: history.OperatorDept,
		Action:       history.Action,
		Comment:      history.Comment,
		Variables:    history.Variables,
		FormData:     history.FormData,
		NodeConfig:   history.NodeConfig,
		Details:      history.Details,
		TenantID:     history.TenantID,
		CreatedBy:    &history.CreatedBy,
		CreatedAt:    &history.CreatedAt,
		UpdatedBy:    &history.UpdatedBy,
		UpdatedAt:    &history.UpdatedAt,
	}
}

func (s *historyService) serializes(histories []*ent.History) []*structs.ReadHistory {
	result := make([]*structs.ReadHistory, len(histories))
	for i, history := range histories {
		result[i] = s.serialize(history)
	}
	return result
}
