package schema

import (
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/entgo/mixin"
)

// Employee holds the schema definition for the Employee entity.
type Employee struct {
	ent.Schema
}

// Annotations of the Employee.
func (Employee) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "employee"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Employee.
func (Employee) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("user", "user_id"),
		mixin.SpaceID,
		mixin.TimeAt{},
	}
}

// Fields of the Employee.
func (Employee) Fields() []ent.Field {
	return []ent.Field{
		field.String("employee_id").Optional().Comment("Employee ID/Number"),
		field.String("department").Optional().Comment("Primary department"),
		field.String("position").Optional().Comment("Job position/title"),
		field.String("manager_id").Optional().Comment("Direct manager user ID"),
		field.Time("hire_date").Optional().Comment("Hire date"),
		field.Time("termination_date").Optional().Nillable().Comment("Termination date"),
		field.Enum("employment_type").Values("full_time", "part_time", "contract", "intern").Default("full_time").Comment("Employment type"),
		field.Enum("status").Values("active", "inactive", "on_leave", "terminated").Default("active").Comment("Employee status"),
		field.Float("salary").Optional().Comment("Base salary"),
		field.String("work_location").Optional().Comment("Primary work location"),
		field.JSON("contact_info", map[string]any{}).Optional().Comment("Emergency contact info"),
		field.JSON("skills", []string{}).Optional().Comment("Employee skills"),
		field.JSON("certifications", []string{}).Optional().Comment("Professional certifications"),
		field.JSON("extras", map[string]any{}).Optional().Comment("Additional employee data"),
	}
}

// Edges of the Employee.
func (Employee) Edges() []ent.Edge {
	return nil
}

// Indexes of the Employee.
func (Employee) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "space_id").Unique(),
		index.Fields("employee_id"),
		index.Fields("department"),
		index.Fields("manager_id"),
		index.Fields("status"),
	}
}
