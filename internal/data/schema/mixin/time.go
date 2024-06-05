package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// CreatedAt adds a created_at time field to the schema.
type CreatedAt struct{ mixin.Schema }

// Fields of the CreatedAt mixin.
func (CreatedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Immutable().
			Optional().
			Default(time.Now).
			Comment("created at"),
	}
}

// Ensure CreatedAt implements the Mixin interface.
var _ ent.Mixin = (*CreatedAt)(nil)

// UpdatedAt adds an updated_at time field to the schema.
type UpdatedAt struct{ mixin.Schema }

// Fields of the UpdatedAt mixin.
func (UpdatedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("updated_at").
			Optional().
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("updated at"),
	}
}

// Ensure UpdatedAt implements the Mixin interface.
var _ ent.Mixin = (*UpdatedAt)(nil)

// DeletedAt adds a deleted_at time field to the schema.
type DeletedAt struct{ mixin.Schema }

// Fields of the DeletedAt mixin.
func (DeletedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").
			Optional().
			Comment("deleted at"),
	}
}

// Ensure DeletedAt implements the Mixin interface.
var _ ent.Mixin = (*DeletedAt)(nil)

// ExpiredAt adds an expired_at time field to the schema.
type ExpiredAt struct{ mixin.Schema }

// Fields of the ExpiredAt mixin.
func (ExpiredAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("expired_at").
			Optional().
			Comment("expired at"),
	}
}

// Ensure ExpiredAt implements the Mixin interface.
var _ ent.Mixin = (*ExpiredAt)(nil)

// Expires adds an expires time field to the schema.
type Expires struct{ mixin.Schema }

// Fields of the Expires mixin.
func (Expires) Fields() []ent.Field {
	return []ent.Field{
		field.Time("expires").
			Optional().
			Comment("expires"),
	}
}

// Ensure Expires implements the Mixin interface.
var _ ent.Mixin = (*Expires)(nil)

// Released adds a released time field to the schema.
type Released struct{ mixin.Schema }

// Fields of the Released mixin.
func (Released) Fields() []ent.Field {
	return []ent.Field{
		field.Time("released").
			Optional().
			Comment("released"),
	}
}

// Ensure Released implements the Mixin interface.
var _ ent.Mixin = (*Released)(nil)

// TimeAt composes created at and updated at time fields.
type TimeAt struct{ mixin.Schema }

// Fields of the TimeAt mixin.
func (TimeAt) Fields() []ent.Field {
	return append(
		CreatedAt{}.Fields(),
		UpdatedAt{}.Fields()...,
	)
}

// Ensure TimeAt implements the Mixin interface.
var _ ent.Mixin = (*TimeAt)(nil)
