// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"ncobase/proxy/data/ent/route"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// Route is the model entity for the Route schema.
type Route struct {
	config `json:"-"`
	// ID of the ent.
	// primary key
	ID string `json:"id,omitempty"`
	// name
	Name string `json:"name,omitempty"`
	// description
	Description string `json:"description,omitempty"`
	// is disabled
	Disabled bool `json:"disabled,omitempty"`
	// Extend properties
	Extras map[string]interface{} `json:"extras,omitempty"`
	// id of the creator
	CreatedBy string `json:"created_by,omitempty"`
	// id of the last updater
	UpdatedBy string `json:"updated_by,omitempty"`
	// created at
	CreatedAt int64 `json:"created_at,omitempty"`
	// updated at
	UpdatedAt int64 `json:"updated_at,omitempty"`
	// ID of the associated endpoint
	EndpointID string `json:"endpoint_id,omitempty"`
	// Path pattern for this route (e.g., /api/users/:id)
	PathPattern string `json:"path_pattern,omitempty"`
	// Target path on the remote API
	TargetPath string `json:"target_path,omitempty"`
	// HTTP method (GET, POST, PUT, DELETE, etc.)
	Method string `json:"method,omitempty"`
	// ID of the transformer to apply to incoming requests
	InputTransformerID string `json:"input_transformer_id,omitempty"`
	// ID of the transformer to apply to outgoing responses
	OutputTransformerID string `json:"output_transformer_id,omitempty"`
	// Whether to cache responses
	CacheEnabled bool `json:"cache_enabled,omitempty"`
	// Time to live for cached responses in seconds
	CacheTTL int `json:"cache_ttl,omitempty"`
	// Rate limit expression (e.g., 100/minute)
	RateLimit string `json:"rate_limit,omitempty"`
	// Whether to strip authentication header when forwarding
	StripAuthHeader bool `json:"strip_auth_header,omitempty"`
	selectValues    sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Route) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case route.FieldExtras:
			values[i] = new([]byte)
		case route.FieldDisabled, route.FieldCacheEnabled, route.FieldStripAuthHeader:
			values[i] = new(sql.NullBool)
		case route.FieldCreatedAt, route.FieldUpdatedAt, route.FieldCacheTTL:
			values[i] = new(sql.NullInt64)
		case route.FieldID, route.FieldName, route.FieldDescription, route.FieldCreatedBy, route.FieldUpdatedBy, route.FieldEndpointID, route.FieldPathPattern, route.FieldTargetPath, route.FieldMethod, route.FieldInputTransformerID, route.FieldOutputTransformerID, route.FieldRateLimit:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Route fields.
func (r *Route) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case route.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				r.ID = value.String
			}
		case route.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				r.Name = value.String
			}
		case route.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				r.Description = value.String
			}
		case route.FieldDisabled:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field disabled", values[i])
			} else if value.Valid {
				r.Disabled = value.Bool
			}
		case route.FieldExtras:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field extras", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &r.Extras); err != nil {
					return fmt.Errorf("unmarshal field extras: %w", err)
				}
			}
		case route.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				r.CreatedBy = value.String
			}
		case route.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				r.UpdatedBy = value.String
			}
		case route.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				r.CreatedAt = value.Int64
			}
		case route.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				r.UpdatedAt = value.Int64
			}
		case route.FieldEndpointID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field endpoint_id", values[i])
			} else if value.Valid {
				r.EndpointID = value.String
			}
		case route.FieldPathPattern:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field path_pattern", values[i])
			} else if value.Valid {
				r.PathPattern = value.String
			}
		case route.FieldTargetPath:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field target_path", values[i])
			} else if value.Valid {
				r.TargetPath = value.String
			}
		case route.FieldMethod:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field method", values[i])
			} else if value.Valid {
				r.Method = value.String
			}
		case route.FieldInputTransformerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field input_transformer_id", values[i])
			} else if value.Valid {
				r.InputTransformerID = value.String
			}
		case route.FieldOutputTransformerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field output_transformer_id", values[i])
			} else if value.Valid {
				r.OutputTransformerID = value.String
			}
		case route.FieldCacheEnabled:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field cache_enabled", values[i])
			} else if value.Valid {
				r.CacheEnabled = value.Bool
			}
		case route.FieldCacheTTL:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field cache_ttl", values[i])
			} else if value.Valid {
				r.CacheTTL = int(value.Int64)
			}
		case route.FieldRateLimit:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field rate_limit", values[i])
			} else if value.Valid {
				r.RateLimit = value.String
			}
		case route.FieldStripAuthHeader:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field strip_auth_header", values[i])
			} else if value.Valid {
				r.StripAuthHeader = value.Bool
			}
		default:
			r.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Route.
// This includes values selected through modifiers, order, etc.
func (r *Route) Value(name string) (ent.Value, error) {
	return r.selectValues.Get(name)
}

// Update returns a builder for updating this Route.
// Note that you need to call Route.Unwrap() before calling this method if this Route
// was returned from a transaction, and the transaction was committed or rolled back.
func (r *Route) Update() *RouteUpdateOne {
	return NewRouteClient(r.config).UpdateOne(r)
}

// Unwrap unwraps the Route entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (r *Route) Unwrap() *Route {
	_tx, ok := r.config.driver.(*txDriver)
	if !ok {
		panic("ent: Route is not a transactional entity")
	}
	r.config.driver = _tx.drv
	return r
}

// String implements the fmt.Stringer.
func (r *Route) String() string {
	var builder strings.Builder
	builder.WriteString("Route(")
	builder.WriteString(fmt.Sprintf("id=%v, ", r.ID))
	builder.WriteString("name=")
	builder.WriteString(r.Name)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(r.Description)
	builder.WriteString(", ")
	builder.WriteString("disabled=")
	builder.WriteString(fmt.Sprintf("%v", r.Disabled))
	builder.WriteString(", ")
	builder.WriteString("extras=")
	builder.WriteString(fmt.Sprintf("%v", r.Extras))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(r.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(r.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(fmt.Sprintf("%v", r.CreatedAt))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(fmt.Sprintf("%v", r.UpdatedAt))
	builder.WriteString(", ")
	builder.WriteString("endpoint_id=")
	builder.WriteString(r.EndpointID)
	builder.WriteString(", ")
	builder.WriteString("path_pattern=")
	builder.WriteString(r.PathPattern)
	builder.WriteString(", ")
	builder.WriteString("target_path=")
	builder.WriteString(r.TargetPath)
	builder.WriteString(", ")
	builder.WriteString("method=")
	builder.WriteString(r.Method)
	builder.WriteString(", ")
	builder.WriteString("input_transformer_id=")
	builder.WriteString(r.InputTransformerID)
	builder.WriteString(", ")
	builder.WriteString("output_transformer_id=")
	builder.WriteString(r.OutputTransformerID)
	builder.WriteString(", ")
	builder.WriteString("cache_enabled=")
	builder.WriteString(fmt.Sprintf("%v", r.CacheEnabled))
	builder.WriteString(", ")
	builder.WriteString("cache_ttl=")
	builder.WriteString(fmt.Sprintf("%v", r.CacheTTL))
	builder.WriteString(", ")
	builder.WriteString("rate_limit=")
	builder.WriteString(r.RateLimit)
	builder.WriteString(", ")
	builder.WriteString("strip_auth_header=")
	builder.WriteString(fmt.Sprintf("%v", r.StripAuthHeader))
	builder.WriteByte(')')
	return builder.String()
}

// Routes is a parsable slice of Route.
type Routes []*Route
