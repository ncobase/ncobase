package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
)

// EndpointBody defines the structure for request body used to create or update endpoints.
type EndpointBody struct {
	Name              string      `json:"name" validate:"required"`
	Description       string      `json:"description"`
	BaseURL           string      `json:"base_url" validate:"required,url"`
	Protocol          string      `json:"protocol" validate:"oneof=HTTP HTTPS WS WSS TCP UDP"`
	AuthType          string      `json:"auth_type" validate:"oneof=None Basic Bearer OAuth ApiKey"`
	AuthConfig        *string     `json:"auth_config"`
	Timeout           int         `json:"timeout"`
	UseCircuitBreaker bool        `json:"use_circuit_breaker"`
	RetryCount        int         `json:"retry_count"`
	ValidateSSL       bool        `json:"validate_ssl"`
	LogRequests       bool        `json:"log_requests"`
	LogResponses      bool        `json:"log_responses"`
	Disabled          bool        `json:"disabled"`
	Extras            *types.JSON `json:"extras,omitempty"`
	CreatedBy         *string     `json:"created_by,omitempty"`
	UpdatedBy         *string     `json:"updated_by,omitempty"`
}

// CreateEndpointBody represents the body for creating an endpoint.
type CreateEndpointBody struct {
	EndpointBody
}

// UpdateEndpointBody represents the body for updating an endpoint.
type UpdateEndpointBody struct {
	ID string `json:"id,omitempty"`
	EndpointBody
}

// ReadEndpoint represents the output schema for retrieving an endpoint.
type ReadEndpoint struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	BaseURL           string      `json:"base_url"`
	Protocol          string      `json:"protocol"`
	AuthType          string      `json:"auth_type"`
	AuthConfig        *string     `json:"auth_config,omitempty"`
	Timeout           int         `json:"timeout"`
	UseCircuitBreaker bool        `json:"use_circuit_breaker"`
	RetryCount        int         `json:"retry_count"`
	ValidateSSL       bool        `json:"validate_ssl"`
	LogRequests       bool        `json:"log_requests"`
	LogResponses      bool        `json:"log_responses"`
	Disabled          bool        `json:"disabled"`
	Extras            *types.JSON `json:"extras,omitempty"`
	CreatedBy         *string     `json:"created_by,omitempty"`
	CreatedAt         *int64      `json:"created_at,omitempty"`
	UpdatedBy         *string     `json:"updated_by,omitempty"`
	UpdatedAt         *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadEndpoint) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// ListEndpointParams represents the query parameters for listing endpoints.
type ListEndpointParams struct {
	Name      string `form:"name,omitempty" json:"name,omitempty"`
	Protocol  string `form:"protocol,omitempty" json:"protocol,omitempty"`
	Disabled  *bool  `form:"disabled,omitempty" json:"disabled,omitempty"`
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
}
