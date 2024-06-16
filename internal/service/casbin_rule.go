package service

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/data/structs"

	"github.com/gin-gonic/gin"
)

// CreateCasbinRuleService creates a new Casbin rule.
func (svc *Service) CreateCasbinRuleService(c *gin.Context, body *structs.CasbinRuleBody) (*resp.Exception, error) {
	// Create a new CasbinRule entity
	casbinRule, err := svc.casbinRule.Create(c, body)
	if exception, err := handleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// UpdateCasbinRuleService updates an existing Casbin rule (full and partial).
func (svc *Service) UpdateCasbinRuleService(c *gin.Context, id string, updates types.JSON) (*resp.Exception, error) {
	if validator.IsEmpty(id) {
		return resp.BadRequest(ecode.FieldIsRequired("id")), nil
	}

	// Validate the updates map
	if len(updates) == 0 {
		return resp.BadRequest(ecode.FieldIsEmpty("updates fields")), nil
	}

	casbinRule, err := svc.casbinRule.Update(c, id, updates)
	if exception, err := handleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// GetCasbinRuleService retrieves a Casbin rule by ID.
func (svc *Service) GetCasbinRuleService(c *gin.Context, id string) (*resp.Exception, error) {
	casbinRule, err := svc.casbinRule.GetByID(c, id)
	if exception, err := handleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRule,
	}, nil
}

// DeleteCasbinRuleService deletes a Casbin rule by ID.
func (svc *Service) DeleteCasbinRuleService(c *gin.Context, id string) (*resp.Exception, error) {
	err := svc.casbinRule.Delete(c, id)
	if exception, err := handleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return nil, nil
}

// ListCasbinRulesService lists all Casbin rules based on query parameters.
func (svc *Service) ListCasbinRulesService(c *gin.Context, params *structs.CasbinRuleParams) (*resp.Exception, error) {
	casbinRules, err := svc.casbinRule.Find(c, params)
	if exception, err := handleError("CasbinRule", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: casbinRules,
	}, nil
}
