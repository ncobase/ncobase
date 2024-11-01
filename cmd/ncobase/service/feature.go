package service

import (
	"context"
	"ncobase/common/feature"
	"ncobase/common/log"
	"ncobase/core/access"
	"ncobase/core/auth"
	"ncobase/core/space"
	"ncobase/core/system"
	"ncobase/core/tenant"
	"ncobase/core/user"
	"ncobase/domain/content"
	"ncobase/domain/realtime"
	"ncobase/domain/resource"
	"strings"
)

// registerFeatures registers all built-in features
func registerFeatures(fm *feature.Manager) {
	// All built-in components
	// Add more components here, disordered
	// dependent sorting through the getInitOrder method
	fs := make([]feature.Interface, 0, 10) // adjust this if you add more features
	// Core features
	fs = append(fs,
		access.New(),
		auth.New(),
		space.New(),
		system.New(),
		tenant.New(),
		user.New(),
	)

	// Domain features
	fs = append(fs,
		resource.New(),
		realtime.New(),
		content.New(),
	)

	// Registered features
	registered := make([]feature.Interface, 0, len(fs))
	// Get feature names
	featureNames := make([]string, 0, len(registered))

	for _, f := range fs {
		if err := fm.Register(f); err != nil {
			log.Errorf(context.Background(), "Failed to register feature %s: %v", f.Name(), err)
			continue // Skip this feature and try to register the next one
		}
		// log.Infof(context.Background(), "Successfully registered feature %s", f.Name())
		registered = append(registered, f)
		featureNames = append(featureNames, f.Name())
	}

	if len(registered) == 0 {
		log.Errorf(context.Background(), "No features were successfully registered.")
		return
	}

	// log.Infof(context.Background(), "Successfully registered %d features", len(registered))
	log.Infof(context.Background(), "Successfully registered %d features, [%s]",
		len(registered),
		strings.Join(featureNames, ", "))

	if err := fm.InitFeatures(); err != nil {
		log.Errorf(context.Background(), "Failed to initialize features: %v", err)
		return
	}
	// log.Infof(context.Background(), "All features initialized successfully")
}
