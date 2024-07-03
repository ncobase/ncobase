package service

import (
	"context"
	"ncobase/app/data/ent"
	"ncobase/app/data/structs"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/helper"
)

// CreateModuleService creates a new module.
func (svc *Service) CreateModuleService(ctx context.Context, body *structs.CreateModuleBody) (*resp.Exception, error) {
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	module, err := svc.module.Create(ctx, body)
	if exception, err := helper.HandleError("Module", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: module,
	}, nil
}

// UpdateModuleService updates an existing module (full and partial).
func (svc *Service) UpdateModuleService(ctx context.Context, slug string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(slug) {
		return resp.BadRequest(ecode.FieldIsRequired("slug / id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	module, err := svc.module.Update(ctx, slug, updates)
	if exception, err := helper.HandleError("Module", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: module,
	}, nil
}

// GetModuleService retrieves a module by slug.
func (svc *Service) GetModuleService(ctx context.Context, slug string) (*resp.Exception, error) {
	module, err := svc.module.GetBySlug(ctx, slug)
	if exception, err := helper.HandleError("Module", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: module,
	}, nil
}

// DeleteModuleService deletes a module by slug.
func (svc *Service) DeleteModuleService(ctx context.Context, slug string) (*resp.Exception, error) {
	err := svc.module.Delete(ctx, slug)
	if exception, err := helper.HandleError("Module", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListModulesService lists all modules.
func (svc *Service) ListModulesService(ctx context.Context, params *structs.ListModuleParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	rows, err := svc.module.List(ctx, params)

	if ent.IsNotFound(err) {
		return resp.NotFound(ecode.FieldIsInvalid("cursor")), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}

	total := svc.module.CountX(ctx, params)

	return &resp.Exception{
		Data: &types.JSON{
			"content": rows,
			"total":   total,
		},
	}, nil
}
