package helper

import (
	"context"
	"fmt"
	"ncobase/common/log"
	repo "ncobase/core/data/repository"
	"ncobase/core/data/structs"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// CasbinAdapter is the adapter for Casbin, using the repo layer.
type CasbinAdapter struct {
	casbinRepo repo.CasbinRule
}

// NewCasbinAdapter creates a new Casbin adapter.
func NewCasbinAdapter(casbinRepo repo.CasbinRule) *CasbinAdapter {
	return &CasbinAdapter{casbinRepo: casbinRepo}
}

// LoadPolicy loads all policy rules from the storage.
func (a *CasbinAdapter) LoadPolicy(model model.Model) error {
	ctx := context.Background()
	rules, err := a.casbinRepo.Find(ctx, &structs.ListCasbinRuleParams{})
	if err != nil {
		log.Errorf(ctx, "failed to load policies: %v", err)
		return err
	}

	for _, rule := range rules {
		line := strings.Join([]string{rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5}, ", ")
		err := persist.LoadPolicyLine(line, model)
		if err != nil {
			log.Errorf(ctx, "failed to load policy line: %v", err)
			return err
		}
	}
	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *CasbinAdapter) SavePolicy(model model.Model) error {
	ctx := context.Background()
	if err := a.casbinRepo.Delete(ctx, ""); err != nil {
		log.Errorf(ctx, "failed to delete policy: %v", err)
		return err
	}

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			ruleBody := &structs.CasbinRuleBody{
				PType: ptype,
				V0:    rule[0],
				V1:    rule[1],
				V2:    rule[2],
				V3:    &rule[3],
				V4:    &rule[4],
				V5:    &rule[5],
			}
			if _, err := a.casbinRepo.Create(ctx, ruleBody); err != nil {
				log.Errorf(ctx, "failed to save policy line: %v", err)
				return err
			}
		}
	}
	return nil
}

// AddPolicy adds a policy rule to the storage.
func (a *CasbinAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()

	ruleBody := &structs.CasbinRuleBody{
		PType: ptype,
		V0:    rule[0],
		V1:    rule[1],
		V2:    rule[2],
		V3:    &rule[3],
		V4:    &rule[4],
		V5:    &rule[5],
	}

	_, err := a.casbinRepo.Create(ctx, ruleBody)
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (a *CasbinAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	ctx := context.Background()

	ruleBody := &structs.CasbinRuleBody{
		PType: ptype,
		V0:    rule[0],
		V1:    rule[1],
		V2:    rule[2],
		V3:    &rule[3],
		V4:    &rule[4],
		V5:    &rule[5],
	}

	rules, err := a.casbinRepo.Find(ctx, &structs.ListCasbinRuleParams{
		PType: &ruleBody.PType,
		V0:    &ruleBody.V0,
		V1:    &ruleBody.V1,
		V2:    &ruleBody.V2,
		V3:    ruleBody.V3,
		V4:    ruleBody.V4,
		V5:    ruleBody.V5,
	})
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}

	err = a.casbinRepo.Delete(ctx, rules[0].ID)
	return err
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *CasbinAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	ctx := context.Background()

	params := &structs.ListCasbinRuleParams{
		PType: &ptype,
	}

	switch fieldIndex {
	case 0:
		if len(fieldValues) > 0 {
			params.V0 = &fieldValues[0]
		}
		if len(fieldValues) > 1 {
			params.V1 = &fieldValues[1]
		}
		if len(fieldValues) > 2 {
			params.V2 = &fieldValues[2]
		}
		if len(fieldValues) > 3 {
			params.V3 = &fieldValues[3]
		}
		if len(fieldValues) > 4 {
			params.V4 = &fieldValues[4]
		}
		if len(fieldValues) > 5 {
			params.V5 = &fieldValues[5]
		}
	case 1:
		if len(fieldValues) > 0 {
			params.V1 = &fieldValues[0]
		}
		if len(fieldValues) > 1 {
			params.V2 = &fieldValues[1]
		}
		if len(fieldValues) > 2 {
			params.V3 = &fieldValues[2]
		}
		if len(fieldValues) > 3 {
			params.V4 = &fieldValues[3]
		}
		if len(fieldValues) > 4 {
			params.V5 = &fieldValues[4]
		}
	case 2:
		if len(fieldValues) > 0 {
			params.V2 = &fieldValues[0]
		}
		if len(fieldValues) > 1 {
			params.V3 = &fieldValues[1]
		}
		if len(fieldValues) > 2 {
			params.V4 = &fieldValues[2]
		}
		if len(fieldValues) > 3 {
			params.V5 = &fieldValues[3]
		}
	case 3:
		if len(fieldValues) > 0 {
			params.V3 = &fieldValues[0]
		}
		if len(fieldValues) > 1 {
			params.V4 = &fieldValues[1]
		}
		if len(fieldValues) > 2 {
			params.V5 = &fieldValues[2]
		}
	case 4:
		if len(fieldValues) > 0 {
			params.V4 = &fieldValues[0]
		}
		if len(fieldValues) > 1 {
			params.V5 = &fieldValues[1]
		}
	case 5:
		if len(fieldValues) > 0 {
			params.V5 = &fieldValues[0]
		}
	default:
		return fmt.Errorf("fieldIndex %d out of range", fieldIndex)
	}

	rules, err := a.casbinRepo.Find(ctx, params)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		err = a.casbinRepo.Delete(ctx, rule.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
