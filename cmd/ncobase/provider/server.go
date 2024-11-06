package provider

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/extension"
	"ncobase/common/log"
	"net/http"
)

// NewServer creates a new server.
func NewServer(conf *config.Config) (http.Handler, func(), error) {

	// Initialize Extension Manager
	em, err := extension.NewManager(conf)
	if err != nil {
		log.Fatalf(context.Background(), "Failed initializing extension manager: %+v", err)
		return nil, nil, err
	}

	// Register built-in extensions
	registerExtensions(em)
	if err := em.LoadPlugins(); err != nil {
		log.Fatalf(context.Background(), "Failed loading plugins: %+v", err)
	}

	// New server
	h, err := ginServer(conf, em)
	if err != nil {
		log.Fatalf(context.Background(), "Failed initializing http: %+v", err)
		// panic(err)
	}

	return h, func() {
		em.Cleanup()
	}, nil
}
