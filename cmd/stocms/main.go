package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stocms/internal/config"
	"stocms/internal/server"
	"stocms/pkg/log"
	"syscall"
	"time"

	_ "stocms/docs"
)

// Version Version number, can be specified during compilation: go build -ldflags "-X main.Version=x.x.x"
var Version = "dev+"

// @title stocms
// @version 0.1.0
// @description a modern content management system
// @termsOfService https://stocms.com

func main() {
	log.SetVersion(Version)

	// Loading config
	conf, err := config.Init()
	if err != nil {
		log.Fatalf(context.Background(), "❌ Config initialization error: %+v", err)
	}

	// Initialize logger
	loggerClean, err := log.Init(conf.Logger)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Logger initialization error: %+v", err)
	}
	defer loggerClean()

	// Print application name
	log.Infof(context.Background(), "%s", conf.AppName)

	// Create server
	handler, cleanup, err := server.New(conf)
	if err != nil {
		log.Fatalf(context.Background(), "❌ Failed to start server: %+v", err)
	}

	// Cleanup
	defer cleanup()

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log.Infof(context.Background(), "🚀 Listening and serving HTTP on: %s", addr)

	go func() {
		// Service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf(context.Background(), "listen: %s", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof(context.Background(), "⌛️ Shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf(context.Background(), "❌ Server shutdown: %+v", err)
	}
	// Catching ctx.Done(). Timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Infof(context.Background(), "⌛️ Timeout of 3 seconds.")
	}
	log.Infof(context.Background(), "👋 Server exiting")
}
