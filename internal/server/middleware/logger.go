package middleware

import (
	"bytes"
	"io"
	"stocms/pkg/log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ResponseLoggerWriter wraps around the original ResponseWriter to capture response data.
type ResponseLoggerWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write overrides the Write method to capture response data.
func (w *ResponseLoggerWriter) Write(data []byte) (int, error) {
	if w.body != nil {
		w.body.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

// LogFormat holds the structure for logging information.
type LogFormat struct {
	Status       int               `json:"status,omitempty"`
	Path         string            `json:"path,omitempty"`
	Method       string            `json:"method,omitempty"`
	QueryParams  map[string]string `json:"params,omitempty"`
	Body         string            `json:"body,omitempty"`
	ResponseBody string            `json:"res,omitempty"`
	IP           string            `json:"ip,omitempty"`
	Latency      time.Duration     `json:"latency,omitempty"`
	StartAt      string            `json:"start_at,omitempty"`
	EndAt        string            `json:"end_at,omitempty"`
}

// ToLogrusFields Convert LogFormat to logrus.Fields
func (entry *LogFormat) ToLogrusFields() logrus.Fields {
	fields := logrus.Fields{}
	if entry.Status != 0 {
		fields["status"] = entry.Status
	}
	if entry.Path != "" {
		fields["path"] = entry.Path
	}
	if entry.Method != "" {
		fields["method"] = entry.Method
	}
	if len(entry.QueryParams) > 0 {
		fields["params"] = entry.QueryParams
	}
	if entry.Body != "" {
		fields["body"] = entry.Body
	}
	if entry.ResponseBody != "" {
		fields["res"] = entry.ResponseBody
	}
	if entry.IP != "" {
		fields["ip"] = entry.IP
	}
	if entry.Latency != 0 {
		fields["latency"] = entry.Latency
	}
	if entry.StartAt != "" {
		fields["start_at"] = entry.StartAt
	}
	if entry.EndAt != "" {
		fields["end_at"] = entry.EndAt
	}
	return fields
}

// // FormatLog converts the log entry struct into a key-value formatted string, excluding empty values.
// func FormatLog(entry *LogFormat, f ...string) string {
// 	if entry == nil {
// 		return ""
// 	}
//
// 	if len(f) > 0 && f[0] == "json" {
// 		jsonData, _ := json.Marshal(entry)
// 		return string(jsonData)
// 	}
//
// 	var b strings.Builder
//
// 	if entry.Status != 0 {
// 		fmt.Fprintf(&b, "status=%d ", entry.Status)
// 	}
// 	if entry.Path != "" {
// 		fmt.Fprintf(&b, "path=%s ", entry.Path)
// 	}
// 	if entry.Method != "" {
// 		fmt.Fprintf(&b, "method=%s ", entry.Method)
// 	}
// 	if len(entry.QueryParams) > 0 {
// 		queryParamsJSON, _ := json.Marshal(entry.QueryParams)
// 		_, _ = fmt.Fprintf(&b, "params=%s ", queryParamsJSON)
// 	}
// 	if entry.Body != "" {
// 		fmt.Fprintf(&b, "body=%s ", entry.Body)
// 	}
// 	if entry.ResponseBody != "" {
// 		fmt.Fprintf(&b, "res=%s ", entry.ResponseBody)
// 	}
// 	if entry.IP != "" {
// 		fmt.Fprintf(&b, "ip=%s ", entry.IP)
// 	}
// 	if entry.Latency != 0 {
// 		fmt.Fprintf(&b, "latency=%v ", entry.Latency)
// 	}
// 	if entry.StartAt != "" {
// 		fmt.Fprintf(&b, "start_at=%s ", entry.StartAt)
// 	}
// 	if entry.EndAt != "" {
// 		fmt.Fprintf(&b, "end_at=%s ", entry.EndAt)
// 	}
//
// 	return strings.TrimSpace(b.String())
// }

// Logger is a middleware for logging requests.
func Logger(c *gin.Context) {
	// Create a ResponseLoggerWriter to capture response data.
	responseWriter := &ResponseLoggerWriter{
		ResponseWriter: c.Writer,
		body:           bytes.NewBufferString(""),
	}
	c.Writer = responseWriter

	startTime := time.Now()

	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}

	entry := &LogFormat{
		StartAt:     startTime.Format(time.RFC3339Nano),
		Method:      c.Request.Method,
		Path:        path,
		IP:          c.ClientIP(),
		QueryParams: make(map[string]string),
	}

	// Read the request body if not in whitelist
	if !isInWhiteList(path) {
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			entry.Body = string(bodyBytes)
		}
	}

	// Capture query parameters
	for k, v := range c.Request.URL.Query() {
		entry.QueryParams[k] = v[0]
	}

	// Proceed with the request
	c.Next()

	// Capture response data if not in whitelist
	if !isInWhiteList(path) {
		entry.Status = c.Writer.Status()
		entry.Latency = time.Since(startTime)
		entry.ResponseBody = responseWriter.body.String()
	}

	entry.EndAt = time.Now().Format(time.RFC3339Nano)

	logEntry := log.StandardLogger().WithFields(entry.ToLogrusFields())
	logEntry.Info("Request completed")

	// config
	// conf := helper.GetConfig(c)
	// logEntry.Infof(FormatLog(entry, conf.Logger.Format))
}

// isInWhiteList checks if the given path is in the whitelist
func isInWhiteList(path string) bool {
	whitelist := []string{"/assets", "/assets/*"}
	for _, allowedPath := range whitelist {
		if path == allowedPath || strings.Contains(path, allowedPath) {
			return true
		}
	}
	return false
}
