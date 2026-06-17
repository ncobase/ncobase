package router

import (
	"ncobase/internal/middleware"
	"ncobase/plugin/ai/handler"

	"github.com/gin-gonic/gin"
)

type Router struct {
	h *handler.Handler
}

func New(h *handler.Handler) *Router {
	return &Router{h: h}
}

func (r *Router) Register(rg *gin.RouterGroup, prefix ...string) {
	if len(prefix) > 0 {
		rg = rg.Group("/" + prefix[0])
	}

	protected := rg.Use(middleware.ValidateContentType(), middleware.RequireAuth())
	read := protected.Use(middleware.HasAnyPermission("read:ai", "use:ai", "manage:ai", "admin:ai"))
	use := protected.Use(middleware.HasAnyPermission("use:ai", "manage:ai", "admin:ai"))
	manage := protected.Use(middleware.HasAnyPermission("manage:ai", "admin:ai"))

	read.GET("/status", r.h.AI.Status)
	read.GET("/providers", r.h.AI.Providers)
	read.GET("/models", r.h.AI.Models)
	read.GET("/actions", r.h.Action.List)
	read.GET("/runs", r.h.Run.List)
	read.GET("/runs/:id", r.h.Run.Get)
	read.GET("/usage", r.h.Run.Usage)

	manage.GET("/health", r.h.AI.Health)

	use.POST("/complete", r.h.AI.Complete)
	use.POST("/stream", r.h.AI.Stream)
	use.POST("/embed", r.h.AI.Embed)
	use.POST("/actions/:action", r.h.Action.Run)
}
