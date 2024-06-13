package service

import (
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/slug"
	"stocms/pkg/types"
	"stocms/pkg/validator"

	"github.com/gin-gonic/gin"
)

// CreateModuleService creates a new module.
func (svc *Service) CreateModuleService(c *gin.Context, body *structs.CreateModuleBody) (*resp.Exception, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	module, err := svc.module.Create(c, body)
	if exception, err := handleError("Module", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: module,
	}, nil
}

// UpdateModuleService updates an existing module (full and partial).
func (svc *Service) UpdateModuleService(c *gin.Context, slug string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug / id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	module, err := svc.module.Update(c, slug, updates)
	if exception, err := handleError("Module", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: module,
	}, nil
}

// GetModuleService retrieves a module by slug.
func (svc *Service) GetModuleService(c *gin.Context, slug string) (*resp.Exception, error) {
	module, err := svc.module.GetBySlug(c, slug)
	if exception, err := handleError("Module", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: module,
	}, nil
}

// DeleteModuleService deletes a module by slug.
func (svc *Service) DeleteModuleService(c *gin.Context, slug string) (*resp.Exception, error) {
	err := svc.module.Delete(c, slug)
	if exception, err := handleError("Module", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListModulesService lists all modules.
func (svc *Service) ListModulesService(c *gin.Context, params *structs.ListModuleParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	modules, err := svc.module.List(c, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	return &resp.Exception{
		Data: modules,
	}, nil
}
