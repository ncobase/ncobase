package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Default adds default field.
type Default struct{ mixin.Schema }

// Fields of the default mixin.
func (Default) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("default").Comment("is default").Optional(), // is default
	}
}

// default mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Default)(nil)

// Markdown adds markdown field.
type Markdown struct{ mixin.Schema }

// Fields of the markdown mixin.
func (Markdown) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("markdown").Comment("is markdown").Optional(), // is markdown
	}
}

// markdown mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Markdown)(nil)

// Temp adds temp field.
type Temp struct{ mixin.Schema }

// Fields of the temp mixin.
func (Temp) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("temp").Comment("is temp").Optional(), // is temp
	}
}

// temp mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Temp)(nil)

// Private adds private field.
type Private struct{ mixin.Schema }

// Fields of the private mixin.
func (Private) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("private").Comment("is private").Optional(), // is private
	}
}

// private mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Private)(nil)

// Approved adds approved field.
type Approved struct{ mixin.Schema }

// Fields of the approved mixin.
func (Approved) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("approved").Comment("is approved").Optional(), // is approved
	}
}

// approved mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Approved)(nil)

// Disabled adds disabled field.
type Disabled struct{ mixin.Schema }

// Fields of the disabled mixin.
func (Disabled) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("disabled").Comment("is disabled").Optional(), // is disabled
	}
}

// disabled mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Disabled)(nil)

// System adds system field.
type System struct{ mixin.Schema }

// Fields of the system mixin.
func (System) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("system").Comment("is system").Optional(), // is system
	}
}

// system mixin must implement `Mixin` interface.
var _ ent.Mixin = (*System)(nil)

// Hidden adds hidden field.
type Hidden struct{ mixin.Schema }

// Fields of the hidden mixin.
func (Hidden) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("hidden").Comment("is hidden").Optional(), // is hidden
	}
}

// hidden mixin must implement `Mixin` interface.
var _ ent.Mixin = (*Hidden)(nil)
