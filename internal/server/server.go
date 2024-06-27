package server

import (
	"context"
	"ncobase/internal/data"
	"ncobase/internal/handler"
	"ncobase/internal/helper"
	"ncobase/internal/service"
	"net/http"
	"os"

	"ncobase/common/config"
	"ncobase/common/log"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

// New creates a new server.
func New(conf *config.Config) (http.Handler, func(), error) {
	d, cleanup, err := data.New(&conf.Data)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Failed initializing data: %+v", err)
		// panic(err)
	}

	// Initialize services
	svc := service.New(d)

	// Initialize Casbin model
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
	if conf.Auth.Casbin.Path != "" {
		// Load model from file
		fileContent, err := os.ReadFile(conf.Auth.Casbin.Path)
		if err != nil {
			return nil, nil, err
		}
		modelSource = string(fileContent)
	} else if conf.Auth.Casbin.Model != "" {
		// Use model provided as a string
		modelSource = conf.Auth.Casbin.Model
	} else {
		// Fallback to the default internal model source
		modelSource = defaultModelSource
	}

	// Load the Casbin model from the chosen model source
	m, err := model.NewModelFromString(modelSource)
	if err != nil {
		return nil, nil, err
	}

	// Initialize Casbin enforcer
	casbinRepo := svc.GetCasbinRuleRepo()
	adapter := helper.NewCasbinAdapter(casbinRepo)
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, cleanup, err
	}

	// Load policies from db
	err = e.LoadPolicy()
	if err != nil {
		return nil, cleanup, err
	}

	// New HTTP server
	h, err := newHTTP(conf, handler.New(svc), svc, e)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Failed initializing http: %+v", err)
		// panic(err)
	}

	return h, cleanup, nil
}

// func initializeRolesAndPermissions(roleRepo repo.Role, permissionRepo repo.Permission, casbinRuleRepo repo.CasbinRule) error {
// 	ctx := context.Background()
//
// 	// Define roles
// 	roles := []string{"admin", "user"}
//
// 	for _, roleName := range roles {
// 		role, err := roleRepo.FindRole(ctx, &structs.FindRole{Name: roleName})
// 		if err != nil {
// 			return err
// 		}
// 		if role == nil {
// 			role = &role.na
// 			if err := roleRepo.Create(ctx, role); err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	// Define permissions
// 	permissions := []string{"read", "write", "delete"}
//
// 	for _, permissionName := range permissions {
// 		permission, err := permissionRepo.FindPermission(ctx, &structs.FindPermission{Action: permissionName})
// 		if err != nil {
// 			return err
// 		}
// 		if permission == nil {
// 			permission = &data.Permission{Name: permissionName}
// 			if err := permissionRepo.Create(ctx, permission); err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	// Define Casbin rules
// 	rules := []struct {
// 		PType string
// 		V0    string
// 		V1    string
// 		V2    string
// 	}{
// 		{"p", "admin", "/v1/*", "GET"},
// 		{"p", "admin", "/v1/*", "POST"},
// 		{"p", "admin", "/v1/*", "PUT"},
// 		{"p", "admin", "/v1/*", "DELETE"},
// 		{"p", "user", "/v1/*", "GET"},
// 	}
//
// 	for _, rule := range rules {
// 		casbinRule := &data.CasbinRule{PType: rule.PType, V0: rule.V0, V1: rule.V1, V2: rule.V2}
// 		if err := casbinRuleRepo.Create(ctx, casbinRule); err != nil {
// 			return err
// 		}
// 	}
//
// 	return nil
// }
