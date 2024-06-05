package mixin

import (
	"stocms/pkg/types"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// ExtraProps adds extend properties field.
// This mixin adds an `extras` field to store additional JSON properties.
type ExtraProps struct{ ent.Schema }

// Fields of the ExtraProps mixin.
func (ExtraProps) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("extras", types.JSON{}).
			Default(types.JSON{}).
			Optional().
			Comment("Extend properties"),
	}
}

// Ensure ExtraProps implements the Mixin interface.
var _ ent.Mixin = (*ExtraProps)(nil)

// Author adds an author field.
// This mixin adds an `author` field to store author information as JSON.
type Author struct{ ent.Schema }

// Fields of the Author mixin.
func (Author) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("author", types.JSON{}).
			Default(types.JSON{}).
			Optional().
			Comment("Author information, e.g., {id: '', name: '', avatar: '', url: '', email: '', ip: ''}"),
	}
}

// Ensure Author implements the Mixin interface.
var _ ent.Mixin = (*Author)(nil)

// Related adds a related field.
// This mixin adds a `related` field to store related entity information as JSON.
type Related struct{ ent.Schema }

// Fields of the Related mixin.
func (Related) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("related", types.JSON{}).
			Default(types.JSON{}).
			Optional().
			Comment("Related entity information, e.g., {id: '', name: '', type: 'user / topic /...'}"),
	}
}

// Ensure Related implements the Mixin interface.
var _ ent.Mixin = (*Related)(nil)

// Leader adds a leader field.
// This mixin adds a `leader` field to store leader information as JSON.
type Leader struct{ ent.Schema }

// Fields of the Leader mixin.
func (Leader) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("leader", types.JSON{}).
			Default(types.JSON{}).
			Optional().
			Comment("Leader information, e.g., {id: '', name: '', avatar: '', url: '', email: '', ip: ''}"),
	}
}

// Ensure Leader implements the Mixin interface.
var _ ent.Mixin = (*Leader)(nil)

// Links adds links field.
// This mixin adds a `links` field to store a list of links as a JSON array.
type Links struct{ ent.Schema }

// Fields of the Links mixin.
func (Links) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("links", types.JSONArray{}).
			Default(types.JSONArray{}).
			Optional().
			Comment("List of social links or profile links"),
	}
}

// Ensure Links implements the Mixin interface.
var _ ent.Mixin = (*Links)(nil)
