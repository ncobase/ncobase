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
func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		l := &format{}
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		ctx.Next()
		end := time.Now()

		l.Status = ctx.Writer.Status()
		l.Method = ctx.Request.Method
		l.Path = path
		l.Body = ctx.Request.Body
		l.Latency = end.Sub(start)
		l.IP = ctx.ClientIP()

		log.Infof(ctx, "| %3d |  %s | %13v | %15s | %s |", l.Status, l.Method, l.Path, l.Latency, l.IP)
	}
}
