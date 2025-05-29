package provider

import (
	"context"
	"net/http"

	"github.com/ncobase/ncore/config"
	extm "github.com/ncobase/ncore/extension/manager"
	"github.com/ncobase/ncore/logging/logger"
)

// NewServer creates a new server.
func NewServer(conf *config.Config) (http.Handler, func(), error) {

	// Initialize Extension Manager
	em, err := extm.NewManager(conf)
	if err != nil {
		logger.Fatalf(context.Background(), "Failed initializing extension manager: %+v", err)
		return nil, nil, err
	}

	// Register built-in extensions
	registerExtensions(em)

	// Register plugins
	if err = em.LoadPlugins(); err != nil {
		logger.Fatalf(context.Background(), "Failed loading plugins: %+v", err)
	}

	// New server
	h, err := ginServer(conf, em)
	if err != nil {
		logger.Fatalf(context.Background(), "Failed initializing http: %+v", err)
		// panic(err)
	}

	return h, func() {
		em.Cleanup()
	}, nil
}
