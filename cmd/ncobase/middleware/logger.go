package middleware

import (
	"bytes"
	"io"
	"ncobase/common/consts"
	"ncobase/common/log"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseLoggerWriter wraps the original ResponseWriter to capture response data
type ResponseLoggerWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w ResponseLoggerWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LogFormat holds the structure for logging information
type LogFormat struct {
	Status       int           `json:"status,omitempty"`
	Method       string        `json:"method,omitempty"`
	Path         string        `json:"path,omitempty"`
	Query        string        `json:"query,omitempty"`
	IP           string        `json:"ip,omitempty"`
	UserAgent    string        `json:"user_agent,omitempty"`
	Latency      time.Duration `json:"latency,omitempty"`
	RequestBody  string        `json:"request_body,omitempty"`
	ResponseBody string        `json:"response_body,omitempty"`
	ErrorMessage string        `json:"error,omitempty"`
}

// Logger is a middleware for logging requests with consistent tracing
func Logger(c *gin.Context) {
	start := time.Now()

	// Ensure context has a trace ID
	ctx := log.ContextWithTraceID(c.Request.Context())
	c.Request = c.Request.WithContext(ctx)

	// Capture request body
	var requestBody []byte
	if c.Request.Body != nil {
		requestBody, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	// Create a response writer that captures the response body
	blw := &ResponseLoggerWriter{body: new(bytes.Buffer), ResponseWriter: c.Writer}
	c.Writer = blw

	// Process request
	c.Next()

	// Prepare log entry
	entry := &LogFormat{
		Status:      c.Writer.Status(),
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		Query:       c.Request.URL.RawQuery,
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		Latency:     time.Since(start),
		RequestBody: string(requestBody),
	}

	// Only include response body for non-binary content types
	if !isBinaryContentType(c.Writer.Header().Get("Content-Type")) {
		entry.ResponseBody = blw.body.String()
	}

	// Log errors if any
	if len(c.Errors) > 0 {
		entry.ErrorMessage = c.Errors.String()
	}

	// Log the entry
	log.EntryFromContext(ctx).WithField("http", entry).Info("HTTP Request")

	// Set trace ID in response header
	c.Header(consts.XMdTraceKey, log.GetTraceID(ctx))
}

func isBinaryContentType(contentType string) bool {
	return contentType == "application/octet-stream" ||
		contentType == "application/pdf" ||
		contentType == "image/jpeg" ||
		contentType == "image/png" ||
		contentType == "audio/mpeg" ||
		contentType == "video/mp4"
}
