// Code generated by ent, DO NOT EDIT.

package taxonomy

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the taxonomy type in the database.
	Label = "taxonomy"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldSlug holds the string denoting the slug field in the database.
	FieldSlug = "slug"
	// FieldCover holds the string denoting the cover field in the database.
	FieldCover = "cover"
	// FieldThumbnail holds the string denoting the thumbnail field in the database.
	FieldThumbnail = "thumbnail"
	// FieldColor holds the string denoting the color field in the database.
	FieldColor = "color"
	// FieldIcon holds the string denoting the icon field in the database.
	FieldIcon = "icon"
	// FieldURL holds the string denoting the url field in the database.
	FieldURL = "url"
	// FieldKeywords holds the string denoting the keywords field in the database.
	FieldKeywords = "keywords"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldExtras holds the string denoting the extras field in the database.
	FieldExtras = "extras"
	// FieldParentID holds the string denoting the parent_id field in the database.
	FieldParentID = "parent_id"
	// FieldSpaceID holds the string denoting the space_id field in the database.
	FieldSpaceID = "space_id"
	// FieldCreatedBy holds the string denoting the created_by field in the database.
	FieldCreatedBy = "created_by"
	// FieldUpdatedBy holds the string denoting the updated_by field in the database.
	FieldUpdatedBy = "updated_by"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// Table holds the table name of the taxonomy in the database.
	Table = "ncse_cms_taxonomy"
)

// Columns holds all SQL columns for taxonomy fields.
var Columns = []string{
	FieldID,
	FieldName,
	FieldType,
	FieldSlug,
	FieldCover,
	FieldThumbnail,
	FieldColor,
	FieldIcon,
	FieldURL,
	FieldKeywords,
	FieldDescription,
	FieldStatus,
	FieldExtras,
	FieldParentID,
	FieldSpaceID,
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
	// DefaultStatus holds the default value on creation for the "status" field.
	DefaultStatus int
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

// OrderOption defines the ordering options for the Taxonomy queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
}

// BySlug orders the results by the slug field.
func BySlug(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSlug, opts...).ToFunc()
}

// ByCover orders the results by the cover field.
func ByCover(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCover, opts...).ToFunc()
}

// ByThumbnail orders the results by the thumbnail field.
func ByThumbnail(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldThumbnail, opts...).ToFunc()
}

// ByColor orders the results by the color field.
func ByColor(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldColor, opts...).ToFunc()
}

// ByIcon orders the results by the icon field.
func ByIcon(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIcon, opts...).ToFunc()
}

// ByURL orders the results by the url field.
func ByURL(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldURL, opts...).ToFunc()
}

// ByKeywords orders the results by the keywords field.
func ByKeywords(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldKeywords, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByParentID orders the results by the parent_id field.
func ByParentID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldParentID, opts...).ToFunc()
}

// BySpaceID orders the results by the space_id field.
func BySpaceID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSpaceID, opts...).ToFunc()
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
