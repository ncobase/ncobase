package mixin

import (
	"stocms/pkg/types"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// ExtraProps adds extend properties field.
type ExtraProps struct{ ent.Schema }

// Fields of the extend properties mixin.
func (ExtraProps) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("extras", types.JSON{}).Default(types.JSON{}).Optional().Comment("extend properties"), // extend info etc...
	}
}

// extra properties mixin must implement `Mixin` interface.
var _ ent.Mixin = (*ExtraProps)(nil)

// Author adds author field.
type Author struct{ ent.Schema }

// Fields of the author mixin.
func (Author) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("author", types.JSON{}).Default(types.JSON{}).Optional().Comment("author, default is empty json, options: {id: '', name: '', avatar: '', url: '', email: '', ip: ''}"), // author info etc...
	}
}

// author mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Author)(nil)

// Related adds related field.
type Related struct{ ent.Schema }

// Fields of the related mixin.
func (Related) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("related", types.JSON{}).Default(types.JSON{}).Optional().Comment("related, default is empty json, options: {id: '', name: '', type: 'user / topic /...'}"), // related info etc...
	}
}

// related mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Related)(nil)

// Leader adds leader field.
type Leader struct{ ent.Schema }

// Fields of the leader mixin.
func (Leader) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("leader", types.JSON{}).Default(types.JSON{}).Optional().Comment("leader, default is empty json, eg: {id: '', name: '', avatar: '', url: '', email: '', ip: ''}"), // leader info etc...
	}
}

// leader mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Leader)(nil)

// Links adds links field.
type Links struct{ ent.Schema }

// Fields of the links mixin.
func (Links) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("links", types.JSONArray{}).Default(types.JSONArray{}).Comment("social link / more profile").Optional(), // object array
	}
}

// links mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Links)(nil)
