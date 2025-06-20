package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// SettingScope represents the scope of a setting
type SettingScope string

const (
	Scope        SettingScope = "space"
	ScopeSystem  SettingScope = "system"
	ScopeUser    SettingScope = "user"
	ScopeFeature SettingScope = "feature"
)

// SettingType represents the data type of setting value
type SettingType string

const (
	TypeString  SettingType = "string"
	TypeNumber  SettingType = "number"
	TypeBoolean SettingType = "boolean"
	TypeJSON    SettingType = "json"
	TypeArray   SettingType = "array"
)

// SpaceSettingBody represents a configuration setting for a space
type SpaceSettingBody struct {
	SpaceID      string       `json:"space_id,omitempty"`
	SettingKey   string       `json:"setting_key,omitempty"`
	SettingName  string       `json:"setting_name,omitempty"`
	SettingValue string       `json:"setting_value,omitempty"`
	DefaultValue string       `json:"default_value,omitempty"`
	SettingType  SettingType  `json:"setting_type,omitempty"`
	Scope        SettingScope `json:"scope,omitempty"`
	Category     string       `json:"category,omitempty"`
	Description  string       `json:"description,omitempty"`
	IsPublic     bool         `json:"is_public,omitempty"`
	IsRequired   bool         `json:"is_required,omitempty"`
	IsReadonly   bool         `json:"is_readonly,omitempty"`
	Validation   *types.JSON  `json:"validation,omitempty"`
	Extras       *types.JSON  `json:"extras,omitempty"`
	CreatedBy    *string      `json:"created_by,omitempty"`
	UpdatedBy    *string      `json:"updated_by,omitempty"`
}

// CreateSpaceSettingBody represents the body for creating space setting
type CreateSpaceSettingBody struct {
	SpaceSettingBody
}

// UpdateSpaceSettingBody represents the body for updating space setting
type UpdateSpaceSettingBody struct {
	ID string `json:"id,omitempty"`
	SpaceSettingBody
}

// ReadSpaceSetting represents the output schema for retrieving space setting
type ReadSpaceSetting struct {
	ID           string       `json:"id"`
	SpaceID      string       `json:"space_id"`
	SettingKey   string       `json:"setting_key"`
	SettingName  string       `json:"setting_name"`
	SettingValue string       `json:"setting_value"`
	DefaultValue string       `json:"default_value"`
	SettingType  SettingType  `json:"setting_type"`
	Scope        SettingScope `json:"scope"`
	Category     string       `json:"category"`
	Description  string       `json:"description"`
	IsPublic     bool         `json:"is_public"`
	IsRequired   bool         `json:"is_required"`
	IsReadonly   bool         `json:"is_readonly"`
	Validation   *types.JSON  `json:"validation,omitempty"`
	Extras       *types.JSON  `json:"extras,omitempty"`
	CreatedBy    *string      `json:"created_by,omitempty"`
	CreatedAt    *int64       `json:"created_at,omitempty"`
	UpdatedBy    *string      `json:"updated_by,omitempty"`
	UpdatedAt    *int64       `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value
func (r *ReadSpaceSetting) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// GetTypedValue returns the setting value converted to appropriate type
func (r *ReadSpaceSetting) GetTypedValue() any {
	switch r.SettingType {
	case TypeBoolean:
		return r.SettingValue == "true"
	case TypeNumber:
		if val, err := convert.StringToInt64(r.SettingValue); err == nil {
			return val
		}
		if val, err := convert.StringToFloat64(r.SettingValue); err == nil {
			return val
		}
		return 0
	case TypeJSON, TypeArray:
		var result any
		if convert.JSONUnmarshal(r.SettingValue, &result) {
			return result
		}
		return nil
	default:
		return r.SettingValue
	}
}

// BulkUpdateSettingsRequest represents request to update multiple settings
type BulkUpdateSettingsRequest struct {
	SpaceID  string            `json:"space_id" validate:"required"`
	Settings map[string]string `json:"settings" validate:"required"`
}

// ListSpaceSettingParams represents the query parameters for listing space settings
type ListSpaceSettingParams struct {
	SpaceID    string       `form:"space_id,omitempty" json:"space_id,omitempty"`
	Category   string       `form:"category,omitempty" json:"category,omitempty"`
	Scope      SettingScope `form:"scope,omitempty" json:"scope,omitempty"`
	IsPublic   *bool        `form:"is_public,omitempty" json:"is_public,omitempty"`
	IsRequired *bool        `form:"is_required,omitempty" json:"is_required,omitempty"`
	Cursor     string       `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit      int          `form:"limit,omitempty" json:"limit,omitempty"`
	Direction  string       `form:"direction,omitempty" json:"direction,omitempty"`
}
