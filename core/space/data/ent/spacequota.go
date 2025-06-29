// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"ncobase/space/data/ent/spacequota"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// SpaceQuota is the model entity for the SpaceQuota schema.
type SpaceQuota struct {
	config `json:"-"`
	// ID of the ent.
	// primary key
	ID string `json:"id,omitempty"`
	// space id, e.g. space id, organization id, store id
	SpaceID string `json:"space_id,omitempty"`
	// description
	Description string `json:"description,omitempty"`
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
	// Type of quota (users, storage, api_calls, etc.)
	QuotaType string `json:"quota_type,omitempty"`
	// Human readable name of the quota
	QuotaName string `json:"quota_name,omitempty"`
	// Maximum allowed value for this quota
	MaxValue int64 `json:"max_value,omitempty"`
	// Current usage of this quota
	CurrentUsed int64 `json:"current_used,omitempty"`
	// Unit of measurement (count, bytes, mb, gb, tb)
	Unit string `json:"unit,omitempty"`
	// Whether this quota is actively enforced
	Enabled      bool `json:"enabled,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SpaceQuota) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case spacequota.FieldExtras:
			values[i] = new([]byte)
		case spacequota.FieldEnabled:
			values[i] = new(sql.NullBool)
		case spacequota.FieldCreatedAt, spacequota.FieldUpdatedAt, spacequota.FieldMaxValue, spacequota.FieldCurrentUsed:
			values[i] = new(sql.NullInt64)
		case spacequota.FieldID, spacequota.FieldSpaceID, spacequota.FieldDescription, spacequota.FieldCreatedBy, spacequota.FieldUpdatedBy, spacequota.FieldQuotaType, spacequota.FieldQuotaName, spacequota.FieldUnit:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SpaceQuota fields.
func (sq *SpaceQuota) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case spacequota.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				sq.ID = value.String
			}
		case spacequota.FieldSpaceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field space_id", values[i])
			} else if value.Valid {
				sq.SpaceID = value.String
			}
		case spacequota.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				sq.Description = value.String
			}
		case spacequota.FieldExtras:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field extras", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sq.Extras); err != nil {
					return fmt.Errorf("unmarshal field extras: %w", err)
				}
			}
		case spacequota.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				sq.CreatedBy = value.String
			}
		case spacequota.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				sq.UpdatedBy = value.String
			}
		case spacequota.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				sq.CreatedAt = value.Int64
			}
		case spacequota.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				sq.UpdatedAt = value.Int64
			}
		case spacequota.FieldQuotaType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field quota_type", values[i])
			} else if value.Valid {
				sq.QuotaType = value.String
			}
		case spacequota.FieldQuotaName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field quota_name", values[i])
			} else if value.Valid {
				sq.QuotaName = value.String
			}
		case spacequota.FieldMaxValue:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field max_value", values[i])
			} else if value.Valid {
				sq.MaxValue = value.Int64
			}
		case spacequota.FieldCurrentUsed:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field current_used", values[i])
			} else if value.Valid {
				sq.CurrentUsed = value.Int64
			}
		case spacequota.FieldUnit:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field unit", values[i])
			} else if value.Valid {
				sq.Unit = value.String
			}
		case spacequota.FieldEnabled:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field enabled", values[i])
			} else if value.Valid {
				sq.Enabled = value.Bool
			}
		default:
			sq.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SpaceQuota.
// This includes values selected through modifiers, order, etc.
func (sq *SpaceQuota) Value(name string) (ent.Value, error) {
	return sq.selectValues.Get(name)
}

// Update returns a builder for updating this SpaceQuota.
// Note that you need to call SpaceQuota.Unwrap() before calling this method if this SpaceQuota
// was returned from a transaction, and the transaction was committed or rolled back.
func (sq *SpaceQuota) Update() *SpaceQuotaUpdateOne {
	return NewSpaceQuotaClient(sq.config).UpdateOne(sq)
}

// Unwrap unwraps the SpaceQuota entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (sq *SpaceQuota) Unwrap() *SpaceQuota {
	_tx, ok := sq.config.driver.(*txDriver)
	if !ok {
		panic("ent: SpaceQuota is not a transactional entity")
	}
	sq.config.driver = _tx.drv
	return sq
}

// String implements the fmt.Stringer.
func (sq *SpaceQuota) String() string {
	var builder strings.Builder
	builder.WriteString("SpaceQuota(")
	builder.WriteString(fmt.Sprintf("id=%v, ", sq.ID))
	builder.WriteString("space_id=")
	builder.WriteString(sq.SpaceID)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(sq.Description)
	builder.WriteString(", ")
	builder.WriteString("extras=")
	builder.WriteString(fmt.Sprintf("%v", sq.Extras))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(sq.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(sq.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(fmt.Sprintf("%v", sq.CreatedAt))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(fmt.Sprintf("%v", sq.UpdatedAt))
	builder.WriteString(", ")
	builder.WriteString("quota_type=")
	builder.WriteString(sq.QuotaType)
	builder.WriteString(", ")
	builder.WriteString("quota_name=")
	builder.WriteString(sq.QuotaName)
	builder.WriteString(", ")
	builder.WriteString("max_value=")
	builder.WriteString(fmt.Sprintf("%v", sq.MaxValue))
	builder.WriteString(", ")
	builder.WriteString("current_used=")
	builder.WriteString(fmt.Sprintf("%v", sq.CurrentUsed))
	builder.WriteString(", ")
	builder.WriteString("unit=")
	builder.WriteString(sq.Unit)
	builder.WriteString(", ")
	builder.WriteString("enabled=")
	builder.WriteString(fmt.Sprintf("%v", sq.Enabled))
	builder.WriteByte(')')
	return builder.String()
}

// SpaceQuotaSlice is a parsable slice of SpaceQuota.
type SpaceQuotaSlice []*SpaceQuota
