package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Status adds status field.
type Status struct{ mixin.Schema }

// Fields of the status mixin.
func (Status) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("status").Comment("status").Default(1).Positive(), // status
	}
}

// status mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Status)(nil)

// Order adds order field.
type Order struct{ mixin.Schema }

// Fields of the order mixin.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("order").Comment("order").Default(99).Positive(), // order
	}
}

// order mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Order)(nil)

// Size adds size field.
type Size struct{ mixin.Schema }

// Fields of the size mixin.
func (Size) Fields() []ent.Field {
	return []ent.Field{
		field.Int("size").Comment("size, byte").Default(0), // size
	}
}

// size mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Size)(nil)
