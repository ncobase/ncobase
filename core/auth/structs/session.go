package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// SessionBody represents session creation request
type SessionBody struct {
	UserID      string      `json:"user_id" validate:"required"`
	DeviceInfo  *types.JSON `json:"device_info,omitempty"`
	IPAddress   string      `json:"ip_address,omitempty"`
	UserAgent   string      `json:"user_agent,omitempty"`
	Location    string      `json:"location,omitempty"`
	LoginMethod string      `json:"login_method,omitempty"`
}

// ReadSession represents session data
type ReadSession struct {
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	TokenID      string      `json:"token_id"`
	DeviceInfo   *types.JSON `json:"device_info,omitempty"`
	IPAddress    string      `json:"ip_address,omitempty"`
	UserAgent    string      `json:"user_agent,omitempty"`
	Location     string      `json:"location,omitempty"`
	LoginMethod  string      `json:"login_method,omitempty"`
	IsActive     bool        `json:"is_active"`
	LastAccessAt *int64      `json:"last_access_at,omitempty"`
	ExpiresAt    *int64      `json:"expires_at,omitempty"`
	CreatedAt    *int64      `json:"created_at,omitempty"`
	UpdatedAt    *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value for pagination
func (s *ReadSession) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", s.ID, convert.ToValue(s.CreatedAt))
}

// ListSessionParams represents query parameters for listing sessions
type ListSessionParams struct {
	UserID    string `form:"user_id,omitempty" json:"user_id,omitempty"`
	IsActive  *bool  `form:"is_active,omitempty" json:"is_active,omitempty"`
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
}

// UpdateSessionBody represents session update request
type UpdateSessionBody struct {
	LastAccessAt *int64      `json:"last_access_at,omitempty"`
	Location     string      `json:"location,omitempty"`
	IsActive     *bool       `json:"is_active,omitempty"`
	DeviceInfo   *types.JSON `json:"device_info,omitempty"`
}
