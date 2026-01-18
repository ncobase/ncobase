package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"
	"github.com/ncobase/ncore/utils/nanoid"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

// Annotations of the Session.
func (Session) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "user_session"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Session.
func (Session) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.CustomPrimaryKey{
			Length:      64,
			DefaultFunc: func() string { return nanoid.Must(64) },
		},
		mixin.UserID,
		mixin.TimeAt{},
	}
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("token_id").Comment("Associated auth token ID"),
		field.JSON("device_info", map[string]any{}).Optional().Comment("Device information"),
		field.String("ip_address").Optional().Comment("IP address"),
		field.String("user_agent").Optional().Comment("User agent string"),
		field.String("location").Optional().Comment("Geographic location"),
		field.String("login_method").Optional().Comment("Login method used"),
		field.Bool("is_active").Default(true).Comment("Session active status"),
		field.Int64("last_access_at").Optional().Comment("Last access timestamp"),
		field.Int64("expires_at").Optional().Comment("Session expiration timestamp"),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return nil
}

// Indexes of the Session.
func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("token_id"),
		index.Fields("is_active"),
		index.Fields("expires_at"),
	}
}
