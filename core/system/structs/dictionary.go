package structs

import (
	"encoding/json"
	"fmt"
	"ncore/pkg/types"
)

// DictionaryBody represents a dictionary.
type DictionaryBody struct {
	Name        string  `json:"name,omitempty"`
	Slug        string  `json:"slug,omitempty"`
	Type        string  `json:"type,omitempty"`
	Value       string  `json:"value,omitempty"`
	TenantID    string  `json:"tenant_id,omitempty"`
	Description string  `json:"description,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	UpdatedBy   *string `json:"updated_by,omitempty"`
}

// CreateDictionaryBody represents the body for creating a dictionary.
type CreateDictionaryBody struct {
	DictionaryBody
}

// UpdateDictionaryBody represents the body for updating a dictionary.
type UpdateDictionaryBody struct {
	ID string `json:"id,omitempty"`
	DictionaryBody
}

// ReadDictionary represents the output schema for retrieving a dictionary.
type ReadDictionary struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Type        string  `json:"type"`
	Value       string  `json:"value"`
	TenantID    string  `json:"tenant_id,omitempty"`
	Description string  `json:"description"`
	CreatedBy   *string `json:"created_by,omitempty"`
	CreatedAt   *int64  `json:"created_at,omitempty"`
	UpdatedBy   *string `json:"updated_by,omitempty"`
	UpdatedAt   *int64  `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadDictionary) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// ParseValue parses the value based on the type.
func (d *ReadDictionary) ParseValue() (any, error) {
	var result any
	switch d.Type {
	case "string":
		result = d.Value
	case "number":
		var num float64
		err := json.Unmarshal([]byte(d.Value), &num)
		if err != nil {
			return nil, err
		}
		result = num
	case "object":
		var obj map[string]any
		err := json.Unmarshal([]byte(d.Value), &obj)
		if err != nil {
			return nil, err
		}
		result = obj
	default:
		return nil, fmt.Errorf("unknown type: %s", d.Type)
	}
	return result, nil
}

// FindDictionary represents the parameters for finding a dictionary.
type FindDictionary struct {
	Dictionary string `form:"dictionary,omitempty" json:"dictionary,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
	Type       string `form:"type,omitempty" json:"type,omitempty"`
}

// ListDictionaryParams represents the query parameters for listing dictionaries.
type ListDictionaryParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
}
