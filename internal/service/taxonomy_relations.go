package service

import (
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/validator"

	"github.com/gin-gonic/gin"
)

// CreateTaxonomyRelationService creates a new taxonomy relation.
func (svc *Service) CreateTaxonomyRelationService(c *gin.Context, body *structs.CreateTaxonomyRelationsBody) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.Create(c, body)
	if exception, err := handleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// UpdateTaxonomyRelationService updates an existing taxonomy relation.
func (svc *Service) UpdateTaxonomyRelationService(c *gin.Context, body *structs.UpdateTaxonomyRelationsBody) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.Update(c, body)
	if exception, err := handleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// GetTaxonomyRelationService retrieves a taxonomy relation by ID.
func (svc *Service) GetTaxonomyRelationService(c *gin.Context, object string) (*resp.Exception, error) {
	relation, err := svc.taxonomyRelations.GetByObject(c, object)
	if exception, err := handleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: relation,
	}, nil
}

// DeleteTaxonomyRelationService deletes a taxonomy relation by ID.
func (svc *Service) DeleteTaxonomyRelationService(c *gin.Context, object string) (*resp.Exception, error) {
	err := svc.taxonomyRelations.Delete(c, object)
	if exception, err := handleError("Taxonomy relation", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListTaxonomyRelationsService lists all taxonomy relations.
func (svc *Service) ListTaxonomyRelationsService(c *gin.Context, params *structs.ListTaxonomyRelationsParams) (*resp.Exception, error) {
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
