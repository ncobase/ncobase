package middleware

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"ncobase/common/log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// responseWriter wraps the original responseWriter to capture response data
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write writes the data to the buffer
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

var (
	// skippedPaths is a list of paths that should be skipped for detailed logging
	skippedPaths = []string{
		"/swagger*",
		"/v1/attachments/*",
		"/attachments*",
		"/v1/swagger*",
	}

	// binaryTypes is a list of content types that should be treated as binary
	binaryTypes = []string{
		"application/octet-stream",
		"application/pdf",
		"image/",
		"audio/",
		"video/",
	}

	// Use a sync.Pool to reduce allocations
	bufferPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
)

// Logger is a middleware for logging
func Logger(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()

	// Check if the path should be skipped
	if shouldSkipPath(c.Request.URL.Path, skippedPaths) {
		c.Next()
		return
	}

	// Capture request body
	var requestBody any
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Errorf(ctx, "Failed to read request body: %v", err)
		} else {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = processBody(bodyBytes, c.ContentType(), c.Request.URL.Path)
		}
	}

	// Wrap response writer
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	w := &responseWriter{body: buf, ResponseWriter: c.Writer}
	c.Writer = w

	c.Next()

	// Prepare log entry
	entry := logrus.Fields{
		"status":     c.Writer.Status(),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"ip":         c.ClientIP(),
		"latency":    time.Since(start),
		"user_agent": c.Request.UserAgent(),
	}

	if requestBody != nil {
		entry["request_body"] = requestBody
	}

	responseBody := processBody(w.body.Bytes(), w.Header().Get("Content-Type"), c.Request.URL.Path)
	if responseBody != nil {
		entry["response_body"] = responseBody
	}

	if len(c.Errors) > 0 {
		entry["error"] = c.Errors.String()
	}

	// Log request
	l := log.EntryWithFields(ctx, entry)
	switch {
	case c.Writer.Status() >= http.StatusInternalServerError:
		l.Error("Internal Server Error")
	case c.Writer.Status() >= http.StatusBadRequest:
		l.Warn("Client Error")
	default:
		l.Info("Request Completed")
	}
}

// processBody processes the body of the request or response
func processBody(body []byte, contentType, path string) any {
	if len(body) == 0 {
		return nil
	}

	if isBinaryContentType(contentType) {
		return base64.StdEncoding.EncodeToString(body)
	}

	var jsonBody any
	if json.Valid(body) {
		if err := json.Unmarshal(body, &jsonBody); err != nil {
			return string(body)
		}
		return jsonBody
	}

	return string(body)
}

// isBinaryContentType checks if the content type is a binary type
func isBinaryContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	for _, t := range binaryTypes {
		if strings.HasPrefix(contentType, t) {
			return true
		}
	}
	return false
}
