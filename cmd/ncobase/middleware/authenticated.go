package middleware

import (
	"ncobase/common/ecode"
	"ncobase/common/helper"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/common/tracing"
	"ncobase/common/validator"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthenticatedTenant checks if the user is related to a tenant and authenticated.
func AuthenticatedTenant(c *gin.Context) {
	ctx := helper.FromGinContext(c)
	contextTraceID := helper.GetTraceID(ctx)
	tenantID := helper.GetTenantID(ctx)

	if validator.IsEmpty(tenantID) {
		log.EntryWithFields(ctx, logrus.Fields{
			tracing.TraceIDKey: contextTraceID,
		}).Warn("Tenant authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}

// AuthenticatedUser checks if the user is authenticated.
func AuthenticatedUser(c *gin.Context) {
	ctx := helper.FromGinContext(c)
	contextTraceID := helper.GetTraceID(ctx)
	userID := helper.GetUserID(ctx)

	if validator.IsEmpty(userID) {
		log.EntryWithFields(ctx, logrus.Fields{
			tracing.TraceIDKey: contextTraceID,
		}).Warn("User authentication failed")
		resp.Fail(c.Writer, resp.UnAuthorized(ecode.Text(ecode.Unauthorized)))
		c.Abort()
		return
	}

	c.Next()
}
