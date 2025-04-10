package service

import (
	"context"
	"errors"
	"fmt"
	nec "github.com/ncobase/ncore/ext/core"
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/paging"
	"github.com/ncobase/ncore/pkg/types"
	"ncobase/core/workflow/data/ent"
	"ncobase/core/workflow/data/repository"
	"ncobase/core/workflow/structs"
	"time"
)

type RuleServiceInterface interface {
	Create(ctx context.Context, body *structs.RuleBody) (*structs.ReadRule, error)
	Get(ctx context.Context, params *structs.FindRuleParams) (*structs.ReadRule, error)
	Update(ctx context.Context, body *structs.UpdateRuleBody) (*structs.ReadRule, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListRuleParams) (paging.Result[*structs.ReadRule], error)

	// Rule specific operations

	EnableRule(ctx context.Context, id string) error
	DisableRule(ctx context.Context, id string) error
	ValidateRule(ctx context.Context, id string) error
	GetActiveRules(ctx context.Context, templateID, nodeKey string) ([]*structs.ReadRule, error)
	EvaluateRules(ctx context.Context, processID string, data map[string]any) error
}

type ruleService struct {
	processRepo repository.ProcessRepositoryInterface
	ruleRepo    repository.RuleRepositoryInterface
	em          nec.ManagerInterface
}

func NewRuleService(repo repository.Repository, em nec.ManagerInterface) RuleServiceInterface {
	return &ruleService{
		processRepo: repo.GetProcess(),
		ruleRepo:    repo.GetRule(),
		em:          em,
	}
}

// Create creates a new rule
func (s *ruleService) Create(ctx context.Context, body *structs.RuleBody) (*structs.ReadRule, error) {
	if body.RuleKey == "" {
		return nil, errors.New(ecode.FieldIsRequired("rule_key"))
	}

	// Validate rule
	if err := s.validateRuleDefinition(body); err != nil {
		return nil, err
	}

	rule, err := s.ruleRepo.Create(ctx, body)
	if err := handleEntError(ctx, "Rule", err); err != nil {
		return nil, err
	}

	return s.serialize(rule), nil
}

// Get retrieves a specific rule
func (s *ruleService) Get(ctx context.Context, params *structs.FindRuleParams) (*structs.ReadRule, error) {
	rule, err := s.ruleRepo.Get(ctx, params)
	if err := handleEntError(ctx, "Rule", err); err != nil {
		return nil, err
	}

	return s.serialize(rule), nil
}

// Update updates an existing rule
func (s *ruleService) Update(ctx context.Context, body *structs.UpdateRuleBody) (*structs.ReadRule, error) {
	if body.ID == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate rule
	if err := s.validateRuleDefinition(&body.RuleBody); err != nil {
		return nil, err
	}

	rule, err := s.ruleRepo.Update(ctx, body)
	if err := handleEntError(ctx, "Rule", err); err != nil {
		return nil, err
	}

	return s.serialize(rule), nil
}

// Delete deletes a rule
func (s *ruleService) Delete(ctx context.Context, id string) error {
	return s.ruleRepo.Delete(ctx, id)
}

// List returns a list of rules
func (s *ruleService) List(ctx context.Context, params *structs.ListRuleParams) (paging.Result[*structs.ReadRule], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadRule, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rules, err := s.ruleRepo.List(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing rules: %v", err)
			return nil, 0, err
		}

		return s.serializes(rules), len(rules), nil
	})
}

// EnableRule enables a rule
func (s *ruleService) EnableRule(ctx context.Context, id string) error {
	return s.ruleRepo.EnableRule(ctx, id)
}

// DisableRule disables a rule
func (s *ruleService) DisableRule(ctx context.Context, id string) error {
	return s.ruleRepo.DisableRule(ctx, id)
}

// ValidateRule validates a rule
func (s *ruleService) ValidateRule(ctx context.Context, id string) error {
	rule, err := s.ruleRepo.Get(ctx, &structs.FindRuleParams{
		RuleKey: id,
	})
	if err != nil {
		return err
	}

	return s.validateRuleDefinition(&structs.RuleBody{
		Conditions: rule.Conditions,
		Actions:    rule.Actions,
	})
}

// GetActiveRules gets active rules
func (s *ruleService) GetActiveRules(ctx context.Context, templateID, nodeKey string) ([]*structs.ReadRule, error) {
	var rules []*ent.Rule
	var err error

	if nodeKey != "" {
		rules, err = s.ruleRepo.GetNodeRules(ctx, nodeKey)
	} else {
		rules, err = s.ruleRepo.GetTemplateRules(ctx, templateID)
	}

	if err != nil {
		return nil, err
	}

	return s.serializes(rules), nil
}

// EvaluateRules evaluates rules
func (s *ruleService) EvaluateRules(ctx context.Context, processID string, data map[string]any) error {
	// Get process
	process, err := s.processRepo.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return err
	}

	// Get active rules
	rules, err := s.GetActiveRules(ctx, process.TemplateID, process.CurrentNode)
	if err != nil {
		return err
	}

	// Evaluate rules
	for _, rule := range rules {
		if err := s.evaluateRule(rule, data); err != nil {
			logger.Errorf(ctx, "Failed to evaluate rule %s: %v", rule.ID, err)
			continue
		}
	}

	return nil
}

// Internal helpers
func (s *ruleService) validateRuleDefinition(body *structs.RuleBody) error {
	if body.Conditions == nil {
		return errors.New("rule conditions are required")
	}
	if body.Actions == nil {
		return errors.New("rule actions are required")
	}

	// Validate conditions structure
	if err := s.validateConditions(body.Conditions); err != nil {
		return fmt.Errorf("invalid conditions: %v", err)
	}

	// Validate actions structure
	if err := s.validateActions(body.Actions); err != nil {
		return fmt.Errorf("invalid actions: %v", err)
	}

	return nil
}

func (s *ruleService) validateConditions(conditions any) error {
	// TODO: Implement condition validation logic
	// Example structure validation:
	// {
	//   "operator": "and",
	//   "conditions": [
	//     {
	//       "field": "amount",
	//       "operator": "gt",
	//       "value": 1000
	//     }
	//   ]
	// }
	return nil
}

func (s *ruleService) validateActions(actions any) error {
	// TODO: Implement action validation logic
	// Example structure validation:
	// {
	//   "actions": [
	//     {
	//       "type": "assign",
	//       "target": "manager",
	//       "params": {}
	//     }
	//   ]
	// }
	return nil
}

func (s *ruleService) evaluateRule(rule *structs.ReadRule, data map[string]any) error {
	// Skip if rule is expired or not yet effective
	now := time.Now().UnixMilli()
	if rule.EffectiveTime != nil && now < *rule.EffectiveTime {
		return nil
	}
	if rule.ExpireTime != nil && now > *rule.ExpireTime {
		return nil
	}

	// Evaluate conditions
	matched, err := s.evaluateConditions(rule.Conditions, data)
	if err != nil {
		return err
	}

	// Execute actions if conditions match
	if matched {
		return s.executeActions(rule.Actions, data)
	}

	return nil
}

func (s *ruleService) evaluateConditions(conditions types.StringArray, data map[string]any) (bool, error) {
	// TODO: Implement condition evaluation logic
	return false, nil
}

func (s *ruleService) executeActions(actions types.JSON, data map[string]any) error {
	// TODO: Implement action execution logic
	return nil
}

// Serialization helpers
func (s *ruleService) serialize(rule *ent.Rule) *structs.ReadRule {
	if rule == nil {
		return nil
	}

	return &structs.ReadRule{
		ID:            rule.ID,
		Name:          rule.Name,
		Code:          rule.Code,
		Description:   rule.Description,
		Type:          rule.Type,
		Status:        rule.Status,
		RuleKey:       rule.RuleKey,
		TemplateID:    rule.TemplateID,
		NodeKey:       rule.NodeKey,
		Conditions:    rule.Conditions,
		Actions:       rule.Actions,
		Priority:      rule.Priority,
		IsEnabled:     rule.IsEnabled,
		EffectiveTime: &rule.EffectiveTime,
		ExpireTime:    &rule.ExpireTime,
		TenantID:      rule.TenantID,
		Extras:        rule.Extras,
		CreatedBy:     &rule.CreatedBy,
		CreatedAt:     &rule.CreatedAt,
		UpdatedBy:     &rule.UpdatedBy,
		UpdatedAt:     &rule.UpdatedAt,
	}
}

func (s *ruleService) serializes(rules []*ent.Rule) []*structs.ReadRule {
	result := make([]*structs.ReadRule, len(rules))
	for i, rule := range rules {
		result[i] = s.serialize(rule)
	}
	return result
}
