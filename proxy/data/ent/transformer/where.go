// Code generated by ent, DO NOT EDIT.

package transformer

import (
	"ncobase/proxy/data/ent/predicate"

	"entgo.io/ent/dialect/sql"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldID, id))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldName, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldDescription, v))
}

// Disabled applies equality check predicate on the "disabled" field. It's identical to DisabledEQ.
func Disabled(v bool) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldDisabled, v))
}

// CreatedBy applies equality check predicate on the "created_by" field. It's identical to CreatedByEQ.
func CreatedBy(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldCreatedBy, v))
}

// UpdatedBy applies equality check predicate on the "updated_by" field. It's identical to UpdatedByEQ.
func UpdatedBy(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldUpdatedBy, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldUpdatedAt, v))
}

// Type applies equality check predicate on the "type" field. It's identical to TypeEQ.
func Type(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldType, v))
}

// Content applies equality check predicate on the "content" field. It's identical to ContentEQ.
func Content(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldContent, v))
}

// ContentType applies equality check predicate on the "content_type" field. It's identical to ContentTypeEQ.
func ContentType(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldContentType, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasSuffix(FieldName, v))
}

// NameIsNil applies the IsNil predicate on the "name" field.
func NameIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldName))
}

// NameNotNil applies the NotNil predicate on the "name" field.
func NameNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldName))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldName, v))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldDescription, v))
}

// DisabledEQ applies the EQ predicate on the "disabled" field.
func DisabledEQ(v bool) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldDisabled, v))
}

// DisabledNEQ applies the NEQ predicate on the "disabled" field.
func DisabledNEQ(v bool) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldDisabled, v))
}

// DisabledIsNil applies the IsNil predicate on the "disabled" field.
func DisabledIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldDisabled))
}

// DisabledNotNil applies the NotNil predicate on the "disabled" field.
func DisabledNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldDisabled))
}

// ExtrasIsNil applies the IsNil predicate on the "extras" field.
func ExtrasIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldExtras))
}

// ExtrasNotNil applies the NotNil predicate on the "extras" field.
func ExtrasNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldExtras))
}

// CreatedByEQ applies the EQ predicate on the "created_by" field.
func CreatedByEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldCreatedBy, v))
}

// CreatedByNEQ applies the NEQ predicate on the "created_by" field.
func CreatedByNEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldCreatedBy, v))
}

// CreatedByIn applies the In predicate on the "created_by" field.
func CreatedByIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldCreatedBy, vs...))
}

// CreatedByNotIn applies the NotIn predicate on the "created_by" field.
func CreatedByNotIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldCreatedBy, vs...))
}

// CreatedByGT applies the GT predicate on the "created_by" field.
func CreatedByGT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldCreatedBy, v))
}

// CreatedByGTE applies the GTE predicate on the "created_by" field.
func CreatedByGTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldCreatedBy, v))
}

// CreatedByLT applies the LT predicate on the "created_by" field.
func CreatedByLT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldCreatedBy, v))
}

// CreatedByLTE applies the LTE predicate on the "created_by" field.
func CreatedByLTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldCreatedBy, v))
}

// CreatedByContains applies the Contains predicate on the "created_by" field.
func CreatedByContains(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContains(FieldCreatedBy, v))
}

// CreatedByHasPrefix applies the HasPrefix predicate on the "created_by" field.
func CreatedByHasPrefix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasPrefix(FieldCreatedBy, v))
}

// CreatedByHasSuffix applies the HasSuffix predicate on the "created_by" field.
func CreatedByHasSuffix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasSuffix(FieldCreatedBy, v))
}

// CreatedByIsNil applies the IsNil predicate on the "created_by" field.
func CreatedByIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldCreatedBy))
}

// CreatedByNotNil applies the NotNil predicate on the "created_by" field.
func CreatedByNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldCreatedBy))
}

// CreatedByEqualFold applies the EqualFold predicate on the "created_by" field.
func CreatedByEqualFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldCreatedBy, v))
}

// CreatedByContainsFold applies the ContainsFold predicate on the "created_by" field.
func CreatedByContainsFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldCreatedBy, v))
}

// UpdatedByEQ applies the EQ predicate on the "updated_by" field.
func UpdatedByEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldUpdatedBy, v))
}

// UpdatedByNEQ applies the NEQ predicate on the "updated_by" field.
func UpdatedByNEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldUpdatedBy, v))
}

// UpdatedByIn applies the In predicate on the "updated_by" field.
func UpdatedByIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldUpdatedBy, vs...))
}

// UpdatedByNotIn applies the NotIn predicate on the "updated_by" field.
func UpdatedByNotIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldUpdatedBy, vs...))
}

// UpdatedByGT applies the GT predicate on the "updated_by" field.
func UpdatedByGT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldUpdatedBy, v))
}

// UpdatedByGTE applies the GTE predicate on the "updated_by" field.
func UpdatedByGTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldUpdatedBy, v))
}

// UpdatedByLT applies the LT predicate on the "updated_by" field.
func UpdatedByLT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldUpdatedBy, v))
}

// UpdatedByLTE applies the LTE predicate on the "updated_by" field.
func UpdatedByLTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldUpdatedBy, v))
}

// UpdatedByContains applies the Contains predicate on the "updated_by" field.
func UpdatedByContains(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContains(FieldUpdatedBy, v))
}

// UpdatedByHasPrefix applies the HasPrefix predicate on the "updated_by" field.
func UpdatedByHasPrefix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasPrefix(FieldUpdatedBy, v))
}

// UpdatedByHasSuffix applies the HasSuffix predicate on the "updated_by" field.
func UpdatedByHasSuffix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasSuffix(FieldUpdatedBy, v))
}

// UpdatedByIsNil applies the IsNil predicate on the "updated_by" field.
func UpdatedByIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldUpdatedBy))
}

// UpdatedByNotNil applies the NotNil predicate on the "updated_by" field.
func UpdatedByNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldUpdatedBy))
}

// UpdatedByEqualFold applies the EqualFold predicate on the "updated_by" field.
func UpdatedByEqualFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldUpdatedBy, v))
}

// UpdatedByContainsFold applies the ContainsFold predicate on the "updated_by" field.
func UpdatedByContainsFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldUpdatedBy, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldCreatedAt, v))
}

// CreatedAtIsNil applies the IsNil predicate on the "created_at" field.
func CreatedAtIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldCreatedAt))
}

// CreatedAtNotNil applies the NotNil predicate on the "created_at" field.
func CreatedAtNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldCreatedAt))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v int64) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldUpdatedAt, v))
}

// UpdatedAtIsNil applies the IsNil predicate on the "updated_at" field.
func UpdatedAtIsNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldIsNull(FieldUpdatedAt))
}

// UpdatedAtNotNil applies the NotNil predicate on the "updated_at" field.
func UpdatedAtNotNil() predicate.Transformer {
	return predicate.Transformer(sql.FieldNotNull(FieldUpdatedAt))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldType, v))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldType, v))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldType, vs...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldType, vs...))
}

// TypeGT applies the GT predicate on the "type" field.
func TypeGT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldType, v))
}

// TypeGTE applies the GTE predicate on the "type" field.
func TypeGTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldType, v))
}

// TypeLT applies the LT predicate on the "type" field.
func TypeLT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldType, v))
}

// TypeLTE applies the LTE predicate on the "type" field.
func TypeLTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldType, v))
}

// TypeContains applies the Contains predicate on the "type" field.
func TypeContains(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContains(FieldType, v))
}

// TypeHasPrefix applies the HasPrefix predicate on the "type" field.
func TypeHasPrefix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasPrefix(FieldType, v))
}

// TypeHasSuffix applies the HasSuffix predicate on the "type" field.
func TypeHasSuffix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasSuffix(FieldType, v))
}

// TypeEqualFold applies the EqualFold predicate on the "type" field.
func TypeEqualFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldType, v))
}

// TypeContainsFold applies the ContainsFold predicate on the "type" field.
func TypeContainsFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldType, v))
}

// ContentEQ applies the EQ predicate on the "content" field.
func ContentEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldContent, v))
}

// ContentNEQ applies the NEQ predicate on the "content" field.
func ContentNEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldContent, v))
}

// ContentIn applies the In predicate on the "content" field.
func ContentIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldContent, vs...))
}

// ContentNotIn applies the NotIn predicate on the "content" field.
func ContentNotIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldContent, vs...))
}

// ContentGT applies the GT predicate on the "content" field.
func ContentGT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldContent, v))
}

// ContentGTE applies the GTE predicate on the "content" field.
func ContentGTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldContent, v))
}

// ContentLT applies the LT predicate on the "content" field.
func ContentLT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldContent, v))
}

// ContentLTE applies the LTE predicate on the "content" field.
func ContentLTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldContent, v))
}

// ContentContains applies the Contains predicate on the "content" field.
func ContentContains(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContains(FieldContent, v))
}

// ContentHasPrefix applies the HasPrefix predicate on the "content" field.
func ContentHasPrefix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasPrefix(FieldContent, v))
}

// ContentHasSuffix applies the HasSuffix predicate on the "content" field.
func ContentHasSuffix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasSuffix(FieldContent, v))
}

// ContentEqualFold applies the EqualFold predicate on the "content" field.
func ContentEqualFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldContent, v))
}

// ContentContainsFold applies the ContainsFold predicate on the "content" field.
func ContentContainsFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldContent, v))
}

// ContentTypeEQ applies the EQ predicate on the "content_type" field.
func ContentTypeEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEQ(FieldContentType, v))
}

// ContentTypeNEQ applies the NEQ predicate on the "content_type" field.
func ContentTypeNEQ(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNEQ(FieldContentType, v))
}

// ContentTypeIn applies the In predicate on the "content_type" field.
func ContentTypeIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldIn(FieldContentType, vs...))
}

// ContentTypeNotIn applies the NotIn predicate on the "content_type" field.
func ContentTypeNotIn(vs ...string) predicate.Transformer {
	return predicate.Transformer(sql.FieldNotIn(FieldContentType, vs...))
}

// ContentTypeGT applies the GT predicate on the "content_type" field.
func ContentTypeGT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGT(FieldContentType, v))
}

// ContentTypeGTE applies the GTE predicate on the "content_type" field.
func ContentTypeGTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldGTE(FieldContentType, v))
}

// ContentTypeLT applies the LT predicate on the "content_type" field.
func ContentTypeLT(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLT(FieldContentType, v))
}

// ContentTypeLTE applies the LTE predicate on the "content_type" field.
func ContentTypeLTE(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldLTE(FieldContentType, v))
}

// ContentTypeContains applies the Contains predicate on the "content_type" field.
func ContentTypeContains(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContains(FieldContentType, v))
}

// ContentTypeHasPrefix applies the HasPrefix predicate on the "content_type" field.
func ContentTypeHasPrefix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasPrefix(FieldContentType, v))
}

// ContentTypeHasSuffix applies the HasSuffix predicate on the "content_type" field.
func ContentTypeHasSuffix(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldHasSuffix(FieldContentType, v))
}

// ContentTypeEqualFold applies the EqualFold predicate on the "content_type" field.
func ContentTypeEqualFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldEqualFold(FieldContentType, v))
}

// ContentTypeContainsFold applies the ContainsFold predicate on the "content_type" field.
func ContentTypeContainsFold(v string) predicate.Transformer {
	return predicate.Transformer(sql.FieldContainsFold(FieldContentType, v))
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Transformer) predicate.Transformer {
	return predicate.Transformer(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Transformer) predicate.Transformer {
	return predicate.Transformer(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Transformer) predicate.Transformer {
	return predicate.Transformer(sql.NotPredicates(p))
}
