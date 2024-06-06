package mixin

import (
	"stocms/pkg/nanoid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// CreatedBy adds a "created_by" field to store the creator's ID.
type CreatedBy struct{ mixin.Schema }

// Fields of the CreatedBy mixin.
func (CreatedBy) Fields() []ent.Field {
	return []ent.Field{
		field.String("created_by").
			Optional().
			MaxLen(nanoid.PrimaryKeySize).
			Comment("ID of the creator"),
	}
}

// UpdatedBy adds an "updated_by" field to store the ID of the person who last updated the entity.
type UpdatedBy struct{ mixin.Schema }

// Fields of the UpdatedBy mixin.
func (UpdatedBy) Fields() []ent.Field {
	return []ent.Field{
		field.String("updated_by").
			Optional().
			MaxLen(nanoid.PrimaryKeySize).
			Comment("ID of the person who last updated the entity"),
	}
}

// DeletedBy adds a "deleted_by" field to store the ID of the person who deleted the entity.
type DeletedBy struct{ mixin.Schema }

// Fields of the DeletedBy mixin.
func (DeletedBy) Fields() []ent.Field {
	return []ent.Field{
		field.String("deleted_by").
			Optional().
			MaxLen(nanoid.PrimaryKeySize).
			Comment("ID of the person who deleted the entity"),
	}
}

// OperatorBy combines CreatedBy, UpdatedBy, and DeletedBy fields into a single mixin.
type OperatorBy struct{ mixin.Schema }

// Fields of the OperatorBy mixin.
func (OperatorBy) Fields() []ent.Field {
	return append(
		CreatedBy{}.Fields(),
		UpdatedBy{}.Fields()...,
	)
}

// Ensure OperatorBy implements the Mixin interface.
var _ ent.Mixin = (*OperatorBy)(nil)
