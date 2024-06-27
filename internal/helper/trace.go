package helper

import (
	"ncobase/common/nanoid"

	"github.com/gin-gonic/gin"
)

// SetRequestID sets request id to gin.Context
func SetRequestID(c *gin.Context, rid string) {
	SetValue(c, "request_id", rid)
}

// GetRequestID gets request id from gin.Context
func GetRequestID(c *gin.Context) string {
	if rid, ok := GetValue(c, "request_id").(string); ok {
		return rid
	}
	return ""
}

// SetTraceID sets trace id to gin.Context
func SetTraceID(c *gin.Context, traceID string) {
	SetValue(c, "trace_id", traceID)
}

// GetTraceID gets trace id from gin.Context
func GetTraceID(c *gin.Context) string {
	if traceID, ok := GetValue(c, "trace_id").(string); ok {
		return traceID
	}
	return ""
}

// NewTraceID creates a new trace ID.
func NewTraceID() string {
	return nanoid.Must(32)
}

// NewRequestID creates a new request ID.
func NewRequestID() string {
	return nanoid.Must(32)
}
