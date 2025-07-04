// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// NcseCounterColumns holds the columns for the "ncse_counter" table.
	NcseCounterColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "identifier", Type: field.TypeString, Nullable: true, Comment: "Identifier"},
		{Name: "name", Type: field.TypeString, Nullable: true, Comment: "name"},
		{Name: "prefix", Type: field.TypeString, Nullable: true, Comment: "prefix"},
		{Name: "suffix", Type: field.TypeString, Nullable: true, Comment: "suffix"},
		{Name: "start_value", Type: field.TypeInt, Comment: "Start value", Default: 1},
		{Name: "increment_step", Type: field.TypeInt, Comment: "Increment step", Default: 1},
		{Name: "date_format", Type: field.TypeString, Nullable: true, Comment: "Date format, default YYYYMMDD", Default: "200601"},
		{Name: "current_value", Type: field.TypeInt, Comment: "Current value", Default: 0},
		{Name: "disabled", Type: field.TypeBool, Nullable: true, Comment: "is disabled", Default: false},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "description"},
		{Name: "space_id", Type: field.TypeString, Nullable: true, Comment: "space id, e.g. space id, organization id, store id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
	}
	// NcseCounterTable holds the schema information for the "ncse_counter" table.
	NcseCounterTable = &schema.Table{
		Name:       "ncse_counter",
		Columns:    NcseCounterColumns,
		PrimaryKey: []*schema.Column{NcseCounterColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "counter_id",
				Unique:  true,
				Columns: []*schema.Column{NcseCounterColumns[0]},
			},
			{
				Name:    "counter_space_id",
				Unique:  false,
				Columns: []*schema.Column{NcseCounterColumns[11]},
			},
			{
				Name:    "counter_id_created_at",
				Unique:  true,
				Columns: []*schema.Column{NcseCounterColumns[0], NcseCounterColumns[14]},
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		NcseCounterTable,
	}
)

func init() {
	NcseCounterTable.Annotation = &entsql.Annotation{
		Table: "ncse_counter",
	}
}
