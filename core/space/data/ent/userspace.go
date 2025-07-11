// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"ncobase/space/data/ent/userspace"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// UserSpace is the model entity for the UserSpace schema.
type UserSpace struct {
	config `json:"-"`
	// ID of the ent.
	// primary key
	ID string `json:"id,omitempty"`
	// user id
	UserID string `json:"user_id,omitempty"`
	// space id, e.g. space id, organization id, store id
	SpaceID string `json:"space_id,omitempty"`
	// id of the creator
	CreatedBy string `json:"created_by,omitempty"`
	// id of the last updater
	UpdatedBy string `json:"updated_by,omitempty"`
	// created at
	CreatedAt int64 `json:"created_at,omitempty"`
	// updated at
	UpdatedAt    int64 `json:"updated_at,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*UserSpace) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case userspace.FieldCreatedAt, userspace.FieldUpdatedAt:
			values[i] = new(sql.NullInt64)
		case userspace.FieldID, userspace.FieldUserID, userspace.FieldSpaceID, userspace.FieldCreatedBy, userspace.FieldUpdatedBy:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the UserSpace fields.
func (us *UserSpace) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case userspace.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				us.ID = value.String
			}
		case userspace.FieldUserID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value.Valid {
				us.UserID = value.String
			}
		case userspace.FieldSpaceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field space_id", values[i])
			} else if value.Valid {
				us.SpaceID = value.String
			}
		case userspace.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				us.CreatedBy = value.String
			}
		case userspace.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				us.UpdatedBy = value.String
			}
		case userspace.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				us.CreatedAt = value.Int64
			}
		case userspace.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				us.UpdatedAt = value.Int64
			}
		default:
			us.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the UserSpace.
// This includes values selected through modifiers, order, etc.
func (us *UserSpace) Value(name string) (ent.Value, error) {
	return us.selectValues.Get(name)
}

// Update returns a builder for updating this UserSpace.
// Note that you need to call UserSpace.Unwrap() before calling this method if this UserSpace
// was returned from a transaction, and the transaction was committed or rolled back.
func (us *UserSpace) Update() *UserSpaceUpdateOne {
	return NewUserSpaceClient(us.config).UpdateOne(us)
}

// Unwrap unwraps the UserSpace entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (us *UserSpace) Unwrap() *UserSpace {
	_tx, ok := us.config.driver.(*txDriver)
	if !ok {
		panic("ent: UserSpace is not a transactional entity")
	}
	us.config.driver = _tx.drv
	return us
}

// String implements the fmt.Stringer.
func (us *UserSpace) String() string {
	var builder strings.Builder
	builder.WriteString("UserSpace(")
	builder.WriteString(fmt.Sprintf("id=%v, ", us.ID))
	builder.WriteString("user_id=")
	builder.WriteString(us.UserID)
	builder.WriteString(", ")
	builder.WriteString("space_id=")
	builder.WriteString(us.SpaceID)
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(us.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(us.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(fmt.Sprintf("%v", us.CreatedAt))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(fmt.Sprintf("%v", us.UpdatedAt))
	builder.WriteByte(')')
	return builder.String()
}

// UserSpaces is a parsable slice of UserSpace.
type UserSpaces []*UserSpace
