package mixin

import (
	"stocms/pkg/nanoid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PrimaryKey adds primary key field.
type PrimaryKey struct{ ent.Schema }

// Fields of the primary key mixin.
func (PrimaryKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Comment("primary key").Immutable().Unique().DefaultFunc(nanoid.PrimaryKey()), // primary key
	}
}

// Indexes of the PrimaryKey.
func (PrimaryKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id"),
	}
}

// primary key mixin must implement `Mixin` interface.
var _ ent.Mixin = (*PrimaryKey)(nil)

// UserPrimaryKeyAlias adds user primary key alias field.
type UserPrimaryKeyAlias struct{ ent.Schema }

// Fields of the user primary key alias mixin.
func (UserPrimaryKeyAlias) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("user_id").Comment("user primary key alias").Immutable().Unique().DefaultFunc(nanoid.PrimaryKey()), // user primary key alias
	}
}

// Indexes of the UserPrimaryKeyAlias.
func (UserPrimaryKeyAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id"),
	}
}

// user primary key alias mixin must implement `Mixin` interface.
var _ ent.Mixin = (*UserPrimaryKeyAlias)(nil)
