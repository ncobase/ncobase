package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserMFA holds the schema definition for the UserMFA entity.
type UserMFA struct {
	ent.Schema
}

// Annotations of the UserMFA.
func (UserMFA) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "user_mfa"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserMFA.
func (UserMFA) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.TimeAt{},
	}
}

// Fields of the UserMFA.
func (UserMFA) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("enabled").
			Default(false).
			Comment("Whether MFA is enabled for the user"),

		field.String("totp_secret").
			Optional().
			Comment("Encrypted TOTP secret"),

		field.Int64("verified_at").
			Optional().
			Comment("TOTP verified timestamp"),

		field.Int64("last_used_at").
			Optional().
			Comment("Last successful MFA timestamp"),

		field.JSON("recovery_code_hashes", []string{}).
			Optional().
			Comment("SHA-256 hashes of recovery codes"),

		field.Int64("recovery_codes_generated_at").
			Optional().
			Comment("Recovery codes generation timestamp"),

		field.Int("failed_attempts").
			Default(0).
			Comment("Consecutive failed verification attempts"),

		field.Int64("locked_until").
			Optional().
			Comment("MFA verification lock timestamp"),
	}
}

// Edges of the UserMFA.
func (UserMFA) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserMFA.
func (UserMFA) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("enabled"),
	}
}
