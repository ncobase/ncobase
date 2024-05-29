package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// CreatedAt adds created at time field.
type CreatedAt struct{ mixin.Schema }

// Fields of the created at mixin.
func (CreatedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Comment("created at").Immutable().Optional().Default(time.Now), // created at
	}
}

// created at mixin must implement `Mixin` interface.
var _ ent.Mixin = (*CreatedAt)(nil)

// UpdatedAt adds updated at time field.
type UpdatedAt struct{ mixin.Schema }

// Fields of the updated at mixin.
func (UpdatedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("updated_at").Comment("updated at").Optional().Default(time.Now).UpdateDefault(time.Now), // updated at
	}
}

// updated at mixin must implement `Mixin` interface.
var _ ent.Mixin = (*UpdatedAt)(nil)

// DeletedAt adds updated at time field.
type DeletedAt struct{ mixin.Schema }

// Fields of the deleted at mixin.
func (DeletedAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").Comment("deleted at").Optional(), // deleted at
	}
}

// deleted at mixin must implement `Mixin` interface.
var _ ent.Mixin = (*DeletedAt)(nil)

// expired at mixin must implement `Mixin` interface.
var _ ent.Mixin = (*UpdatedAt)(nil)

// ExpiredAt adds expired at time field.
type ExpiredAt struct{ mixin.Schema }

// Fields of the expired at mixin.
func (ExpiredAt) Fields() []ent.Field {
	return []ent.Field{
		field.Time("expired_at").Comment("expired at").Optional(), // expired at
	}
}

// expired at mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ExpiredAt)(nil)

// Expires adds expires field.
type Expires struct{ mixin.Schema }

// Fields of the expires mixin.
func (Expires) Fields() []ent.Field {
	return []ent.Field{
		field.Time("expires").Comment("expires").Optional(), // expires
	}
}

// expires mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Expires)(nil)

// Released adds released time field.
type Released struct{ mixin.Schema }

// Fields of the released mixin.
func (Released) Fields() []ent.Field {
	return []ent.Field{
		field.Time("released").Comment("released").Optional(), // released
	}
}

// released mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Released)(nil)

// TimeAt composes created by / updated by time at mixin.
type TimeAt struct{ mixin.Schema }

// Fields of the time at mixin.
func (TimeAt) Fields() []ent.Field {
	return append(
		CreatedAt{}.Fields(),
		UpdatedAt{}.Fields()...,
	)
}

// time at mixin must implement `Mixin` interface.
var _ ent.Mixin = (*TimeAt)(nil)
