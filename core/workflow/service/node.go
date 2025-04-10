package service

import (
	"context"
	"errors"
	"ncobase/core/workflow/data/ent"
	"ncobase/core/workflow/data/repository"
	"ncobase/core/workflow/structs"

	ext "github.com/ncobase/ncore/ext/types"
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/paging"
	"github.com/ncobase/ncore/pkg/validator"
)

type NodeServiceInterface interface {
	Create(ctx context.Context, body *structs.NodeBody) (*structs.ReadNode, error)
	Get(ctx context.Context, params *structs.FindNodeParams) (*structs.ReadNode, error)
	Update(ctx context.Context, body *structs.UpdateNodeBody) (*structs.ReadNode, error)
	Delete(ctx context.Context, params *structs.FindNodeParams) error
	List(ctx context.Context, params *structs.ListNodeParams) (paging.Result[*structs.ReadNode], error)

	// Node specific operations

	UpdateStatus(ctx context.Context, nodeID string, status string) error
	GetProcessNodes(ctx context.Context, processID string) ([]*structs.ReadNode, error)
	ValidateNodeConfig(ctx context.Context, nodeID string) error
}

type nodeService struct {
	nodeRepo repository.NodeRepositoryInterface
	em       ext.ManagerInterface
}

func NewNodeService(repo repository.Repository, em ext.ManagerInterface) NodeServiceInterface {
	return &nodeService{
		nodeRepo: repo.GetNode(),
		em:       em,
	}
}

// Create creates a new node
func (s *nodeService) Create(ctx context.Context, body *structs.NodeBody) (*structs.ReadNode, error) {
	if body.ProcessID == "" {
		return nil, errors.New(ecode.FieldIsRequired("process_id"))
	}

	node, err := s.nodeRepo.Create(ctx, body)
	if err := handleEntError(ctx, "Node", err); err != nil {
		return nil, err
	}

	return s.serialize(node), nil
}

// Get retrieves a node
func (s *nodeService) Get(ctx context.Context, params *structs.FindNodeParams) (*structs.ReadNode, error) {
	node, err := s.nodeRepo.Get(ctx, params)
	if err := handleEntError(ctx, "Node", err); err != nil {
		return nil, err
	}

	return s.serialize(node), nil
}

// Update updates a node
func (s *nodeService) Update(ctx context.Context, body *structs.UpdateNodeBody) (*structs.ReadNode, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	node, err := s.nodeRepo.Update(ctx, body)
	if err := handleEntError(ctx, "Node", err); err != nil {
		return nil, err
	}

	return s.serialize(node), nil
}

// Delete deletes a node
func (s *nodeService) Delete(ctx context.Context, params *structs.FindNodeParams) error {
	return handleEntError(ctx, "Node", s.nodeRepo.Delete(ctx, params))
}

// List returns a list of nodes
func (s *nodeService) List(ctx context.Context, params *structs.ListNodeParams) (paging.Result[*structs.ReadNode], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadNode, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		nodes, err := s.nodeRepo.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing nodes: %v", err)
			return nil, 0, err
		}

		total := s.nodeRepo.CountX(ctx, params)

		return s.serializes(nodes), total, nil
	})
}

// UpdateStatus updates node status
func (s *nodeService) UpdateStatus(ctx context.Context, nodeID string, status string) error {
	return s.nodeRepo.UpdateStatus(ctx, nodeID, status)
}

// GetProcessNodes gets all nodes for a process
func (s *nodeService) GetProcessNodes(ctx context.Context, processID string) ([]*structs.ReadNode, error) {
	nodes, err := s.nodeRepo.GetNodesByProcessID(ctx, processID)
	if err != nil {
		return nil, err
	}
	return s.serializes(nodes), nil
}

// ValidateNodeConfig validates node configuration
func (s *nodeService) ValidateNodeConfig(ctx context.Context, nodeID string) error {
	node, err := s.nodeRepo.Get(ctx, &structs.FindNodeParams{
		NodeKey: nodeID,
	})
	if err != nil {
		return err
	}

	return s.validateNodeConfig(node)
}

func (s *nodeService) validateNodeConfig(node *ent.Node) error {
	// Implement node configuration validation logic
	switch structs.NodeType(node.Type) {
	case structs.NodeApproval:
		return s.validateApprovalNode(node)
	case structs.NodeService:
		return s.validateServiceNode(node)
	case structs.NodeExclusive:
		return s.validateExclusiveNode(node)
	case structs.NodeParallel:
		return s.validateParallelNode(node)
	}
	return nil
}

func (s *nodeService) validateApprovalNode(node *ent.Node) error {
	if validator.IsEmpty(node.Assignees) {
		return errors.New("approval node must have assignee configuration")
	}
	return nil
}

func (s *nodeService) validateServiceNode(node *ent.Node) error {
	if node.Handlers == nil {
		return errors.New("service node must have handler configuration")
	}
	return nil
}

func (s *nodeService) validateExclusiveNode(node *ent.Node) error {
	if len(node.Conditions) == 0 {
		return errors.New("exclusive node must have conditions")
	}
	return nil
}

func (s *nodeService) validateParallelNode(node *ent.Node) error {
	if len(node.ParallelNodes) == 0 {
		return errors.New("parallel node must have parallel nodes")
	}
	return nil
}

// Serialization helpers
func (s *nodeService) serialize(node *ent.Node) *structs.ReadNode {
	if node == nil {
		return nil
	}

	return &structs.ReadNode{
		ID:            node.ID,
		Name:          node.Name,
		Type:          node.Type,
		Description:   node.Description,
		Status:        node.Status,
		NodeKey:       node.NodeKey,
		ProcessID:     node.ProcessID,
		PrevNodes:     node.PrevNodes,
		NextNodes:     node.NextNodes,
		ParallelNodes: node.ParallelNodes,
		Conditions:    node.Conditions,
		Properties:    node.Properties,
		FormConfig:    node.FormConfig,
		Permissions:   node.Permissions,
		Assignees:     node.Assignees,
		Handlers:      node.Handlers,
		RetryTimes:    node.RetryTimes,
		RetryInterval: node.RetryInterval,
		IsWorkingDay:  node.IsWorkingDay,
		TenantID:      node.TenantID,
		Extras:        node.Extras,
		CreatedBy:     &node.CreatedBy,
		CreatedAt:     &node.CreatedAt,
		UpdatedBy:     &node.UpdatedBy,
		UpdatedAt:     &node.UpdatedAt,
	}
}

func (s *nodeService) serializes(nodes []*ent.Node) []*structs.ReadNode {
	result := make([]*structs.ReadNode, len(nodes))
	for i, node := range nodes {
		result[i] = s.serialize(node)
	}
	return result
}
