package handler

import (
	"errors"
	"ncobase/plugin/ai/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
)

func handleAIError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	logger.Errorf(c.Request.Context(), "AI request failed: %v", err)
	switch {
	case errors.Is(err, service.ErrAIDisabled):
		resp.Fail(c.Writer, resp.ServiceUnavailable("AI service is disabled"))
	case errors.Is(err, service.ErrAIUnconfigured):
		resp.Fail(c.Writer, resp.ServiceUnavailable("AI provider is not configured"))
	case errors.Is(err, service.ErrPolicyDenied):
		resp.Fail(c.Writer, resp.Forbidden("AI request is denied by policy"))
	case errors.Is(err, service.ErrActionNotFound):
		resp.Fail(c.Writer, resp.NotFound("AI action is not registered"))
	default:
		resp.Fail(c.Writer, resp.InternalServer("AI request failed", err.Error()))
	}
}

func canManageAI(c *gin.Context) bool {
	ctx := c.Request.Context()
	if ctxutil.GetUserIsAdmin(ctx) {
		return true
	}
	for _, permission := range ctxutil.GetUserPermissions(ctx) {
		switch strings.TrimSpace(permission) {
		case "*:*", "*", "manage:ai", "admin:ai":
			return true
		}
	}
	return false
}
