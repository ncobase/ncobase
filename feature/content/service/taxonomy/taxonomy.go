package taxonomy

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/content/data"
	"ncobase/feature/content/data/ent"
	"ncobase/feature/content/data/repository/taxonomy"
	"ncobase/feature/content/structs"
	"ncobase/helper"
)

// ServiceInterface is the interface for the service.
type ServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*resp.Exception, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*resp.Exception, error)
	Get(ctx context.Context, slug string) (*resp.Exception, error)
	List(ctx context.Context, params *structs.ListTaxonomyParams) (*resp.Exception, error)
	Delete(ctx context.Context, slug string) (*resp.Exception, error)
	CreateTaxonomyRelation(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*resp.Exception, error)
	UpdateTaxonomyRelation(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*resp.Exception, error)
	GetTaxonomyRelation(ctx context.Context, object string) (*resp.Exception, error)
	ListTaxonomyRelations(ctx context.Context, params *structs.ListTaxonomyRelationParams) (*resp.Exception, error)
	DeleteTaxonomyRelation(ctx context.Context, object string) (*resp.Exception, error)
}

// Service is the struct for the service.
type Service struct {
	taxonomy          taxonomy.RepositoryInterface
	taxonomyRelations taxonomy.RelationRepositoryInterface
}

// New creates a new service.
func New(d *data.Data) ServiceInterface {
	return &Service{
		taxonomy:          taxonomy.NewTaxonomyRepo(d),
		taxonomyRelations: taxonomy.NewTaxonomyRelationRepo(d),
	}
}

// Create creates a new taxonomy.
func (svc *Service) Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*resp.Exception, error) {
	if validator.IsEmpty(body.Name) {
		return resp.BadRequest(ecode.FieldIsRequired("name")), nil
	}
	if validator.IsEmpty(body.Type) {
		return resp.BadRequest(ecode.FieldIsRequired("type")), nil
	}
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	row, err := svc.taxonomy.Create(ctx, body)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Update updates an existing taxonomy (full and partial)..
func (svc *Service) Update(ctx context.Context, slug string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug / id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	row, err := svc.taxonomy.Update(ctx, slug, updates)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Get retrieves a taxonomy by ID.
func (svc *Service) Get(ctx context.Context, slug string) (*resp.Exception, error) {
	row, err := svc.taxonomy.GetBySlug(ctx, slug)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Delete deletes a taxonomy by ID.
func (svc *Service) Delete(ctx context.Context, slug string) (*resp.Exception, error) {
	err := svc.taxonomy.Delete(ctx, slug)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// List lists all taxonomies.
func (svc *Service) List(ctx context.Context, params *structs.ListTaxonomyParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	rows, err := svc.taxonomy.List(ctx, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	total := svc.taxonomy.CountX(ctx, params)

	return &resp.Exception{
		Data: &types.JSON{
			"content": rows,
			"total":   total,
		},
	}, nil
}

// CreateTaxonomyRelation creates a new taxonomy relation.
func (svc *Service) CreateTaxonomyRelation(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.Create(ctx, body)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// UpdateTaxonomyRelation updates an existing taxonomy relation.
func (svc *Service) UpdateTaxonomyRelation(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.Update(ctx, body)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// GetTaxonomyRelation retrieves a taxonomy relation by ID.
func (svc *Service) GetTaxonomyRelation(ctx context.Context, object string) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.GetByObject(ctx, object)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// DeleteTaxonomyRelation deletes a taxonomy relation by ID.
func (svc *Service) DeleteTaxonomyRelation(ctx context.Context, object string) (*resp.Exception, error) {
	err := svc.taxonomyRelations.Delete(ctx, object)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListTaxonomyRelations lists all taxonomy relations.
func (svc *Service) ListTaxonomyRelations(ctx context.Context, params *structs.ListTaxonomyRelationParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	relations, err := svc.taxonomyRelations.List(ctx, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	return &resp.Exception{
		Data: relations,
	}, nil
}
