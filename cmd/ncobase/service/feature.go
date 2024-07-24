package service

import (
	"context"
	"ncobase/common/feature"
	"ncobase/common/log"
	"ncobase/feature/access"
	"ncobase/feature/auth"
	"ncobase/feature/group"
	"ncobase/feature/linker"
	"ncobase/feature/system"
	"ncobase/feature/tenant"
	"ncobase/feature/user"
)

// registerFeatures registers all built-in features
func registerFeatures(fm *feature.Manager) {
	fs := []feature.Interface{
		user.New(),
		group.New(),
		access.New(),
		tenant.New(),
		system.New(),
		auth.New(),
		// add more built-in components here, disordered
		linker.New(), // Must be the last one
	}

	var sfs []feature.Interface
	for _, f := range fs {
		if err := fm.Register(f); err != nil {
			log.Errorf(context.Background(), "❌ Failed to register feature %s: %v", f.Name(), err)
			continue // Skip this feature and try to register the next one
		}
		log.Infof(context.Background(), "✅ Successfully registered feature %s", f.Name())
		sfs = append(sfs, f)
	}

	if len(sfs) == 0 {
		log.Errorf(context.Background(), "❌ No features were successfully registered.")
		return
	}

	log.Infof(context.Background(), "✅ Successfully registered %d features", len(sfs))

	if err := fm.InitFeatures(); err != nil {
		log.Errorf(context.Background(), "❌ Failed to initialize features: %v", err)
		return
	}

	log.Infof(context.Background(), "✅ All features initialized successfully")
}
