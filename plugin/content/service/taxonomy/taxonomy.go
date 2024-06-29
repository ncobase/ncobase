package taxonomy

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/helper"
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/data/ent"
	"ncobase/plugin/content/data/repository/taxonomy"
	"ncobase/plugin/content/structs"

	"github.com/gin-gonic/gin"
)

// ServiceInterface is the interface for the taxonomy service.
type ServiceInterface interface {
	Create(c *gin.Context, body *structs.CreateTaxonomyBody) (*resp.Exception, error)
	Update(c *gin.Context, slug string, updates types.JSON) (*resp.Exception, error)
	Get(c *gin.Context, slug string) (*resp.Exception, error)
	List(c *gin.Context, params *structs.ListTaxonomyParams) (*resp.Exception, error)
	Delete(c *gin.Context, slug string) (*resp.Exception, error)
	CreateTaxonomyRelation(c *gin.Context, body *structs.CreateTaxonomyRelationBody) (*resp.Exception, error)
	UpdateTaxonomyRelation(c *gin.Context, body *structs.UpdateTaxonomyRelationBody) (*resp.Exception, error)
	GetTaxonomyRelation(c *gin.Context, object string) (*resp.Exception, error)
	ListTaxonomyRelations(c *gin.Context, params *structs.ListTaxonomyRelationParams) (*resp.Exception, error)
	DeleteTaxonomyRelation(c *gin.Context, object string) (*resp.Exception, error)
}

type Service struct {
	taxonomy          taxonomy.RepositoryInterface
	taxonomyRelations taxonomy.RelationRepositoryInterface
}

func New(d *data.Data) ServiceInterface {
	return &Service{
		taxonomy:          taxonomy.NewTaxonomyRepo(d),
		taxonomyRelations: taxonomy.NewTaxonomyRelationRepo(d),
	}
}

// Create creates a new taxonomy.
func (svc *Service) Create(c *gin.Context, body *structs.CreateTaxonomyBody) (*resp.Exception, error) {
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
	row, err := svc.taxonomy.Create(c, body)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Update updates an existing taxonomy (full and partial)..
func (svc *Service) Update(c *gin.Context, slug string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug / id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	row, err := svc.taxonomy.Update(c, slug, updates)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Get retrieves a taxonomy by ID.
func (svc *Service) Get(c *gin.Context, slug string) (*resp.Exception, error) {
	row, err := svc.taxonomy.GetBySlug(c, slug)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: row,
	}, nil
}

// Delete deletes a taxonomy by ID.
func (svc *Service) Delete(c *gin.Context, slug string) (*resp.Exception, error) {
	err := svc.taxonomy.Delete(c, slug)
	if exception, err := helper.HandleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// List lists all taxonomies.
func (svc *Service) List(c *gin.Context, params *structs.ListTaxonomyParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	rows, err := svc.taxonomy.List(c, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	total := svc.taxonomy.CountX(c, params)

	return &resp.Exception{
		Data: &types.JSON{
			"content": rows,
			"total":   total,
		},
	}, nil
}

// CreateTaxonomyRelation creates a new taxonomy relation.
func (svc *Service) CreateTaxonomyRelation(c *gin.Context, body *structs.CreateTaxonomyRelationBody) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.Create(c, body)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// UpdateTaxonomyRelation updates an existing taxonomy relation.
func (svc *Service) UpdateTaxonomyRelation(c *gin.Context, body *structs.UpdateTaxonomyRelationBody) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.Update(c, body)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// GetTaxonomyRelation retrieves a taxonomy relation by ID.
func (svc *Service) GetTaxonomyRelation(c *gin.Context, object string) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.GetByObject(c, object)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// DeleteTaxonomyRelation deletes a taxonomy relation by ID.
func (svc *Service) DeleteTaxonomyRelation(c *gin.Context, object string) (*resp.Exception, error) {
	err := svc.taxonomyRelations.Delete(c, object)
	if exception, err := helper.HandleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListTaxonomyRelations lists all taxonomy relations.
func (svc *Service) ListTaxonomyRelations(c *gin.Context, params *structs.ListTaxonomyRelationParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	relations, err := svc.taxonomyRelations.List(c, params)

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
