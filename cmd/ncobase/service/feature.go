package service

import (
	"context"
	"ncobase/common/feature"
	"ncobase/common/log"
	"ncobase/domain/content"
	"ncobase/domain/resource"
	"ncobase/feature/access"
	"ncobase/feature/auth"
	"ncobase/feature/group"
	"ncobase/feature/socket"
	"ncobase/feature/system"
	"ncobase/feature/tenant"
	"ncobase/feature/user"
)

// registerFeatures registers all built-in features
func registerFeatures(fm *feature.Manager) {
	// all built-in components
	components := map[string][]feature.Interface{
		"features": {
			user.New(),
			group.New(),
			access.New(),
			tenant.New(),
			system.New(),
			auth.New(),
			socket.New(),
		},
		"domains": {
			resource.New(),
			content.New(),
		},
	}

	var sfs []feature.Interface

	// Register each component
	for category, comps := range components {
		for _, f := range comps {
			if err := fm.Register(f); err != nil {
				log.Errorf(context.Background(), "Failed to register %s %s: %v", category, f.Name(), err)
				continue // Skip this feature and try to register the next one
			}
			log.Infof(context.Background(), "Successfully registered %s %s", category, f.Name())
			sfs = append(sfs, f)
		}
	}

	if len(sfs) == 0 {
		log.Errorf(context.Background(), "No features were successfully registered.")
		return
	}

	log.Infof(context.Background(), "Successfully registered %d features", len(sfs))

	// Initialize all registered features
	if err := fm.InitFeatures(); err != nil {
		log.Errorf(context.Background(), "Failed to initialize features: %v", err)
		return
	}
	// log.Infof(context.Background(), "All features initialized successfully")
}
