package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/data/ent"
	"ncobase/plugin/content/data/repository"
	"ncobase/plugin/content/structs"
	"sort"
)

// TaxonomyServiceInterface is the interface for the service.
type TaxonomyServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*structs.ReadTaxonomy, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTaxonomy, error)
	Get(ctx context.Context, slug string) (*structs.ReadTaxonomy, error)
	List(ctx context.Context, params *structs.ListTaxonomyParams) (paging.Result[*structs.ReadTaxonomy], error)
	CountX(ctx context.Context, params *structs.ListTaxonomyParams) int
	GetTree(ctx context.Context, params *structs.FindTaxonomy) (paging.Result[*structs.ReadTaxonomy], error)
	Delete(ctx context.Context, slug string) error
	Serializes(rows []*ent.Taxonomy) []*structs.ReadTaxonomy
	Serialize(row *ent.Taxonomy) *structs.ReadTaxonomy
}

// taxonomyService is the struct for the service.
type taxonomyService struct {
	taxonomy          repository.TaxonomyRepositoryInterface
	taxonomyRelations repository.TaxonomyRelationsRepositoryInterface
}

// NewTaxonomyService creates a new service.
func NewTaxonomyService(d *data.Data) TaxonomyServiceInterface {
	return &taxonomyService{
		taxonomy:          repository.NewTaxonomyRepository(d),
		taxonomyRelations: repository.NewTaxonomyRelationsRepository(d),
	}
}

// Create creates a new taxonomy.
func (s *taxonomyService) Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*structs.ReadTaxonomy, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsRequired("name"))
	}
	if validator.IsEmpty(body.Type) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	row, err := s.taxonomy.Create(ctx, body)
	if err := handleEntError("Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing taxonomy (full and partial)..
func (s *taxonomyService) Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTaxonomy, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug / id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.taxonomy.Update(ctx, slug, updates)
	if err := handleEntError("Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a taxonomy by ID.
func (s *taxonomyService) Get(ctx context.Context, slug string) (*structs.ReadTaxonomy, error) {
	row, err := s.taxonomy.GetBySlug(ctx, slug)
	if err := handleEntError("Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a taxonomy by ID.
func (s *taxonomyService) Delete(ctx context.Context, slug string) error {
	err := s.taxonomy.Delete(ctx, slug)
	if err := handleEntError("Taxonomy", err); err != nil {
		return err
	}
	return nil
}

// List lists all taxonomies.
func (s *taxonomyService) List(ctx context.Context, params *structs.ListTaxonomyParams) (paging.Result[*structs.ReadTaxonomy], error) {
	if params.Children {
		return s.GetTree(ctx, &structs.FindTaxonomy{
			Children: true,
			Tenant:   params.Tenant,
			Taxonomy: params.Parent,
		})
	}
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTaxonomy, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.taxonomy.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing taxonomies: %v\n", err)
			return nil, 0, err
		}

		total := s.taxonomy.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes taxonomies.
func (s *taxonomyService) Serializes(rows []*ent.Taxonomy) []*structs.ReadTaxonomy {
	var rs []*structs.ReadTaxonomy
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a taxonomy.
func (s *taxonomyService) Serialize(row *ent.Taxonomy) *structs.ReadTaxonomy {
	return &structs.ReadTaxonomy{
		ID:          row.ID,
		Name:        row.Name,
		Type:        row.Type,
		Slug:        row.Slug,
		Cover:       row.Cover,
		Thumbnail:   row.Thumbnail,
		Color:       row.Color,
		Icon:        row.Icon,
		URL:         row.URL,
		Keywords:    row.Keywords,
		Description: row.Description,
		Status:      row.Status,
		Extras:      &row.Extras,
		ParentID:    &row.ParentID,
		TenantID:    &row.ParentID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// CountX gets a count of taxonomys.
func (s *taxonomyService) CountX(ctx context.Context, params *structs.ListTaxonomyParams) int {
	return s.taxonomy.CountX(ctx, params)
}

// GetTree retrieves the taxonomy tree.
func (s *taxonomyService) GetTree(ctx context.Context, params *structs.FindTaxonomy) (paging.Result[*structs.ReadTaxonomy], error) {
	rows, err := s.taxonomy.GetTree(ctx, params)
	if err := handleEntError("Taxonomy", err); err != nil {
		return paging.Result[*structs.ReadTaxonomy]{}, err
	}

	return paging.Result[*structs.ReadTaxonomy]{
		Items: s.buildTaxonomyTree(rows),
		Total: len(rows),
	}, nil
}

// buildTaxonomyTree builds a taxonomy tree structure.
func (s *taxonomyService) buildTaxonomyTree(taxonomys []*ent.Taxonomy) []*structs.ReadTaxonomy {
	// Convert taxonomys to ReadTaxonomy objects
	taxonomyNodes := make([]types.TreeNode, len(taxonomys))
	for i, m := range taxonomys {
		taxonomyNodes[i] = s.Serialize(m)
	}

	// Sort taxonomy nodes
	sortTaxonomyNodes(taxonomyNodes)

	// Build tree structure
	tree := types.BuildTree(taxonomyNodes)

	result := make([]*structs.ReadTaxonomy, len(tree))
	for i, node := range tree {
		result[i] = node.(*structs.ReadTaxonomy)
	}

	return result
}

// sortTaxonomyNodes sorts taxonomy nodes.
func sortTaxonomyNodes(taxonomyNodes []types.TreeNode) {
	// Recursively sort children nodes first
	for _, node := range taxonomyNodes {
		children := node.GetChildren()
		sortTaxonomyNodes(children)

		// Sort children and set back to node
		sort.SliceStable(children, func(i, j int) bool {
			nodeI := children[i].(*structs.ReadTaxonomy)
			nodeJ := children[j].(*structs.ReadTaxonomy)
			return types.ToValue(nodeI.CreatedAt) < (types.ToValue(nodeJ.CreatedAt))
		})
		node.SetChildren(children)
	}

	// Sort the immediate children of the current level
	sort.SliceStable(taxonomyNodes, func(i, j int) bool {
		nodeI := taxonomyNodes[i].(*structs.ReadTaxonomy)
		nodeJ := taxonomyNodes[j].(*structs.ReadTaxonomy)
		return types.ToValue(nodeI.CreatedAt) < (types.ToValue(nodeJ.CreatedAt))
	})
}
