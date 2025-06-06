// Code generated by ent, DO NOT EDIT.

package group

import (
	"ncobase/space/data/ent/predicate"

	"entgo.io/ent/dialect/sql"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldID, id))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldName, v))
}

// Slug applies equality check predicate on the "slug" field. It's identical to SlugEQ.
func Slug(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldSlug, v))
}

// Type applies equality check predicate on the "type" field. It's identical to TypeEQ.
func Type(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldType, v))
}

// Disabled applies equality check predicate on the "disabled" field. It's identical to DisabledEQ.
func Disabled(v bool) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldDisabled, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldDescription, v))
}

// ParentID applies equality check predicate on the "parent_id" field. It's identical to ParentIDEQ.
func ParentID(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldParentID, v))
}

// CreatedBy applies equality check predicate on the "created_by" field. It's identical to CreatedByEQ.
func CreatedBy(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldCreatedBy, v))
}

// UpdatedBy applies equality check predicate on the "updated_by" field. It's identical to UpdatedByEQ.
func UpdatedBy(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldUpdatedBy, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v int64) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v int64) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldUpdatedAt, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.Group {
	return predicate.Group(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasSuffix(FieldName, v))
}

// NameIsNil applies the IsNil predicate on the "name" field.
func NameIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldName))
}

// NameNotNil applies the NotNil predicate on the "name" field.
func NameNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldName))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldName, v))
}

// SlugEQ applies the EQ predicate on the "slug" field.
func SlugEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldSlug, v))
}

// SlugNEQ applies the NEQ predicate on the "slug" field.
func SlugNEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldSlug, v))
}

// SlugIn applies the In predicate on the "slug" field.
func SlugIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldSlug, vs...))
}

// SlugNotIn applies the NotIn predicate on the "slug" field.
func SlugNotIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldSlug, vs...))
}

// SlugGT applies the GT predicate on the "slug" field.
func SlugGT(v string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldSlug, v))
}

// SlugGTE applies the GTE predicate on the "slug" field.
func SlugGTE(v string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldSlug, v))
}

// SlugLT applies the LT predicate on the "slug" field.
func SlugLT(v string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldSlug, v))
}

// SlugLTE applies the LTE predicate on the "slug" field.
func SlugLTE(v string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldSlug, v))
}

// SlugContains applies the Contains predicate on the "slug" field.
func SlugContains(v string) predicate.Group {
	return predicate.Group(sql.FieldContains(FieldSlug, v))
}

// SlugHasPrefix applies the HasPrefix predicate on the "slug" field.
func SlugHasPrefix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasPrefix(FieldSlug, v))
}

// SlugHasSuffix applies the HasSuffix predicate on the "slug" field.
func SlugHasSuffix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasSuffix(FieldSlug, v))
}

// SlugIsNil applies the IsNil predicate on the "slug" field.
func SlugIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldSlug))
}

// SlugNotNil applies the NotNil predicate on the "slug" field.
func SlugNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldSlug))
}

// SlugEqualFold applies the EqualFold predicate on the "slug" field.
func SlugEqualFold(v string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldSlug, v))
}

// SlugContainsFold applies the ContainsFold predicate on the "slug" field.
func SlugContainsFold(v string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldSlug, v))
}

// TypeEQ applies the EQ predicate on the "type" field.
func TypeEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldType, v))
}

// TypeNEQ applies the NEQ predicate on the "type" field.
func TypeNEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldType, v))
}

// TypeIn applies the In predicate on the "type" field.
func TypeIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldType, vs...))
}

// TypeNotIn applies the NotIn predicate on the "type" field.
func TypeNotIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldType, vs...))
}

// TypeGT applies the GT predicate on the "type" field.
func TypeGT(v string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldType, v))
}

// TypeGTE applies the GTE predicate on the "type" field.
func TypeGTE(v string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldType, v))
}

// TypeLT applies the LT predicate on the "type" field.
func TypeLT(v string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldType, v))
}

// TypeLTE applies the LTE predicate on the "type" field.
func TypeLTE(v string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldType, v))
}

// TypeContains applies the Contains predicate on the "type" field.
func TypeContains(v string) predicate.Group {
	return predicate.Group(sql.FieldContains(FieldType, v))
}

// TypeHasPrefix applies the HasPrefix predicate on the "type" field.
func TypeHasPrefix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasPrefix(FieldType, v))
}

// TypeHasSuffix applies the HasSuffix predicate on the "type" field.
func TypeHasSuffix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasSuffix(FieldType, v))
}

// TypeIsNil applies the IsNil predicate on the "type" field.
func TypeIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldType))
}

// TypeNotNil applies the NotNil predicate on the "type" field.
func TypeNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldType))
}

// TypeEqualFold applies the EqualFold predicate on the "type" field.
func TypeEqualFold(v string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldType, v))
}

// TypeContainsFold applies the ContainsFold predicate on the "type" field.
func TypeContainsFold(v string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldType, v))
}

// DisabledEQ applies the EQ predicate on the "disabled" field.
func DisabledEQ(v bool) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldDisabled, v))
}

// DisabledNEQ applies the NEQ predicate on the "disabled" field.
func DisabledNEQ(v bool) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldDisabled, v))
}

// DisabledIsNil applies the IsNil predicate on the "disabled" field.
func DisabledIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldDisabled))
}

// DisabledNotNil applies the NotNil predicate on the "disabled" field.
func DisabledNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldDisabled))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.Group {
	return predicate.Group(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldDescription, v))
}

// LeaderIsNil applies the IsNil predicate on the "leader" field.
func LeaderIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldLeader))
}

// LeaderNotNil applies the NotNil predicate on the "leader" field.
func LeaderNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldLeader))
}

// ExtrasIsNil applies the IsNil predicate on the "extras" field.
func ExtrasIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldExtras))
}

// ExtrasNotNil applies the NotNil predicate on the "extras" field.
func ExtrasNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldExtras))
}

// ParentIDEQ applies the EQ predicate on the "parent_id" field.
func ParentIDEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldParentID, v))
}

// ParentIDNEQ applies the NEQ predicate on the "parent_id" field.
func ParentIDNEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldParentID, v))
}

// ParentIDIn applies the In predicate on the "parent_id" field.
func ParentIDIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldParentID, vs...))
}

// ParentIDNotIn applies the NotIn predicate on the "parent_id" field.
func ParentIDNotIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldParentID, vs...))
}

// ParentIDGT applies the GT predicate on the "parent_id" field.
func ParentIDGT(v string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldParentID, v))
}

// ParentIDGTE applies the GTE predicate on the "parent_id" field.
func ParentIDGTE(v string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldParentID, v))
}

// ParentIDLT applies the LT predicate on the "parent_id" field.
func ParentIDLT(v string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldParentID, v))
}

// ParentIDLTE applies the LTE predicate on the "parent_id" field.
func ParentIDLTE(v string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldParentID, v))
}

// ParentIDContains applies the Contains predicate on the "parent_id" field.
func ParentIDContains(v string) predicate.Group {
	return predicate.Group(sql.FieldContains(FieldParentID, v))
}

// ParentIDHasPrefix applies the HasPrefix predicate on the "parent_id" field.
func ParentIDHasPrefix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasPrefix(FieldParentID, v))
}

// ParentIDHasSuffix applies the HasSuffix predicate on the "parent_id" field.
func ParentIDHasSuffix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasSuffix(FieldParentID, v))
}

// ParentIDIsNil applies the IsNil predicate on the "parent_id" field.
func ParentIDIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldParentID))
}

// ParentIDNotNil applies the NotNil predicate on the "parent_id" field.
func ParentIDNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldParentID))
}

// ParentIDEqualFold applies the EqualFold predicate on the "parent_id" field.
func ParentIDEqualFold(v string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldParentID, v))
}

// ParentIDContainsFold applies the ContainsFold predicate on the "parent_id" field.
func ParentIDContainsFold(v string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldParentID, v))
}

// CreatedByEQ applies the EQ predicate on the "created_by" field.
func CreatedByEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldCreatedBy, v))
}

// CreatedByNEQ applies the NEQ predicate on the "created_by" field.
func CreatedByNEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldCreatedBy, v))
}

// CreatedByIn applies the In predicate on the "created_by" field.
func CreatedByIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldCreatedBy, vs...))
}

// CreatedByNotIn applies the NotIn predicate on the "created_by" field.
func CreatedByNotIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldCreatedBy, vs...))
}

// CreatedByGT applies the GT predicate on the "created_by" field.
func CreatedByGT(v string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldCreatedBy, v))
}

// CreatedByGTE applies the GTE predicate on the "created_by" field.
func CreatedByGTE(v string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldCreatedBy, v))
}

// CreatedByLT applies the LT predicate on the "created_by" field.
func CreatedByLT(v string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldCreatedBy, v))
}

// CreatedByLTE applies the LTE predicate on the "created_by" field.
func CreatedByLTE(v string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldCreatedBy, v))
}

// CreatedByContains applies the Contains predicate on the "created_by" field.
func CreatedByContains(v string) predicate.Group {
	return predicate.Group(sql.FieldContains(FieldCreatedBy, v))
}

// CreatedByHasPrefix applies the HasPrefix predicate on the "created_by" field.
func CreatedByHasPrefix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasPrefix(FieldCreatedBy, v))
}

// CreatedByHasSuffix applies the HasSuffix predicate on the "created_by" field.
func CreatedByHasSuffix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasSuffix(FieldCreatedBy, v))
}

// CreatedByIsNil applies the IsNil predicate on the "created_by" field.
func CreatedByIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldCreatedBy))
}

// CreatedByNotNil applies the NotNil predicate on the "created_by" field.
func CreatedByNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldCreatedBy))
}

// CreatedByEqualFold applies the EqualFold predicate on the "created_by" field.
func CreatedByEqualFold(v string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldCreatedBy, v))
}

// CreatedByContainsFold applies the ContainsFold predicate on the "created_by" field.
func CreatedByContainsFold(v string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldCreatedBy, v))
}

// UpdatedByEQ applies the EQ predicate on the "updated_by" field.
func UpdatedByEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldUpdatedBy, v))
}

// UpdatedByNEQ applies the NEQ predicate on the "updated_by" field.
func UpdatedByNEQ(v string) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldUpdatedBy, v))
}

// UpdatedByIn applies the In predicate on the "updated_by" field.
func UpdatedByIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldUpdatedBy, vs...))
}

// UpdatedByNotIn applies the NotIn predicate on the "updated_by" field.
func UpdatedByNotIn(vs ...string) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldUpdatedBy, vs...))
}

// UpdatedByGT applies the GT predicate on the "updated_by" field.
func UpdatedByGT(v string) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldUpdatedBy, v))
}

// UpdatedByGTE applies the GTE predicate on the "updated_by" field.
func UpdatedByGTE(v string) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldUpdatedBy, v))
}

// UpdatedByLT applies the LT predicate on the "updated_by" field.
func UpdatedByLT(v string) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldUpdatedBy, v))
}

// UpdatedByLTE applies the LTE predicate on the "updated_by" field.
func UpdatedByLTE(v string) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldUpdatedBy, v))
}

// UpdatedByContains applies the Contains predicate on the "updated_by" field.
func UpdatedByContains(v string) predicate.Group {
	return predicate.Group(sql.FieldContains(FieldUpdatedBy, v))
}

// UpdatedByHasPrefix applies the HasPrefix predicate on the "updated_by" field.
func UpdatedByHasPrefix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasPrefix(FieldUpdatedBy, v))
}

// UpdatedByHasSuffix applies the HasSuffix predicate on the "updated_by" field.
func UpdatedByHasSuffix(v string) predicate.Group {
	return predicate.Group(sql.FieldHasSuffix(FieldUpdatedBy, v))
}

// UpdatedByIsNil applies the IsNil predicate on the "updated_by" field.
func UpdatedByIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldUpdatedBy))
}

// UpdatedByNotNil applies the NotNil predicate on the "updated_by" field.
func UpdatedByNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldUpdatedBy))
}

// UpdatedByEqualFold applies the EqualFold predicate on the "updated_by" field.
func UpdatedByEqualFold(v string) predicate.Group {
	return predicate.Group(sql.FieldEqualFold(FieldUpdatedBy, v))
}

// UpdatedByContainsFold applies the ContainsFold predicate on the "updated_by" field.
func UpdatedByContainsFold(v string) predicate.Group {
	return predicate.Group(sql.FieldContainsFold(FieldUpdatedBy, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v int64) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v int64) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...int64) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...int64) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v int64) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v int64) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v int64) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v int64) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldCreatedAt, v))
}

// CreatedAtIsNil applies the IsNil predicate on the "created_at" field.
func CreatedAtIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldCreatedAt))
}

// CreatedAtNotNil applies the NotNil predicate on the "created_at" field.
func CreatedAtNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldCreatedAt))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v int64) predicate.Group {
	return predicate.Group(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v int64) predicate.Group {
	return predicate.Group(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...int64) predicate.Group {
	return predicate.Group(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...int64) predicate.Group {
	return predicate.Group(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v int64) predicate.Group {
	return predicate.Group(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v int64) predicate.Group {
	return predicate.Group(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v int64) predicate.Group {
	return predicate.Group(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v int64) predicate.Group {
	return predicate.Group(sql.FieldLTE(FieldUpdatedAt, v))
}

// UpdatedAtIsNil applies the IsNil predicate on the "updated_at" field.
func UpdatedAtIsNil() predicate.Group {
	return predicate.Group(sql.FieldIsNull(FieldUpdatedAt))
}

// UpdatedAtNotNil applies the NotNil predicate on the "updated_at" field.
func UpdatedAtNotNil() predicate.Group {
	return predicate.Group(sql.FieldNotNull(FieldUpdatedAt))
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Group) predicate.Group {
	return predicate.Group(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Group) predicate.Group {
	return predicate.Group(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Group) predicate.Group {
	return predicate.Group(sql.NotPredicates(p))
}
