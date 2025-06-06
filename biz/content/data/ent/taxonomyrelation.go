// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"ncobase/content/data/ent/taxonomyrelation"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// TaxonomyRelation is the model entity for the TaxonomyRelation schema.
type TaxonomyRelation struct {
	config `json:"-"`
	// ID of the ent.
	// primary key
	ID string `json:"id,omitempty"`
	// object id
	ObjectID string `json:"object_id,omitempty"`
	// taxonomy id
	TaxonomyID string `json:"taxonomy_id,omitempty"`
	// type
	Type string `json:"type,omitempty"`
	// display order
	Order int `json:"order,omitempty"`
	// id of the creator
	CreatedBy string `json:"created_by,omitempty"`
	// created at
	CreatedAt    int64 `json:"created_at,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*TaxonomyRelation) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case taxonomyrelation.FieldOrder, taxonomyrelation.FieldCreatedAt:
			values[i] = new(sql.NullInt64)
		case taxonomyrelation.FieldID, taxonomyrelation.FieldObjectID, taxonomyrelation.FieldTaxonomyID, taxonomyrelation.FieldType, taxonomyrelation.FieldCreatedBy:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the TaxonomyRelation fields.
func (tr *TaxonomyRelation) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case taxonomyrelation.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				tr.ID = value.String
			}
		case taxonomyrelation.FieldObjectID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field object_id", values[i])
			} else if value.Valid {
				tr.ObjectID = value.String
			}
		case taxonomyrelation.FieldTaxonomyID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field taxonomy_id", values[i])
			} else if value.Valid {
				tr.TaxonomyID = value.String
			}
		case taxonomyrelation.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				tr.Type = value.String
			}
		case taxonomyrelation.FieldOrder:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field order", values[i])
			} else if value.Valid {
				tr.Order = int(value.Int64)
			}
		case taxonomyrelation.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				tr.CreatedBy = value.String
			}
		case taxonomyrelation.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				tr.CreatedAt = value.Int64
			}
		default:
			tr.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the TaxonomyRelation.
// This includes values selected through modifiers, order, etc.
func (tr *TaxonomyRelation) Value(name string) (ent.Value, error) {
	return tr.selectValues.Get(name)
}

// Update returns a builder for updating this TaxonomyRelation.
// Note that you need to call TaxonomyRelation.Unwrap() before calling this method if this TaxonomyRelation
// was returned from a transaction, and the transaction was committed or rolled back.
func (tr *TaxonomyRelation) Update() *TaxonomyRelationUpdateOne {
	return NewTaxonomyRelationClient(tr.config).UpdateOne(tr)
}

// Unwrap unwraps the TaxonomyRelation entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (tr *TaxonomyRelation) Unwrap() *TaxonomyRelation {
	_tx, ok := tr.config.driver.(*txDriver)
	if !ok {
		panic("ent: TaxonomyRelation is not a transactional entity")
	}
	tr.config.driver = _tx.drv
	return tr
}

// String implements the fmt.Stringer.
func (tr *TaxonomyRelation) String() string {
	var builder strings.Builder
	builder.WriteString("TaxonomyRelation(")
	builder.WriteString(fmt.Sprintf("id=%v, ", tr.ID))
	builder.WriteString("object_id=")
	builder.WriteString(tr.ObjectID)
	builder.WriteString(", ")
	builder.WriteString("taxonomy_id=")
	builder.WriteString(tr.TaxonomyID)
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(tr.Type)
	builder.WriteString(", ")
	builder.WriteString("order=")
	builder.WriteString(fmt.Sprintf("%v", tr.Order))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(tr.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(fmt.Sprintf("%v", tr.CreatedAt))
	builder.WriteByte(')')
	return builder.String()
}

// TaxonomyRelations is a parsable slice of TaxonomyRelation.
type TaxonomyRelations []*TaxonomyRelation
