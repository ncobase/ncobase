// Code generated by ent, DO NOT EDIT.

package menu

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the menu type in the database.
	Label = "menu"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldLabel holds the string denoting the label field in the database.
	FieldLabel = "label"
	// FieldSlug holds the string denoting the slug field in the database.
	FieldSlug = "slug"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldPath holds the string denoting the path field in the database.
	FieldPath = "path"
	// FieldTarget holds the string denoting the target field in the database.
	FieldTarget = "target"
	// FieldIcon holds the string denoting the icon field in the database.
	FieldIcon = "icon"
	// FieldPerms holds the string denoting the perms field in the database.
	FieldPerms = "perms"
	// FieldHidden holds the string denoting the hidden field in the database.
	FieldHidden = "hidden"
	// FieldOrder holds the string denoting the order field in the database.
	FieldOrder = "order"
	// FieldDisabled holds the string denoting the disabled field in the database.
	FieldDisabled = "disabled"
	// FieldExtras holds the string denoting the extras field in the database.
	FieldExtras = "extras"
	// FieldParentID holds the string denoting the parent_id field in the database.
	FieldParentID = "parent_id"
	// FieldCreatedBy holds the string denoting the created_by field in the database.
	FieldCreatedBy = "created_by"
	// FieldUpdatedBy holds the string denoting the updated_by field in the database.
	FieldUpdatedBy = "updated_by"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// Table holds the table name of the menu in the database.
	Table = "ncse_sys_menu"
)

// Columns holds all SQL columns for menu fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldLabel,
	FieldSlug,
	FieldType,
	FieldPath,
	FieldTarget,
	FieldIcon,
	FieldPerms,
	FieldHidden,
	FieldOrder,
	FieldDisabled,
	FieldExtras,
	FieldParentID,
	FieldCreatedBy,
	FieldUpdatedBy,
	FieldCreatedAt,
	FieldUpdatedAt,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultHidden holds the default value on creation for the "hidden" field.
	DefaultHidden bool
	// DefaultOrder holds the default value on creation for the "order" field.
	DefaultOrder int
	// DefaultDisabled holds the default value on creation for the "disabled" field.
	DefaultDisabled bool
	// DefaultExtras holds the default value on creation for the "extras" field.
	DefaultExtras map[string]interface{}
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() int64
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() int64
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() int64
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
	// IDValidator is a validator for the "id" field. It is called by the builders before save.
	IDValidator func(string) error
)

// OrderOption defines the ordering options for the Menu queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByLabel orders the results by the label field.
func ByLabel(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLabel, opts...).ToFunc()
}

// BySlug orders the results by the slug field.
func BySlug(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSlug, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
}

// ByPath orders the results by the path field.
func ByPath(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPath, opts...).ToFunc()
}

// ByTarget orders the results by the target field.
func ByTarget(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTarget, opts...).ToFunc()
}

// ByIcon orders the results by the icon field.
func ByIcon(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIcon, opts...).ToFunc()
}

// ByPerms orders the results by the perms field.
func ByPerms(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPerms, opts...).ToFunc()
}

// ByHidden orders the results by the hidden field.
func ByHidden(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHidden, opts...).ToFunc()
}

// ByOrder orders the results by the order field.
func ByOrder(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOrder, opts...).ToFunc()
}

// ByDisabled orders the results by the disabled field.
func ByDisabled(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDisabled, opts...).ToFunc()
}

// ByParentID orders the results by the parent_id field.
func ByParentID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldParentID, opts...).ToFunc()
}

// ByCreatedBy orders the results by the created_by field.
func ByCreatedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedBy, opts...).ToFunc()
}

// ByUpdatedBy orders the results by the updated_by field.
func ByUpdatedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedBy, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}
