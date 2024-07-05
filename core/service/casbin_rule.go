package service

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/core/data/structs"
	"ncobase/helper"
)

// CreateCasbinRuleService creates a new Casbin rule.
func (svc *Service) CreateCasbinRuleService(ctx context.Context, body *structs.CasbinRuleBody) (*resp.Exception, error) {
	// Create a new CasbinRule entity
	casbinRule, err := svc.casbinRule.Create(ctx, body)
	if exception, err := helper.HandleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// UpdateCasbinRuleService updates an existing Casbin rule (full and partial).
func (svc *Service) UpdateCasbinRuleService(ctx context.Context, id string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(id) {
		return resp.BadRequest(ecode.FieldIsRequired("id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	casbinRule, err := svc.casbinRule.Update(ctx, id, updates)
	if exception, err := helper.HandleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// GetCasbinRuleService retrieves a Casbin rule by ID.
func (svc *Service) GetCasbinRuleService(ctx context.Context, id string) (*resp.Exception, error) {
	casbinRule, err := svc.casbinRule.GetByID(ctx, id)
	if exception, err := helper.HandleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// DeleteCasbinRuleService deletes a Casbin rule by ID.
func (svc *Service) DeleteCasbinRuleService(ctx context.Context, id string) (*resp.Exception, error) {
	err := svc.casbinRule.Delete(ctx, id)
	if exception, err := helper.HandleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListCasbinRulesService lists all Casbin rules based on query parameters.
func (svc *Service) ListCasbinRulesService(ctx context.Context, params *structs.ListCasbinRuleParams) (*resp.Exception, error) {
	casbinRules, err := svc.casbinRule.Find(ctx, params)
	if exception, err := helper.HandleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRules,
	}, nil
}
