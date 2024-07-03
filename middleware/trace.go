package middleware

import (
	"ncobase/common/log"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

func Trace(c *gin.Context) {
	ctx := helper.FromGinContext(c)
	// Get the trace ID from the request
	traceID := helper.GetTraceID(ctx)

	// If trace ID is not present in the request, generate a new one
	if traceID == "" {
		traceID = helper.NewTraceID()
		// Set the trace ID in the request context
		c.Request = c.Request.WithContext(log.NewTraceIDContext(ctx, traceID))
	}

	// Set trace header in the response
	c.Writer.Header().Set("X-Trace-ID", traceID)

	c.Next()
}
