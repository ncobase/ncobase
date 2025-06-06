// Code generated by ent, DO NOT EDIT.

package oauthuser

import (
	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the oauthuser type in the database.
	Label = "oauth_user"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldOauthID holds the string denoting the oauth_id field in the database.
	FieldOauthID = "oauth_id"
	// FieldAccessToken holds the string denoting the access_token field in the database.
	FieldAccessToken = "access_token"
	// FieldProvider holds the string denoting the provider field in the database.
	FieldProvider = "provider"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// Table holds the table name of the oauthuser in the database.
	Table = "ncse_oauth_user"
)

// Columns holds all SQL columns for oauthuser fields.
var Columns = []string{
	FieldID,
	FieldOauthID,
	FieldAccessToken,
	FieldProvider,
	FieldUserID,
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
	// AccessTokenValidator is a validator for the "access_token" field. It is called by the builders before save.
	AccessTokenValidator func(string) error
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

// OrderOption defines the ordering options for the OAuthUser queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByOauthID orders the results by the oauth_id field.
func ByOauthID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOauthID, opts...).ToFunc()
}

// ByAccessToken orders the results by the access_token field.
func ByAccessToken(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAccessToken, opts...).ToFunc()
}

// ByProvider orders the results by the provider field.
func ByProvider(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProvider, opts...).ToFunc()
}

// ByUserID orders the results by the user_id field.
func ByUserID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUserID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}
