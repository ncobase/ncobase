package server

import (
	"context"
	"ncobase/common/config"
	"ncobase/common/log"
	"ncobase/internal/data"
	"ncobase/internal/handler"
	"ncobase/internal/service"
	"net/http"
)

// New creates a new server.
func New(conf *config.Config) (http.Handler, func(), error) {
	d, cleanup, err := data.New(&conf.Data)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Failed initializing data: %+v", err)
		// panic(err)
	}

	// Initialize services
	svc := service.New(d)

	// New HTTP server
	h, err := newHTTP(conf, handler.New(svc), svc)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Failed initializing http: %+v", err)
		// panic(err)
	}

	return h, cleanup, nil
}
