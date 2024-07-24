package service

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/feature"
	"ncobase/common/log"
	"net/http"
)

// NewServer creates a new server.
func NewServer(conf *config.Config) (http.Handler, func(), error) {

	// Initialize Feature Manager
	fm := feature.NewManager(conf)
	registerFeatures(fm) // register built-in features
	if err := fm.LoadPlugins(); err != nil {
		log.Fatalf(context.Background(), "❌ Failed loading plugins: %+v", err)
	}

	// New server
	h, err := ginServer(conf, fm)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Failed initializing http: %+v", err)
		// panic(err)
	}

	return h, func() {
		fm.Cleanup()
	}, nil
}
