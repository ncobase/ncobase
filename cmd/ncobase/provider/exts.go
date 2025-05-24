package provider

import (
	"context"
	// Core modules
	_ "ncobase/core"

	// Domain modules
	_ "ncobase/domain"

	// Proxy modules
	_ "ncobase/proxy"

	// Plugins
	_ "ncobase/plugin"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// registerExtensions registers all built-in extensions
func registerExtensions(em ext.ManagerInterface) {
	// Registration is handled by the registry system through init() functions
	if err := em.InitExtensions(); err != nil {
		logger.Errorf(context.Background(), "Failed to initialize extensions: %v", err)
		return
	}
}
