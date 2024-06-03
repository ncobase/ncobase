package server

import (
	"net/http"
	"stocms/internal/config"
	"stocms/internal/data"
	"stocms/internal/handler"
	"stocms/internal/service"
	"stocms/pkg/log"
)

// New creates a new server.
func New(conf *config.Config) (http.Handler, func(), error) {
	d, cleanup, err := data.New(&conf.Data)
	if err != nil {
		log.Fatalf(nil, "❌ Failed initializing data: %+v", err)
		// panic(err)
	}

	// Initialize services
	svc := service.New(d)

	// New HTTP server
	h, err := newHTTP(conf, handler.New(svc), svc)
	if err != nil {
		log.Fatalf(nil, "❌ Failed initializing http: %+v", err)
		// panic(err)
	}

	return h, cleanup, nil
}
