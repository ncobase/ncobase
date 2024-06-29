package server

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/internal/data"
	"ncobase/internal/handler"
	"ncobase/internal/helper"
	"ncobase/internal/service"
	"net/http"
	"os"

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
	m, err := initializeCasbinModel(conf)
	if err != nil {
		return nil, cleanup, err
	}

	// Initialize Casbin enforcer
	e, err := initializeCasbinEnforcer(m, svc)
	if err != nil {
		return nil, cleanup, err
	}

	// Initialize Plugin Manager
	pm := NewPluginManager(conf)
	if err := pm.LoadPlugins(); err != nil {
		log.Fatalf(context.Background(), "❌ Failed loading plugins: %+v", err)
	}

	// New HTTP server
	h, err := newHTTP(conf, handler.New(svc), svc, e, pm)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Failed initializing http: %+v", err)
		// panic(err)
	}

	return h, func() {
		cleanup()
		pm.CleanupPlugins()
	}, nil
}

// Initialize Casbin model
func initializeCasbinModel(conf *config.Config) (model.Model, error) {
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
			return nil, err
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
		return nil, err
	}

	return m, nil
}

// Initialize Casbin enforcer
func initializeCasbinEnforcer(m model.Model, svc *service.Service) (*casbin.Enforcer, error) {
	casbinRepo := svc.GetCasbinRuleRepo()
	adapter := helper.NewCasbinAdapter(casbinRepo)
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}
	// Load policies from db
	err = e.LoadPolicy()
	if err != nil {
		return nil, err
	}
	return e, nil
}
