package middleware

import (
	"io"
	"stocms/pkg/log"
	"time"

	"github.com/gin-gonic/gin"
)

type format struct {
	Status  int           `json:"status,omitempty"`
	Method  string        `json:"method,omitempty"`
	Path    string        `json:"path,omitempty"`
	Body    io.ReadCloser `json:"body,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
	IP      string        `json:"ip,omitempty"`
}

// Logger is a middleware for logging requests.
func Logger(c *gin.Context) {
	l := &format{}
	start := time.Now()
	path := c.Request.URL.Path
	raw := c.Request.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}
	c.Next()
	end := time.Now()

	l.Status = c.Writer.Status()
	l.Method = c.Request.Method
	l.Path = path
	l.Body = c.Request.Body
	l.Latency = end.Sub(start)
	l.IP = c.ClientIP()

	log.Infof(c.Request.Context(), "| %3d |  %s | %13v | %15s | %s |", l.Status, l.Method, l.Path, l.Latency, l.IP)
}
