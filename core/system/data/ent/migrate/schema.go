// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// NcseSysDictionaryColumns holds the columns for the "ncse_sys_dictionary" table.
	NcseSysDictionaryColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "name", Type: field.TypeString, Nullable: true, Comment: "name"},
		{Name: "slug", Type: field.TypeString, Unique: true, Nullable: true, Comment: "slug / alias"},
		{Name: "type", Type: field.TypeString, Nullable: true, Comment: "type"},
		{Name: "value", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "value"},
		{Name: "description", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "description"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
	}
	// NcseSysDictionaryTable holds the schema information for the "ncse_sys_dictionary" table.
	NcseSysDictionaryTable = &schema.Table{
		Name:       "ncse_sys_dictionary",
		Columns:    NcseSysDictionaryColumns,
		PrimaryKey: []*schema.Column{NcseSysDictionaryColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "dictionary_id",
				Unique:  true,
				Columns: []*schema.Column{NcseSysDictionaryColumns[0]},
			},
			{
				Name:    "dictionary_slug",
				Unique:  true,
				Columns: []*schema.Column{NcseSysDictionaryColumns[2]},
			},
		},
	}
	// NcseSysMenuColumns holds the columns for the "ncse_sys_menu" table.
	NcseSysMenuColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "name", Type: field.TypeString, Nullable: true, Comment: "name"},
		{Name: "label", Type: field.TypeString, Nullable: true, Comment: "label"},
		{Name: "slug", Type: field.TypeString, Unique: true, Nullable: true, Comment: "slug / alias"},
		{Name: "type", Type: field.TypeString, Nullable: true, Comment: "type"},
		{Name: "path", Type: field.TypeString, Nullable: true, Comment: "path"},
		{Name: "target", Type: field.TypeString, Nullable: true, Comment: "target"},
		{Name: "icon", Type: field.TypeString, Nullable: true, Comment: "icon"},
		{Name: "perms", Type: field.TypeString, Nullable: true, Comment: "perms"},
		{Name: "hidden", Type: field.TypeBool, Nullable: true, Comment: "is hidden", Default: false},
		{Name: "order", Type: field.TypeInt, Comment: "display order", Default: 0},
		{Name: "disabled", Type: field.TypeBool, Nullable: true, Comment: "is disabled", Default: false},
		{Name: "extras", Type: field.TypeJSON, Nullable: true, Comment: "Extend properties"},
		{Name: "parent_id", Type: field.TypeString, Nullable: true, Comment: "parent id"},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
	}
	// NcseSysMenuTable holds the schema information for the "ncse_sys_menu" table.
	NcseSysMenuTable = &schema.Table{
		Name:       "ncse_sys_menu",
		Columns:    NcseSysMenuColumns,
		PrimaryKey: []*schema.Column{NcseSysMenuColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "menu_id",
				Unique:  true,
				Columns: []*schema.Column{NcseSysMenuColumns[0]},
			},
			{
				Name:    "menu_slug",
				Unique:  true,
				Columns: []*schema.Column{NcseSysMenuColumns[3]},
			},
			{
				Name:    "menu_parent_id",
				Unique:  false,
				Columns: []*schema.Column{NcseSysMenuColumns[13]},
			},
			{
				Name:    "menu_id_created_at",
				Unique:  true,
				Columns: []*schema.Column{NcseSysMenuColumns[0], NcseSysMenuColumns[16]},
			},
		},
	}
	// NcseSysOptionColumns holds the columns for the "ncse_sys_option" table.
	NcseSysOptionColumns = []*schema.Column{
		{Name: "id", Type: field.TypeString, Unique: true, Size: 16, Comment: "primary key"},
		{Name: "name", Type: field.TypeString, Unique: true, Nullable: true, Comment: "name"},
		{Name: "type", Type: field.TypeString, Nullable: true, Comment: "type"},
		{Name: "value", Type: field.TypeString, Nullable: true, Size: 2147483647, Comment: "value"},
		{Name: "autoload", Type: field.TypeBool, Nullable: true, Comment: "Whether to load the option automatically", Default: true},
		{Name: "created_by", Type: field.TypeString, Nullable: true, Comment: "id of the creator"},
		{Name: "updated_by", Type: field.TypeString, Nullable: true, Comment: "id of the last updater"},
		{Name: "created_at", Type: field.TypeInt64, Nullable: true, Comment: "created at"},
		{Name: "updated_at", Type: field.TypeInt64, Nullable: true, Comment: "updated at"},
	}
	// NcseSysOptionTable holds the schema information for the "ncse_sys_option" table.
	NcseSysOptionTable = &schema.Table{
		Name:       "ncse_sys_option",
		Columns:    NcseSysOptionColumns,
		PrimaryKey: []*schema.Column{NcseSysOptionColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "options_id",
				Unique:  true,
				Columns: []*schema.Column{NcseSysOptionColumns[0]},
			},
			{
				Name:    "options_name",
				Unique:  true,
				Columns: []*schema.Column{NcseSysOptionColumns[1]},
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		NcseSysDictionaryTable,
		NcseSysMenuTable,
		NcseSysOptionTable,
	}
)

func init() {
	NcseSysDictionaryTable.Annotation = &entsql.Annotation{
		Table: "ncse_sys_dictionary",
	}
	NcseSysMenuTable.Annotation = &entsql.Annotation{
		Table: "ncse_sys_menu",
	}
	NcseSysOptionTable.Annotation = &entsql.Annotation{
		Table: "ncse_sys_option",
	}
}
