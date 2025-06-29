// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"ncobase/content/data/ent/media"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// Media is the model entity for the Media schema.
type Media struct {
	config `json:"-"`
	// ID of the ent.
	// primary key
	ID string `json:"id,omitempty"`
	// title
	Title string `json:"title,omitempty"`
	// type
	Type string `json:"type,omitempty"`
	// url, website / link...
	URL string `json:"url,omitempty"`
	// Extend properties
	Extras map[string]interface{} `json:"extras,omitempty"`
	// space id, e.g. space id, organization id, store id
	SpaceID string `json:"space_id,omitempty"`
	// id of the creator
	CreatedBy string `json:"created_by,omitempty"`
	// id of the last updater
	UpdatedBy string `json:"updated_by,omitempty"`
	// created at
	CreatedAt int64 `json:"created_at,omitempty"`
	// updated at
	UpdatedAt int64 `json:"updated_at,omitempty"`
	// Media owner identifier
	OwnerID string `json:"owner_id,omitempty"`
	// Reference to resource plugin file ID
	ResourceID string `json:"resource_id,omitempty"`
	// Media description
	Description string `json:"description,omitempty"`
	// Alternative text for accessibility
	Alt          string `json:"alt,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Media) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case media.FieldExtras:
			values[i] = new([]byte)
		case media.FieldCreatedAt, media.FieldUpdatedAt:
			values[i] = new(sql.NullInt64)
		case media.FieldID, media.FieldTitle, media.FieldType, media.FieldURL, media.FieldSpaceID, media.FieldCreatedBy, media.FieldUpdatedBy, media.FieldOwnerID, media.FieldResourceID, media.FieldDescription, media.FieldAlt:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Media fields.
func (m *Media) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case media.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				m.ID = value.String
			}
		case media.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				m.Title = value.String
			}
		case media.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				m.Type = value.String
			}
		case media.FieldURL:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field url", values[i])
			} else if value.Valid {
				m.URL = value.String
			}
		case media.FieldExtras:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field extras", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &m.Extras); err != nil {
					return fmt.Errorf("unmarshal field extras: %w", err)
				}
			}
		case media.FieldSpaceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field space_id", values[i])
			} else if value.Valid {
				m.SpaceID = value.String
			}
		case media.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				m.CreatedBy = value.String
			}
		case media.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				m.UpdatedBy = value.String
			}
		case media.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				m.CreatedAt = value.Int64
			}
		case media.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				m.UpdatedAt = value.Int64
			}
		case media.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				m.OwnerID = value.String
			}
		case media.FieldResourceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field resource_id", values[i])
			} else if value.Valid {
				m.ResourceID = value.String
			}
		case media.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				m.Description = value.String
			}
		case media.FieldAlt:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field alt", values[i])
			} else if value.Valid {
				m.Alt = value.String
			}
		default:
			m.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Media.
// This includes values selected through modifiers, order, etc.
func (m *Media) Value(name string) (ent.Value, error) {
	return m.selectValues.Get(name)
}

// Update returns a builder for updating this Media.
// Note that you need to call Media.Unwrap() before calling this method if this Media
// was returned from a transaction, and the transaction was committed or rolled back.
func (m *Media) Update() *MediaUpdateOne {
	return NewMediaClient(m.config).UpdateOne(m)
}

// Unwrap unwraps the Media entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (m *Media) Unwrap() *Media {
	_tx, ok := m.config.driver.(*txDriver)
	if !ok {
		panic("ent: Media is not a transactional entity")
	}
	m.config.driver = _tx.drv
	return m
}

// String implements the fmt.Stringer.
func (m *Media) String() string {
	var builder strings.Builder
	builder.WriteString("Media(")
	builder.WriteString(fmt.Sprintf("id=%v, ", m.ID))
	builder.WriteString("title=")
	builder.WriteString(m.Title)
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(m.Type)
	builder.WriteString(", ")
	builder.WriteString("url=")
	builder.WriteString(m.URL)
	builder.WriteString(", ")
	builder.WriteString("extras=")
	builder.WriteString(fmt.Sprintf("%v", m.Extras))
	builder.WriteString(", ")
	builder.WriteString("space_id=")
	builder.WriteString(m.SpaceID)
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(m.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(m.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(fmt.Sprintf("%v", m.CreatedAt))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(fmt.Sprintf("%v", m.UpdatedAt))
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(m.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("resource_id=")
	builder.WriteString(m.ResourceID)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(m.Description)
	builder.WriteString(", ")
	builder.WriteString("alt=")
	builder.WriteString(m.Alt)
	builder.WriteByte(')')
	return builder.String()
}

// MediaSlice is a parsable slice of Media.
type MediaSlice []*Media
