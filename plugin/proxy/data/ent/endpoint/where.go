// Code generated by ent, DO NOT EDIT.

package endpoint

import (
	"ncobase/proxy/data/ent/predicate"

	"entgo.io/ent/dialect/sql"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldID, id))
}

// Name applies equality check predicate on the "name" field. It's identical to NameEQ.
func Name(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldName, v))
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldDescription, v))
}

// Disabled applies equality check predicate on the "disabled" field. It's identical to DisabledEQ.
func Disabled(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldDisabled, v))
}

// CreatedBy applies equality check predicate on the "created_by" field. It's identical to CreatedByEQ.
func CreatedBy(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldCreatedBy, v))
}

// UpdatedBy applies equality check predicate on the "updated_by" field. It's identical to UpdatedByEQ.
func UpdatedBy(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldUpdatedBy, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldUpdatedAt, v))
}

// BaseURL applies equality check predicate on the "base_url" field. It's identical to BaseURLEQ.
func BaseURL(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldBaseURL, v))
}

// Protocol applies equality check predicate on the "protocol" field. It's identical to ProtocolEQ.
func Protocol(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldProtocol, v))
}

// AuthType applies equality check predicate on the "auth_type" field. It's identical to AuthTypeEQ.
func AuthType(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldAuthType, v))
}

// AuthConfig applies equality check predicate on the "auth_config" field. It's identical to AuthConfigEQ.
func AuthConfig(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldAuthConfig, v))
}

// Timeout applies equality check predicate on the "timeout" field. It's identical to TimeoutEQ.
func Timeout(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldTimeout, v))
}

// UseCircuitBreaker applies equality check predicate on the "use_circuit_breaker" field. It's identical to UseCircuitBreakerEQ.
func UseCircuitBreaker(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldUseCircuitBreaker, v))
}

// RetryCount applies equality check predicate on the "retry_count" field. It's identical to RetryCountEQ.
func RetryCount(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldRetryCount, v))
}

// ValidateSsl applies equality check predicate on the "validate_ssl" field. It's identical to ValidateSslEQ.
func ValidateSsl(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldValidateSsl, v))
}

// LogRequests applies equality check predicate on the "log_requests" field. It's identical to LogRequestsEQ.
func LogRequests(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldLogRequests, v))
}

// LogResponses applies equality check predicate on the "log_responses" field. It's identical to LogResponsesEQ.
func LogResponses(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldLogResponses, v))
}

// NameEQ applies the EQ predicate on the "name" field.
func NameEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldName, v))
}

// NameNEQ applies the NEQ predicate on the "name" field.
func NameNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldName, v))
}

// NameIn applies the In predicate on the "name" field.
func NameIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldName, vs...))
}

// NameNotIn applies the NotIn predicate on the "name" field.
func NameNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldName, vs...))
}

// NameGT applies the GT predicate on the "name" field.
func NameGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldName, v))
}

// NameGTE applies the GTE predicate on the "name" field.
func NameGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldName, v))
}

// NameLT applies the LT predicate on the "name" field.
func NameLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldName, v))
}

// NameLTE applies the LTE predicate on the "name" field.
func NameLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldName, v))
}

// NameContains applies the Contains predicate on the "name" field.
func NameContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldName, v))
}

// NameHasPrefix applies the HasPrefix predicate on the "name" field.
func NameHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldName, v))
}

// NameHasSuffix applies the HasSuffix predicate on the "name" field.
func NameHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldName, v))
}

// NameIsNil applies the IsNil predicate on the "name" field.
func NameIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldName))
}

// NameNotNil applies the NotNil predicate on the "name" field.
func NameNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldName))
}

// NameEqualFold applies the EqualFold predicate on the "name" field.
func NameEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldName, v))
}

// NameContainsFold applies the ContainsFold predicate on the "name" field.
func NameContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldName, v))
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldDescription, v))
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldDescription, v))
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldDescription, vs...))
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldDescription, vs...))
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldDescription, v))
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldDescription, v))
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldDescription, v))
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldDescription, v))
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldDescription, v))
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldDescription, v))
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldDescription, v))
}

// DescriptionIsNil applies the IsNil predicate on the "description" field.
func DescriptionIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldDescription))
}

// DescriptionNotNil applies the NotNil predicate on the "description" field.
func DescriptionNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldDescription))
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldDescription, v))
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldDescription, v))
}

// DisabledEQ applies the EQ predicate on the "disabled" field.
func DisabledEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldDisabled, v))
}

// DisabledNEQ applies the NEQ predicate on the "disabled" field.
func DisabledNEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldDisabled, v))
}

// DisabledIsNil applies the IsNil predicate on the "disabled" field.
func DisabledIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldDisabled))
}

// DisabledNotNil applies the NotNil predicate on the "disabled" field.
func DisabledNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldDisabled))
}

// ExtrasIsNil applies the IsNil predicate on the "extras" field.
func ExtrasIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldExtras))
}

// ExtrasNotNil applies the NotNil predicate on the "extras" field.
func ExtrasNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldExtras))
}

// CreatedByEQ applies the EQ predicate on the "created_by" field.
func CreatedByEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldCreatedBy, v))
}

// CreatedByNEQ applies the NEQ predicate on the "created_by" field.
func CreatedByNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldCreatedBy, v))
}

// CreatedByIn applies the In predicate on the "created_by" field.
func CreatedByIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldCreatedBy, vs...))
}

// CreatedByNotIn applies the NotIn predicate on the "created_by" field.
func CreatedByNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldCreatedBy, vs...))
}

// CreatedByGT applies the GT predicate on the "created_by" field.
func CreatedByGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldCreatedBy, v))
}

// CreatedByGTE applies the GTE predicate on the "created_by" field.
func CreatedByGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldCreatedBy, v))
}

// CreatedByLT applies the LT predicate on the "created_by" field.
func CreatedByLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldCreatedBy, v))
}

// CreatedByLTE applies the LTE predicate on the "created_by" field.
func CreatedByLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldCreatedBy, v))
}

// CreatedByContains applies the Contains predicate on the "created_by" field.
func CreatedByContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldCreatedBy, v))
}

// CreatedByHasPrefix applies the HasPrefix predicate on the "created_by" field.
func CreatedByHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldCreatedBy, v))
}

// CreatedByHasSuffix applies the HasSuffix predicate on the "created_by" field.
func CreatedByHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldCreatedBy, v))
}

// CreatedByIsNil applies the IsNil predicate on the "created_by" field.
func CreatedByIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldCreatedBy))
}

// CreatedByNotNil applies the NotNil predicate on the "created_by" field.
func CreatedByNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldCreatedBy))
}

// CreatedByEqualFold applies the EqualFold predicate on the "created_by" field.
func CreatedByEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldCreatedBy, v))
}

// CreatedByContainsFold applies the ContainsFold predicate on the "created_by" field.
func CreatedByContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldCreatedBy, v))
}

// UpdatedByEQ applies the EQ predicate on the "updated_by" field.
func UpdatedByEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldUpdatedBy, v))
}

// UpdatedByNEQ applies the NEQ predicate on the "updated_by" field.
func UpdatedByNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldUpdatedBy, v))
}

// UpdatedByIn applies the In predicate on the "updated_by" field.
func UpdatedByIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldUpdatedBy, vs...))
}

// UpdatedByNotIn applies the NotIn predicate on the "updated_by" field.
func UpdatedByNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldUpdatedBy, vs...))
}

// UpdatedByGT applies the GT predicate on the "updated_by" field.
func UpdatedByGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldUpdatedBy, v))
}

// UpdatedByGTE applies the GTE predicate on the "updated_by" field.
func UpdatedByGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldUpdatedBy, v))
}

// UpdatedByLT applies the LT predicate on the "updated_by" field.
func UpdatedByLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldUpdatedBy, v))
}

// UpdatedByLTE applies the LTE predicate on the "updated_by" field.
func UpdatedByLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldUpdatedBy, v))
}

// UpdatedByContains applies the Contains predicate on the "updated_by" field.
func UpdatedByContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldUpdatedBy, v))
}

// UpdatedByHasPrefix applies the HasPrefix predicate on the "updated_by" field.
func UpdatedByHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldUpdatedBy, v))
}

// UpdatedByHasSuffix applies the HasSuffix predicate on the "updated_by" field.
func UpdatedByHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldUpdatedBy, v))
}

// UpdatedByIsNil applies the IsNil predicate on the "updated_by" field.
func UpdatedByIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldUpdatedBy))
}

// UpdatedByNotNil applies the NotNil predicate on the "updated_by" field.
func UpdatedByNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldUpdatedBy))
}

// UpdatedByEqualFold applies the EqualFold predicate on the "updated_by" field.
func UpdatedByEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldUpdatedBy, v))
}

// UpdatedByContainsFold applies the ContainsFold predicate on the "updated_by" field.
func UpdatedByContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldUpdatedBy, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldCreatedAt, v))
}

// CreatedAtIsNil applies the IsNil predicate on the "created_at" field.
func CreatedAtIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldCreatedAt))
}

// CreatedAtNotNil applies the NotNil predicate on the "created_at" field.
func CreatedAtNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldCreatedAt))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v int64) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldUpdatedAt, v))
}

// UpdatedAtIsNil applies the IsNil predicate on the "updated_at" field.
func UpdatedAtIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldUpdatedAt))
}

// UpdatedAtNotNil applies the NotNil predicate on the "updated_at" field.
func UpdatedAtNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldUpdatedAt))
}

// BaseURLEQ applies the EQ predicate on the "base_url" field.
func BaseURLEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldBaseURL, v))
}

// BaseURLNEQ applies the NEQ predicate on the "base_url" field.
func BaseURLNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldBaseURL, v))
}

// BaseURLIn applies the In predicate on the "base_url" field.
func BaseURLIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldBaseURL, vs...))
}

// BaseURLNotIn applies the NotIn predicate on the "base_url" field.
func BaseURLNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldBaseURL, vs...))
}

// BaseURLGT applies the GT predicate on the "base_url" field.
func BaseURLGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldBaseURL, v))
}

// BaseURLGTE applies the GTE predicate on the "base_url" field.
func BaseURLGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldBaseURL, v))
}

// BaseURLLT applies the LT predicate on the "base_url" field.
func BaseURLLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldBaseURL, v))
}

// BaseURLLTE applies the LTE predicate on the "base_url" field.
func BaseURLLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldBaseURL, v))
}

// BaseURLContains applies the Contains predicate on the "base_url" field.
func BaseURLContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldBaseURL, v))
}

// BaseURLHasPrefix applies the HasPrefix predicate on the "base_url" field.
func BaseURLHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldBaseURL, v))
}

// BaseURLHasSuffix applies the HasSuffix predicate on the "base_url" field.
func BaseURLHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldBaseURL, v))
}

// BaseURLEqualFold applies the EqualFold predicate on the "base_url" field.
func BaseURLEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldBaseURL, v))
}

// BaseURLContainsFold applies the ContainsFold predicate on the "base_url" field.
func BaseURLContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldBaseURL, v))
}

// ProtocolEQ applies the EQ predicate on the "protocol" field.
func ProtocolEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldProtocol, v))
}

// ProtocolNEQ applies the NEQ predicate on the "protocol" field.
func ProtocolNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldProtocol, v))
}

// ProtocolIn applies the In predicate on the "protocol" field.
func ProtocolIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldProtocol, vs...))
}

// ProtocolNotIn applies the NotIn predicate on the "protocol" field.
func ProtocolNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldProtocol, vs...))
}

// ProtocolGT applies the GT predicate on the "protocol" field.
func ProtocolGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldProtocol, v))
}

// ProtocolGTE applies the GTE predicate on the "protocol" field.
func ProtocolGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldProtocol, v))
}

// ProtocolLT applies the LT predicate on the "protocol" field.
func ProtocolLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldProtocol, v))
}

// ProtocolLTE applies the LTE predicate on the "protocol" field.
func ProtocolLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldProtocol, v))
}

// ProtocolContains applies the Contains predicate on the "protocol" field.
func ProtocolContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldProtocol, v))
}

// ProtocolHasPrefix applies the HasPrefix predicate on the "protocol" field.
func ProtocolHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldProtocol, v))
}

// ProtocolHasSuffix applies the HasSuffix predicate on the "protocol" field.
func ProtocolHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldProtocol, v))
}

// ProtocolEqualFold applies the EqualFold predicate on the "protocol" field.
func ProtocolEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldProtocol, v))
}

// ProtocolContainsFold applies the ContainsFold predicate on the "protocol" field.
func ProtocolContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldProtocol, v))
}

// AuthTypeEQ applies the EQ predicate on the "auth_type" field.
func AuthTypeEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldAuthType, v))
}

// AuthTypeNEQ applies the NEQ predicate on the "auth_type" field.
func AuthTypeNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldAuthType, v))
}

// AuthTypeIn applies the In predicate on the "auth_type" field.
func AuthTypeIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldAuthType, vs...))
}

// AuthTypeNotIn applies the NotIn predicate on the "auth_type" field.
func AuthTypeNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldAuthType, vs...))
}

// AuthTypeGT applies the GT predicate on the "auth_type" field.
func AuthTypeGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldAuthType, v))
}

// AuthTypeGTE applies the GTE predicate on the "auth_type" field.
func AuthTypeGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldAuthType, v))
}

// AuthTypeLT applies the LT predicate on the "auth_type" field.
func AuthTypeLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldAuthType, v))
}

// AuthTypeLTE applies the LTE predicate on the "auth_type" field.
func AuthTypeLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldAuthType, v))
}

// AuthTypeContains applies the Contains predicate on the "auth_type" field.
func AuthTypeContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldAuthType, v))
}

// AuthTypeHasPrefix applies the HasPrefix predicate on the "auth_type" field.
func AuthTypeHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldAuthType, v))
}

// AuthTypeHasSuffix applies the HasSuffix predicate on the "auth_type" field.
func AuthTypeHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldAuthType, v))
}

// AuthTypeEqualFold applies the EqualFold predicate on the "auth_type" field.
func AuthTypeEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldAuthType, v))
}

// AuthTypeContainsFold applies the ContainsFold predicate on the "auth_type" field.
func AuthTypeContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldAuthType, v))
}

// AuthConfigEQ applies the EQ predicate on the "auth_config" field.
func AuthConfigEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldAuthConfig, v))
}

// AuthConfigNEQ applies the NEQ predicate on the "auth_config" field.
func AuthConfigNEQ(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldAuthConfig, v))
}

// AuthConfigIn applies the In predicate on the "auth_config" field.
func AuthConfigIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldAuthConfig, vs...))
}

// AuthConfigNotIn applies the NotIn predicate on the "auth_config" field.
func AuthConfigNotIn(vs ...string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldAuthConfig, vs...))
}

// AuthConfigGT applies the GT predicate on the "auth_config" field.
func AuthConfigGT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldAuthConfig, v))
}

// AuthConfigGTE applies the GTE predicate on the "auth_config" field.
func AuthConfigGTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldAuthConfig, v))
}

// AuthConfigLT applies the LT predicate on the "auth_config" field.
func AuthConfigLT(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldAuthConfig, v))
}

// AuthConfigLTE applies the LTE predicate on the "auth_config" field.
func AuthConfigLTE(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldAuthConfig, v))
}

// AuthConfigContains applies the Contains predicate on the "auth_config" field.
func AuthConfigContains(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContains(FieldAuthConfig, v))
}

// AuthConfigHasPrefix applies the HasPrefix predicate on the "auth_config" field.
func AuthConfigHasPrefix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasPrefix(FieldAuthConfig, v))
}

// AuthConfigHasSuffix applies the HasSuffix predicate on the "auth_config" field.
func AuthConfigHasSuffix(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldHasSuffix(FieldAuthConfig, v))
}

// AuthConfigIsNil applies the IsNil predicate on the "auth_config" field.
func AuthConfigIsNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIsNull(FieldAuthConfig))
}

// AuthConfigNotNil applies the NotNil predicate on the "auth_config" field.
func AuthConfigNotNil() predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotNull(FieldAuthConfig))
}

// AuthConfigEqualFold applies the EqualFold predicate on the "auth_config" field.
func AuthConfigEqualFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEqualFold(FieldAuthConfig, v))
}

// AuthConfigContainsFold applies the ContainsFold predicate on the "auth_config" field.
func AuthConfigContainsFold(v string) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldContainsFold(FieldAuthConfig, v))
}

// TimeoutEQ applies the EQ predicate on the "timeout" field.
func TimeoutEQ(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldTimeout, v))
}

// TimeoutNEQ applies the NEQ predicate on the "timeout" field.
func TimeoutNEQ(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldTimeout, v))
}

// TimeoutIn applies the In predicate on the "timeout" field.
func TimeoutIn(vs ...int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldTimeout, vs...))
}

// TimeoutNotIn applies the NotIn predicate on the "timeout" field.
func TimeoutNotIn(vs ...int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldTimeout, vs...))
}

// TimeoutGT applies the GT predicate on the "timeout" field.
func TimeoutGT(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldTimeout, v))
}

// TimeoutGTE applies the GTE predicate on the "timeout" field.
func TimeoutGTE(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldTimeout, v))
}

// TimeoutLT applies the LT predicate on the "timeout" field.
func TimeoutLT(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldTimeout, v))
}

// TimeoutLTE applies the LTE predicate on the "timeout" field.
func TimeoutLTE(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldTimeout, v))
}

// UseCircuitBreakerEQ applies the EQ predicate on the "use_circuit_breaker" field.
func UseCircuitBreakerEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldUseCircuitBreaker, v))
}

// UseCircuitBreakerNEQ applies the NEQ predicate on the "use_circuit_breaker" field.
func UseCircuitBreakerNEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldUseCircuitBreaker, v))
}

// RetryCountEQ applies the EQ predicate on the "retry_count" field.
func RetryCountEQ(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldRetryCount, v))
}

// RetryCountNEQ applies the NEQ predicate on the "retry_count" field.
func RetryCountNEQ(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldRetryCount, v))
}

// RetryCountIn applies the In predicate on the "retry_count" field.
func RetryCountIn(vs ...int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldIn(FieldRetryCount, vs...))
}

// RetryCountNotIn applies the NotIn predicate on the "retry_count" field.
func RetryCountNotIn(vs ...int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNotIn(FieldRetryCount, vs...))
}

// RetryCountGT applies the GT predicate on the "retry_count" field.
func RetryCountGT(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGT(FieldRetryCount, v))
}

// RetryCountGTE applies the GTE predicate on the "retry_count" field.
func RetryCountGTE(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldGTE(FieldRetryCount, v))
}

// RetryCountLT applies the LT predicate on the "retry_count" field.
func RetryCountLT(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLT(FieldRetryCount, v))
}

// RetryCountLTE applies the LTE predicate on the "retry_count" field.
func RetryCountLTE(v int) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldLTE(FieldRetryCount, v))
}

// ValidateSslEQ applies the EQ predicate on the "validate_ssl" field.
func ValidateSslEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldValidateSsl, v))
}

// ValidateSslNEQ applies the NEQ predicate on the "validate_ssl" field.
func ValidateSslNEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldValidateSsl, v))
}

// LogRequestsEQ applies the EQ predicate on the "log_requests" field.
func LogRequestsEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldLogRequests, v))
}

// LogRequestsNEQ applies the NEQ predicate on the "log_requests" field.
func LogRequestsNEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldLogRequests, v))
}

// LogResponsesEQ applies the EQ predicate on the "log_responses" field.
func LogResponsesEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldEQ(FieldLogResponses, v))
}

// LogResponsesNEQ applies the NEQ predicate on the "log_responses" field.
func LogResponsesNEQ(v bool) predicate.Endpoint {
	return predicate.Endpoint(sql.FieldNEQ(FieldLogResponses, v))
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Endpoint) predicate.Endpoint {
	return predicate.Endpoint(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Endpoint) predicate.Endpoint {
	return predicate.Endpoint(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Endpoint) predicate.Endpoint {
	return predicate.Endpoint(sql.NotPredicates(p))
}
