package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Status adds a status field to the schema.
type Status struct{ mixin.Schema }

// Fields of the Status mixin.
func (Status) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("status").
			Default(0).
			Comment("status: 0 activated, 1 unactivated, 2 disabled"),
	}
}

// Ensure Status implements the Mixin interface.
var _ ent.Mixin = (*Status)(nil)

// Order adds an order field to the schema.
type Order struct{ mixin.Schema }

// Fields of the Order mixin.
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("order").
			Default(99).
			Positive().
			Comment("display order"),
	}
}

// Ensure Order implements the Mixin interface.
var _ ent.Mixin = (*Order)(nil)

// Size adds a size field to the schema.
type Size struct{ mixin.Schema }

// Fields of the Size mixin.
func (Size) Fields() []ent.Field {
	return []ent.Field{
		field.Int("size").
			Default(0).
			Comment("size in bytes"),
	}
}

// Ensure Size implements the Mixin interface.
var _ ent.Mixin = (*Size)(nil)
