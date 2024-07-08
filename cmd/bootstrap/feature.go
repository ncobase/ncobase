package bootstrap

import (
	"context"
	"ncobase/common/log"
	"ncobase/feature"
	"ncobase/feature/menu"
)

// registerFeatures registers all built-in features
func registerFeatures(fm *feature.Manager) {
	builtinFeatures := []feature.Interface{
		menu.New(),
		// add more built-in components here
	}
	for _, f := range builtinFeatures {
		if err := fm.Register(f); err != nil {
			log.Errorf(context.Background(), "❌ Failed to register feature %s: %v", f.Name(), err)
		}
		log.Infof(context.Background(), "✅ Successfully registered feature %s", f.Name())
	}

	// log.Infof(context.Background(), "All built-in features registered successfully")

	if err := fm.InitFeatures(); err != nil {
		log.Errorf(context.Background(), "❌ Failed to initialize feature: %v", err)
	}

	// log.Infof(context.Background(), "All features initialized successfully")
}
