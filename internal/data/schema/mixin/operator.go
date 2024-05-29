package mixin

import (
	"stocms/pkg/nanoid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// CreatedBy adds created by operator field.
type CreatedBy struct{ mixin.Schema }

// Fields of the created by mixin.
func (CreatedBy) Fields() []ent.Field {
	return []ent.Field{
		field.String("created_by").Comment("created by").Optional().MaxLen(nanoid.PrimaryKeySize), // created by
	}
}

// created by mixin must implement `Mixin` interface.
var _ ent.Mixin = (*CreatedBy)(nil)

// UpdatedBy adds updated by operator field.
type UpdatedBy struct{ mixin.Schema }

// Fields of the updated by mixin.
func (UpdatedBy) Fields() []ent.Field {
	return []ent.Field{
		field.String("updated_by").Comment("updated by").Optional().MaxLen(nanoid.PrimaryKeySize), // updated by
	}
}

// updated by mixin must implement `Mixin` interface.
var _ ent.Mixin = (*UpdatedBy)(nil)

// DeletedBy adds deleted by operator field.
type DeletedBy struct{ mixin.Schema }

// Fields of the deleted by mixin.
func (DeletedBy) Fields() []ent.Field {
	return []ent.Field{
		field.String("deleted_by").Comment("deleted by").Optional().MaxLen(nanoid.PrimaryKeySize), // deleted by
	}
}

// deleted by mixin must implement `Mixin` interface.
var _ ent.Mixin = (*DeletedBy)(nil)

type OperatorBy struct{ mixin.Schema }

// Fields of the created at mixin.
func (OperatorBy) Fields() []ent.Field {
	return append(
		CreatedBy{}.Fields(),
		UpdatedBy{}.Fields()...,
	)
}

// operator mixin must implement `Mixin` interface.
var _ ent.Mixin = (*OperatorBy)(nil)
