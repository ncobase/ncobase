package service

import (
	"context"
	"fmt"
	"ncobase/core/access/data"
	"ncobase/core/access/data/repository"
	"ncobase/core/access/structs"
	"ncore/pkg/config"
	"ncore/pkg/logger"
	"os"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// CasbinAdapterServiceInterface is the interface for the service.
type CasbinAdapterServiceInterface interface {
	InitEnforcer() (*casbin.Enforcer, error)
	LoadPolicy(model model.Model) error
	SavePolicy(model model.Model) error
	AddPolicy(sec string, ptype string, rule []string) error
	RemovePolicy(sec string, ptype string, rule []string) error
	RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error
}

// casbinAdapterService is the struct for the service.
type casbinAdapterService struct {
	conf   *config.Config
	casbin repository.CasbinRuleRepositoryInterface
}

// NewCasbinAdapterService creates a new service.
func NewCasbinAdapterService(conf *config.Config, d *data.Data) CasbinAdapterServiceInterface {
	return &casbinAdapterService{
		conf:   conf,
		casbin: repository.NewCasbinRule(d),
	}
}

// InitModel initializes the casbin model.
func (s *casbinAdapterService) initModel() (model.Model, error) {
	casbinConf := s.conf.Auth.Casbin
	var modelSource string
	// Define the default model source
	defaultModelSource := `
		[request_definition]
		r = sub, obj, act

		[policy_definition]
		p = sub, obj, act

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
	`
	if casbinConf.Path != "" {
		// Load model from file
		fileContent, err := os.ReadFile(casbinConf.Path)
		if err != nil {
			return nil, err
		}
		modelSource = string(fileContent)
	} else if casbinConf.Model != "" {
		// Use model provided as a string
		modelSource = casbinConf.Model
	} else {
		// Fallback to the default internal model source
		modelSource = defaultModelSource
	}

	// Load the Casbin model from the chosen model source
	m, err := model.NewModelFromString(modelSource)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// InitEnforcer initializes the casbin enforcer.
func (s *casbinAdapterService) InitEnforcer() (*casbin.Enforcer, error) {
	ctx := context.Background()

	m, err := s.initModel()
	if err != nil {
		logger.Errorf(ctx, "failed to initialize model: %v", err)
		return nil, err
	}

	// Create the enforcer
	e, err := casbin.NewEnforcer(m, s)
	if err != nil {
		logger.Errorf(ctx, "failed to create enforcer: %v", err)
		return nil, err
	}

	// Load policies from db
	err = e.LoadPolicy()
	if err != nil {
		logger.Errorf(ctx, "failed to load policies: %v", err)
		return nil, err
	}

	logger.Debugf(ctx, "Enforcer initialized and policies loaded successfully")
	return e, nil
}

// LoadPolicy loads all policy rules from the storage.
func (s *casbinAdapterService) LoadPolicy(model model.Model) error {
	ctx := context.Background()
	rules, err := s.casbin.Find(ctx, &structs.ListCasbinRuleParams{})
	if err != nil {
		logger.Errorf(ctx, "failed to load policies: %v", err)
		return err
	}

	for _, rule := range rules {
		line := strings.Join([]string{rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5}, ", ")
		err := persist.LoadPolicyLine(line, model)
		if err != nil {
			logger.Errorf(ctx, "failed to load policy line: %v", err)
			return err
		}
	}
	return nil
}

// SavePolicy saves all policy rules to the storage.
func (s *casbinAdapterService) SavePolicy(model model.Model) error {
	ctx := context.Background()
	if err := s.casbin.Delete(ctx, ""); err != nil {
		logger.Errorf(ctx, "failed to delete policy: %v", err)
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
			if _, err := s.casbin.Create(ctx, ruleBody); err != nil {
				logger.Errorf(ctx, "failed to save policy line: %v", err)
				return err
			}
		}
	}
	return nil
}

// AddPolicy adds a policy rule to the storage.
func (s *casbinAdapterService) AddPolicy(sec string, ptype string, rule []string) error {
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

	_, err := s.casbin.Create(ctx, ruleBody)
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (s *casbinAdapterService) RemovePolicy(sec string, ptype string, rule []string) error {
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

	rules, err := s.casbin.Find(ctx, &structs.ListCasbinRuleParams{
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

	err = s.casbin.Delete(ctx, rules[0].ID)
	return err
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (s *casbinAdapterService) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
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

	rules, err := s.casbin.Find(ctx, params)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		err = s.casbin.Delete(ctx, rule.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
