package repository

import (
	"context"
	"fmt"
	"ncobase/workflow/data"
	"ncobase/workflow/data/ent"
	nodeEnt "ncobase/workflow/data/ent/node"
	"ncobase/workflow/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// NodeRelations represents the relations between nodes
type NodeRelations struct {
	PrevNodes     []*ent.Node // Before nodes
	NextNodes     []*ent.Node // After nodes
	ParallelNodes []*ent.Node // Parallel nodes
	BranchNodes   []*ent.Node // Branch nodes
}

type NodeRepositoryInterface interface {
	Create(context.Context, *structs.NodeBody) (*ent.Node, error)
	Get(context.Context, *structs.FindNodeParams) (*ent.Node, error)
	Update(context.Context, *structs.UpdateNodeBody) (*ent.Node, error)
	Delete(context.Context, *structs.FindNodeParams) error
	List(context.Context, *structs.ListNodeParams) ([]*ent.Node, error)
	CountX(context.Context, *structs.ListNodeParams) int

	GetNodePath(ctx context.Context, fromNode, toNode string) ([]*ent.Node, error)
	GetNodeRelations(ctx context.Context, nodeID string) (*NodeRelations, error)
	UpdateNodeRelations(ctx context.Context, nodeID string, relations *NodeRelations) error
	ValidateNodeRelations(ctx context.Context, nodeID string) error

	UpdateStatus(context.Context, string, string) error
	GetNodesByProcessID(context.Context, string) ([]*ent.Node, error)
}

type nodeRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Node]
}

func NewNodeRepository(d *data.Data) NodeRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &nodeRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Node](rc, "workflow_node", false),
	}
}

// Create creates a new node
func (r *nodeRepository) Create(ctx context.Context, body *structs.NodeBody) (*ent.Node, error) {
	builder := r.ec.Node.Create()

	if body.Name != "" {
		builder.SetName(body.Name)
	}
	if body.Type != "" {
		builder.SetType(body.Type)
	}
	if body.Description != "" {
		builder.SetDescription(body.Description)
	}
	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if body.NodeKey != "" {
		builder.SetNodeKey(body.NodeKey)
	}
	if body.ProcessID != "" {
		builder.SetProcessID(body.ProcessID)
	}

	if validator.IsNotNil(body.PrevNodes) {
		builder.SetPrevNodes(body.PrevNodes)
	}
	if validator.IsNotNil(body.NextNodes) {
		builder.SetNextNodes(body.NextNodes)
	}
	if validator.IsNotNil(body.ParallelNodes) {
		builder.SetParallelNodes(body.ParallelNodes)
	}
	if body.Conditions != nil {
		builder.SetConditions(body.Conditions)
	}
	if body.Properties != nil {
		builder.SetProperties(body.Properties)
	}
	if body.FormConfig != nil {
		builder.SetFormConfig(body.FormConfig)
	}
	if body.Permissions != nil {
		builder.SetPermissions(body.Permissions)
	}
	if body.Handlers != nil {
		builder.SetHandlers(body.Handlers)
	}

	builder.SetRetryTimes(body.RetryTimes)
	builder.SetRetryInterval(body.RetryInterval)
	builder.SetIsWorkingDay(body.IsWorkingDay)

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "nodeRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.ms.IndexDocuments("nodes", row); err != nil {
		logger.Errorf(ctx, "nodeRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific node
func (r *nodeRepository) Get(ctx context.Context, params *structs.FindNodeParams) (*ent.Node, error) {
	builder := r.ec.Node.Query()

	if params.ProcessID != "" {
		builder.Where(nodeEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.Type != "" {
		builder.Where(nodeEnt.TypeEQ(params.Type))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(nodeEnt.StatusEQ(params.Status))
	}
	if params.NodeKey != "" {
		builder.Where(nodeEnt.NodeKeyEQ(params.NodeKey))
	}
	if params.Name != "" {
		builder.Where(nodeEnt.NameEQ(params.Name))
	}

	row, err := builder.First(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// Update updates an existing node
func (r *nodeRepository) Update(ctx context.Context, body *structs.UpdateNodeBody) (*ent.Node, error) {
	builder := r.ec.Node.UpdateOneID(body.ID)

	if validator.IsNotEmpty(body.Status) {
		builder.SetStatus(body.Status)
	}
	if validator.IsNotNil(body.Name) {
		builder.SetNextNodes(body.NextNodes)
	}
	if body.Properties != nil {
		builder.SetProperties(body.Properties)
	}
	if body.FormConfig != nil {
		builder.SetFormConfig(body.FormConfig)
	}
	if body.Handlers != nil {
		builder.SetHandlers(body.Handlers)
	}

	if body.Extras != nil {
		builder.SetExtras(body.Extras)
	}

	return builder.Save(ctx)
}

// Delete deletes a node
func (r *nodeRepository) Delete(ctx context.Context, params *structs.FindNodeParams) error {
	builder := r.ec.Node.Delete()

	if params.ProcessID != "" {
		builder.Where(nodeEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.NodeKey != "" {
		builder.Where(nodeEnt.NodeKeyEQ(params.NodeKey))
	}

	_, err := builder.Exec(ctx)
	return err
}

// List returns a list of nodes
func (r *nodeRepository) List(ctx context.Context, params *structs.ListNodeParams) ([]*ent.Node, error) {
	builder := r.ec.Node.Query()

	if params.ProcessID != "" {
		builder.Where(nodeEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.Type != "" {
		builder.Where(nodeEnt.TypeEQ(params.Type))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(nodeEnt.StatusEQ(params.Status))
	}

	// Add sorting
	switch params.SortBy {
	case structs.SortByName:
		builder.Order(ent.Asc(nodeEnt.FieldName))
	default:
		builder.Order(ent.Desc(nodeEnt.FieldCreatedAt))
	}

	builder.Limit(params.Limit)

	return builder.All(ctx)
}

// CountX returns the total count of nodes
func (r *nodeRepository) CountX(ctx context.Context, params *structs.ListNodeParams) int {
	builder := r.ec.Node.Query()

	if params.ProcessID != "" {
		builder.Where(nodeEnt.ProcessIDEQ(params.ProcessID))
	}
	if params.Type != "" {
		builder.Where(nodeEnt.TypeEQ(params.Type))
	}
	if validator.IsNotEmpty(params.Status) {
		builder.Where(nodeEnt.StatusEQ(params.Status))
	}

	return builder.CountX(ctx)
}

// GetNodePath returns the path between two nodes
func (r *nodeRepository) GetNodePath(ctx context.Context, fromNode, toNode string) ([]*ent.Node, error) {
	// get start and end nodes
	startNode, err := r.ec.Node.Get(ctx, fromNode)
	if err != nil {
		return nil, fmt.Errorf("failed to get start node: %w", err)
	}

	endNode, err := r.ec.Node.Get(ctx, toNode)
	if err != nil {
		return nil, fmt.Errorf("failed to get end node: %w", err)
	}

	// verify that the nodes belong to the same process
	if startNode.ProcessID != endNode.ProcessID {
		return nil, fmt.Errorf("nodes belong to different processes")
	}

	// bfs to find the path
	visited := make(map[string]bool)
	nodeQueue := [][]*ent.Node{{startNode}}

	for len(nodeQueue) > 0 {
		currentPath := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		currentNode := currentPath[len(currentPath)-1]

		// check if the current node is the end node
		if currentNode.ID == endNode.ID {
			return currentPath, nil
		}

		// get next nodes
		nextNodeIDs := currentNode.NextNodes
		if len(nextNodeIDs) == 0 {
			continue
		}

		nextNodes, err := r.ec.Node.Query().
			Where(
				nodeEnt.IDIn(nextNodeIDs...),
				nodeEnt.ProcessIDEQ(startNode.ProcessID),
			).All(ctx)
		if err != nil {
			logger.Warnf(ctx, "Failed to get next nodes for %s: %v", currentNode.ID, err)
			continue
		}

		// for each next node
		for _, nextNode := range nextNodes {
			if visited[nextNode.ID] {
				continue
			}
			visited[nextNode.ID] = true

			// create new path
			newPath := make([]*ent.Node, len(currentPath))
			copy(newPath, currentPath)
			newPath = append(newPath, nextNode)
			nodeQueue = append(nodeQueue, newPath)
		}
	}

	return nil, fmt.Errorf("no path found between nodes")
}

// GetNodeRelations returns the relations of a node
func (r *nodeRepository) GetNodeRelations(ctx context.Context, nodeID string) (*NodeRelations, error) {
	node, err := r.ec.Node.Get(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	relations := &NodeRelations{}

	// get related nodes
	relatedNodes, err := r.ec.Node.Query().
		Where(
			nodeEnt.ProcessIDEQ(node.ProcessID),
			nodeEnt.IDIn(
				append(
					append(
						append(
							node.PrevNodes,
							node.NextNodes...,
						),
						node.ParallelNodes...,
					),
				)...,
			),
		).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get related nodes: %w", err)
	}

	// map related nodes
	nodeMap := make(map[string]*ent.Node)
	for _, n := range relatedNodes {
		nodeMap[n.ID] = n
	}

	// distribute related nodes
	for _, id := range node.PrevNodes {
		if n, ok := nodeMap[id]; ok {
			relations.PrevNodes = append(relations.PrevNodes, n)
		}
	}

	for _, id := range node.NextNodes {
		if n, ok := nodeMap[id]; ok {
			relations.NextNodes = append(relations.NextNodes, n)
		}
	}

	for _, id := range node.ParallelNodes {
		if n, ok := nodeMap[id]; ok {
			relations.ParallelNodes = append(relations.ParallelNodes, n)
		}
	}

	// get branch nodes, other nodes sharing the same parent
	if len(node.PrevNodes) > 0 {
		allNodes, err := r.ec.Node.Query().
			Where(
				nodeEnt.ProcessIDEQ(node.ProcessID),
				nodeEnt.IDNEQ(node.ID),
			).All(ctx)
		if err != nil {
			logger.Warnf(ctx, "Failed to get process nodes: %v", err)
		} else {
			prevNodesSet := make(map[string]bool)
			for _, id := range node.PrevNodes {
				prevNodesSet[id] = true
			}

			for _, n := range allNodes {
				hasSamePrev := false
				for _, prevID := range n.PrevNodes {
					if prevNodesSet[prevID] {
						hasSamePrev = true
						break
					}
				}
				if hasSamePrev {
					relations.BranchNodes = append(relations.BranchNodes, n)
				}
			}
		}
	}

	return relations, nil
}

// UpdateNodeRelations updates the relations of a node
func (r *nodeRepository) UpdateNodeRelations(ctx context.Context, nodeID string, relations *NodeRelations) error {
	updates := r.ec.Node.UpdateOneID(nodeID)

	if relations.PrevNodes != nil {
		prevIDs := make([]string, len(relations.PrevNodes))
		for i, n := range relations.PrevNodes {
			prevIDs[i] = n.ID
		}
		updates.SetPrevNodes(prevIDs)
	}

	if relations.NextNodes != nil {
		nextIDs := make([]string, len(relations.NextNodes))
		for i, n := range relations.NextNodes {
			nextIDs[i] = n.ID
		}
		updates.SetNextNodes(nextIDs)
	}

	if relations.ParallelNodes != nil {
		parallelIDs := make([]string, len(relations.ParallelNodes))
		for i, n := range relations.ParallelNodes {
			parallelIDs[i] = n.ID
		}
		updates.SetParallelNodes(parallelIDs)
	}

	return updates.Exec(ctx)
}

// ValidateNodeRelations validates the relations of a node
func (r *nodeRepository) ValidateNodeRelations(ctx context.Context, nodeID string) error {
	node, err := r.ec.Node.Get(ctx, nodeID)
	if err != nil {
		return err
	}

	// validate related nodes in the same process
	allNodeIDs := append(
		append(
			append(
				[]string{},
				node.PrevNodes...,
			),
			node.NextNodes...,
		),
		node.ParallelNodes...,
	)

	if len(allNodeIDs) == 0 {
		return nil
	}

	count, err := r.ec.Node.Query().
		Where(
			nodeEnt.ProcessIDEQ(node.ProcessID),
			nodeEnt.IDIn(allNodeIDs...),
		).Count(ctx)
	if err != nil {
		return err
	}

	if count != len(allNodeIDs) {
		return fmt.Errorf("some related nodes do not exist or belong to different process")
	}

	return nil
}

// UpdateStatus updates the status of a node
func (r *nodeRepository) UpdateStatus(ctx context.Context, nodeID string, status string) error {
	return r.ec.Node.UpdateOneID(nodeID).
		SetStatus(status).
		Exec(ctx)
}

// GetNodesByProcessID returns all nodes for a specific process
func (r *nodeRepository) GetNodesByProcessID(ctx context.Context, processID string) ([]*ent.Node, error) {
	return r.ec.Node.Query().
		Where(nodeEnt.ProcessIDEQ(processID)).
		Order(ent.Asc(nodeEnt.FieldCreatedAt)).
		All(ctx)
}
