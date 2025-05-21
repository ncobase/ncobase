package provider

import (
	"context"
	"strings"

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
	// Auto-registration is handled by the registry system through init() functions
	// Just need to initialize the extensions
	if err := em.InitExtensions(); err != nil {
		logger.Errorf(context.Background(), "Failed to initialize extensions: %v", err)
		return
	}

	// Log all initialized extensions
	extensions := em.GetExtensions()
	extensionNames := make([]string, 0, len(extensions))
	for name := range extensions {
		extensionNames = append(extensionNames, name)
	}

	logger.Debugf(context.Background(), "Successfully initialized %d extensions: [%s]",
		len(extensionNames),
		strings.Join(extensionNames, ", "))
}
