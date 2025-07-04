// Code generated by ent, DO NOT EDIT.

package spacesetting

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the spacesetting type in the database.
	Label = "space_setting"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldSpaceID holds the string denoting the space_id field in the database.
	FieldSpaceID = "space_id"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldExtras holds the string denoting the extras field in the database.
	FieldExtras = "extras"
	// FieldCreatedBy holds the string denoting the created_by field in the database.
	FieldCreatedBy = "created_by"
	// FieldUpdatedBy holds the string denoting the updated_by field in the database.
	FieldUpdatedBy = "updated_by"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldSettingKey holds the string denoting the setting_key field in the database.
	FieldSettingKey = "setting_key"
	// FieldSettingName holds the string denoting the setting_name field in the database.
	FieldSettingName = "setting_name"
	// FieldSettingValue holds the string denoting the setting_value field in the database.
	FieldSettingValue = "setting_value"
	// FieldDefaultValue holds the string denoting the default_value field in the database.
	FieldDefaultValue = "default_value"
	// FieldSettingType holds the string denoting the setting_type field in the database.
	FieldSettingType = "setting_type"
	// FieldScope holds the string denoting the scope field in the database.
	FieldScope = "scope"
	// FieldCategory holds the string denoting the category field in the database.
	FieldCategory = "category"
	// FieldIsPublic holds the string denoting the is_public field in the database.
	FieldIsPublic = "is_public"
	// FieldIsRequired holds the string denoting the is_required field in the database.
	FieldIsRequired = "is_required"
	// FieldIsReadonly holds the string denoting the is_readonly field in the database.
	FieldIsReadonly = "is_readonly"
	// FieldValidation holds the string denoting the validation field in the database.
	FieldValidation = "validation"
	// Table holds the table name of the spacesetting in the database.
	Table = "ncse_sys_space_setting"
)

// Columns holds all SQL columns for spacesetting fields.
var Columns = []string{
	FieldID,
	FieldSpaceID,
	FieldDescription,
	FieldExtras,
	FieldCreatedBy,
	FieldUpdatedBy,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldSettingKey,
	FieldSettingName,
	FieldSettingValue,
	FieldDefaultValue,
	FieldSettingType,
	FieldScope,
	FieldCategory,
	FieldIsPublic,
	FieldIsRequired,
	FieldIsReadonly,
	FieldValidation,
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
	// DefaultExtras holds the default value on creation for the "extras" field.
	DefaultExtras map[string]interface{}
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() int64
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() int64
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() int64
	// SettingKeyValidator is a validator for the "setting_key" field. It is called by the builders before save.
	SettingKeyValidator func(string) error
	// SettingNameValidator is a validator for the "setting_name" field. It is called by the builders before save.
	SettingNameValidator func(string) error
	// DefaultSettingType holds the default value on creation for the "setting_type" field.
	DefaultSettingType string
	// DefaultScope holds the default value on creation for the "scope" field.
	DefaultScope string
	// DefaultCategory holds the default value on creation for the "category" field.
	DefaultCategory string
	// DefaultIsPublic holds the default value on creation for the "is_public" field.
	DefaultIsPublic bool
	// DefaultIsRequired holds the default value on creation for the "is_required" field.
	DefaultIsRequired bool
	// DefaultIsReadonly holds the default value on creation for the "is_readonly" field.
	DefaultIsReadonly bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
	// IDValidator is a validator for the "id" field. It is called by the builders before save.
	IDValidator func(string) error
)

// OrderOption defines the ordering options for the SpaceSetting queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// BySpaceID orders the results by the space_id field.
func BySpaceID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSpaceID, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
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

// BySettingKey orders the results by the setting_key field.
func BySettingKey(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSettingKey, opts...).ToFunc()
}

// BySettingName orders the results by the setting_name field.
func BySettingName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSettingName, opts...).ToFunc()
}

// BySettingValue orders the results by the setting_value field.
func BySettingValue(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSettingValue, opts...).ToFunc()
}

// ByDefaultValue orders the results by the default_value field.
func ByDefaultValue(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDefaultValue, opts...).ToFunc()
}

// BySettingType orders the results by the setting_type field.
func BySettingType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSettingType, opts...).ToFunc()
}

// ByScope orders the results by the scope field.
func ByScope(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldScope, opts...).ToFunc()
}

// ByCategory orders the results by the category field.
func ByCategory(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCategory, opts...).ToFunc()
}

// ByIsPublic orders the results by the is_public field.
func ByIsPublic(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsPublic, opts...).ToFunc()
}

// ByIsRequired orders the results by the is_required field.
func ByIsRequired(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsRequired, opts...).ToFunc()
}

// ByIsReadonly orders the results by the is_readonly field.
func ByIsReadonly(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsReadonly, opts...).ToFunc()
}
