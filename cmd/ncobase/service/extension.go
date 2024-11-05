package service

import (
	"context"
	"ncobase/common/extension"
	"ncobase/common/log"
	"ncobase/core/access"
	"ncobase/core/auth"
	"ncobase/core/realtime"
	"ncobase/core/space"
	"ncobase/core/system"
	"ncobase/core/tenant"
	"ncobase/core/user"
	"ncobase/domain/content"
	"ncobase/domain/resource"
	"strings"
)

// registerExtensions registers all built-in extensions
func registerExtensions(em *extension.Manager) {
	// All built-in components
	// Add more components here, disordered
	// dependent sorting through the getInitOrder method
	fs := make([]extension.Interface, 0, 10) // adjust this if you add more extensions
	// Core extensions
	fs = append(fs,
		access.New(),
		auth.New(),
		space.New(),
		system.New(),
		tenant.New(),
		user.New(),
	)

	// Domain extensions
	fs = append(fs,
		resource.New(),
		realtime.New(),
		content.New(),
	)

	// Registered extensions
	registered := make([]extension.Interface, 0, len(fs))
	// Get extension names
	extensionNames := make([]string, 0, len(registered))

	for _, f := range fs {
		if err := em.Register(f); err != nil {
			log.Errorf(context.Background(), "Failed to register extension %s: %v", f.Name(), err)
			continue // Skip this extension and try to register the next one
		}
		// log.Infof(context.Background(), "Successfully registered extension %s", f.Name())
		registered = append(registered, f)
		extensionNames = append(extensionNames, f.Name())
	}

	if len(registered) == 0 {
		log.Errorf(context.Background(), "No extensions were successfully registered.")
		return
	}

	// log.Infof(context.Background(), "Successfully registered %d extensions", len(registered))
	log.Infof(context.Background(), "Successfully registered %d extensions, [%s]",
		len(registered),
		strings.Join(extensionNames, ", "))

	if err := em.InitExtensions(); err != nil {
		log.Errorf(context.Background(), "Failed to initialize extensions: %v", err)
		return
	}
	// log.Infof(context.Background(), "All extensions initialized successfully")
}
