package provider

import (
	"context"
	"ncobase/core/access"
	"ncobase/core/auth"
	"ncobase/core/realtime"
	"ncobase/core/space"
	"ncobase/core/system"
	"ncobase/core/tenant"
	"ncobase/core/user"
	"ncobase/core/workflow"
	"ncobase/domain/content"
	"ncobase/domain/resource"
	"ncobase/proxy"
	"strings"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// registerExtensions registers all built-in extensions
func registerExtensions(em ext.ManagerInterface) {
	// All built-in components
	// Add more components here, disordered
	// dependent sorting through the getInitOrder method
	fs := make([]ext.Interface, 0)

	// Core extensions
	fs = append(fs,
		proxy.New(),
		access.New(),
		auth.New(),
		realtime.New(),
		space.New(),
		system.New(),
		tenant.New(),
		user.New(),
		workflow.New(),
	)

	// Domain extensions
	fs = append(fs,
		resource.New(),
		content.New(),
	)

	// Registered extensions
	registered := make([]ext.Interface, 0, len(fs))
	// Get extension names
	extensionNames := make([]string, 0, len(registered))

	for _, f := range fs {
		if err := em.Register(f); err != nil {
			logger.Errorf(context.Background(), "Failed to register extension %s: %v", f.Name(), err)
			continue // Skip this extension and try to register the next one
		}
		// log.Infof(context.Background(), "Successfully registered extension %s", f.Name())
		registered = append(registered, f)
		extensionNames = append(extensionNames, f.Name())
	}

	if len(registered) == 0 {
		logger.Errorf(context.Background(), "No extensions were successfully registered.")
		return
	}

	// log.Infof(context.Background(), "Successfully registered %d extensions", len(registered))
	logger.Debugf(context.Background(), "Successfully registered %d extensions, [%s]",
		len(registered),
		strings.Join(extensionNames, ", "))

	if err := em.InitExtensions(); err != nil {
		logger.Errorf(context.Background(), "Failed to initialize extensions: %v", err)
		return
	}
	// log.Infof(context.Background(), "All extensions initialized successfully")
}
