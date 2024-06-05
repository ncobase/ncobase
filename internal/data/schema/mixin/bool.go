package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Default adds a default field to the schema.
type Default struct{ mixin.Schema }

// Fields of the Default mixin.
func (Default) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("default").Optional().Comment("is default"),
	}
}

// Ensure Default implements the Mixin interface.
var _ ent.Mixin = (*Default)(nil)

// Markdown adds a markdown field to the schema.
type Markdown struct{ mixin.Schema }

// Fields of the Markdown mixin.
func (Markdown) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("markdown").Optional().Comment("is markdown"),
	}
}

// Ensure Markdown implements the Mixin interface.
var _ ent.Mixin = (*Markdown)(nil)

// Temp adds a temp field to the schema.
type Temp struct{ mixin.Schema }

// Fields of the Temp mixin.
func (Temp) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("temp").Optional().Comment("is temp"),
	}
}

// Ensure Temp implements the Mixin interface.
var _ ent.Mixin = (*Temp)(nil)

// Private adds a private field to the schema.
type Private struct{ mixin.Schema }

// Fields of the Private mixin.
func (Private) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("private").Optional().Comment("is private"),
	}
}

// Ensure Private implements the Mixin interface.
var _ ent.Mixin = (*Private)(nil)

// Approved adds an approved field to the schema.
type Approved struct{ mixin.Schema }

// Fields of the Approved mixin.
func (Approved) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("approved").Optional().Comment("is approved"),
	}
}

// Ensure Approved implements the Mixin interface.
var _ ent.Mixin = (*Approved)(nil)

// Disabled adds a disabled field to the schema.
type Disabled struct{ mixin.Schema }

// Fields of the Disabled mixin.
func (Disabled) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("disabled").Optional().Comment("is disabled"),
	}
}

// Ensure Disabled implements the Mixin interface.
var _ ent.Mixin = (*Disabled)(nil)

// Logged adds a logged field to the schema.
type Logged struct{ mixin.Schema }

// Fields of the Logged mixin.
func (Logged) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("logged").Optional().Comment("is logged"),
	}
}

// Ensure Logged implements the Mixin interface.
var _ ent.Mixin = (*Logged)(nil)

// System adds a system field to the schema.
type System struct{ mixin.Schema }

// Fields of the System mixin.
func (System) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("system").Optional().Comment("is system"),
	}
}

// Ensure System implements the Mixin interface.
var _ ent.Mixin = (*System)(nil)

// Hidden adds a hidden field to the schema.
type Hidden struct{ mixin.Schema }

// Fields of the Hidden mixin.
func (Hidden) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("hidden").Optional().Comment("is hidden"),
	}
}

// Ensure Hidden implements the Mixin interface.
var _ ent.Mixin = (*Hidden)(nil)

// IsCertified adds an is_certified field to the schema.
type IsCertified struct{ mixin.Schema }

// Fields of the IsCertified mixin.
func (IsCertified) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("is_certified").Optional().Comment("is certified"),
	}
}

// Ensure IsCertified implements the Mixin interface.
var _ ent.Mixin = (*IsCertified)(nil)

// IsAdmin adds an is_admin field to the schema.
type IsAdmin struct{ mixin.Schema }

// Fields of the IsAdmin mixin.
func (IsAdmin) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("is_admin").Optional().Comment("is admin"),
	}
}

// Ensure IsAdmin implements the Mixin interface.
var _ ent.Mixin = (*IsAdmin)(nil)
