package service

import (
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/slug"
	"stocms/pkg/validator"

	"github.com/gin-gonic/gin"
)

// CreateTaxonomyService creates a new taxonomy.
func (svc *Service) CreateTaxonomyService(c *gin.Context, body *structs.CreateTaxonomyBody) (*resp.Exception, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	taxonomy, err := svc.taxonomy.Create(c, body)
	if exception, err := handleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: taxonomy,
	}, nil
}

// UpdateTaxonomyService updates an existing taxonomy.
func (svc *Service) UpdateTaxonomyService(c *gin.Context, body *structs.UpdateTaxonomyBody) (*resp.Exception, error) {
	taxonomy, err := svc.taxonomy.Update(c, body)
	if exception, err := handleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: taxonomy,
	}, nil
}

// GetTaxonomyService retrieves a taxonomy by ID.
func (svc *Service) GetTaxonomyService(c *gin.Context, slug string) (*resp.Exception, error) {
	taxonomy, err := svc.taxonomy.GetBySlug(c, slug)
	if exception, err := handleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: taxonomy,
	}, nil
}

// DeleteTaxonomyService deletes a taxonomy by ID.
func (svc *Service) DeleteTaxonomyService(c *gin.Context, slug string) (*resp.Exception, error) {
	err := svc.taxonomy.Delete(c, slug)
	if exception, err := handleError("Taxonomy", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListTaxonomiesService lists all taxonomies.
func (svc *Service) ListTaxonomiesService(c *gin.Context, params *structs.ListTaxonomyParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	taxonomies, err := svc.taxonomy.List(c, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	return &resp.Exception{
		Data: taxonomies,
	}, nil
}
