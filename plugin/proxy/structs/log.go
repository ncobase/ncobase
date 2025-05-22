package structs

import (
	"fmt"
	"net/http"

	"github.com/ncobase/ncore/utils/convert"
)

// LogBody defines the structure for recording proxy logs.
type LogBody struct {
	EndpointID      string      `json:"endpoint_id"`
	RouteID         string      `json:"route_id"`
	RequestMethod   string      `json:"request_method"`
	RequestPath     string      `json:"request_path"`
	RequestHeaders  http.Header `json:"request_headers,omitempty"`
	RequestBody     string      `json:"request_body,omitempty"`
	StatusCode      int         `json:"status_code"`
	ResponseHeaders http.Header `json:"response_headers,omitempty"`
	ResponseBody    string      `json:"response_body,omitempty"`
	Duration        int         `json:"duration"`
	Error           string      `json:"error,omitempty"`
	ClientIP        string      `json:"client_ip,omitempty"`
	UserID          string      `json:"user_id,omitempty"`
}

// CreateLogBody represents the body for creating a proxy log.
type CreateLogBody struct {
	LogBody
}

// ReadLog represents the output schema for retrieving a proxy log.
type ReadLog struct {
	ID              string      `json:"id"`
	EndpointID      string      `json:"endpoint_id"`
	RouteID         string      `json:"route_id"`
	RequestMethod   string      `json:"request_method"`
	RequestPath     string      `json:"request_path"`
	RequestHeaders  http.Header `json:"request_headers,omitempty"`
	RequestBody     string      `json:"request_body,omitempty"`
	StatusCode      int         `json:"status_code"`
	ResponseHeaders http.Header `json:"response_headers,omitempty"`
	ResponseBody    string      `json:"response_body,omitempty"`
	Duration        int         `json:"duration"`
	Error           string      `json:"error,omitempty"`
	ClientIP        string      `json:"client_ip,omitempty"`
	UserID          string      `json:"user_id,omitempty"`
	CreatedAt       *int64      `json:"created_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadLog) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListLogParams represents the query parameters for listing proxy logs.
type ListLogParams struct {
	EndpointID    string `form:"endpoint_id,omitempty" json:"endpoint_id,omitempty"`
	RouteID       string `form:"route_id,omitempty" json:"route_id,omitempty"`
	RequestMethod string `form:"request_method,omitempty" json:"request_method,omitempty"`
	StatusCode    *int   `form:"status_code,omitempty" json:"status_code,omitempty"`
	Error         *bool  `form:"error,omitempty" json:"error,omitempty"`
	UserID        string `form:"user_id,omitempty" json:"user_id,omitempty"`
	FromTime      *int64 `form:"from_time,omitempty" json:"from_time,omitempty"`
	ToTime        *int64 `form:"to_time,omitempty" json:"to_time,omitempty"`
	Cursor        string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit         int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction     string `form:"direction,omitempty" json:"direction,omitempty"`
}
