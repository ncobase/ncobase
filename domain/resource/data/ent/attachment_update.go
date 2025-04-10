// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"ncobase/domain/resource/data/ent/attachment"
	"ncobase/domain/resource/data/ent/predicate"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// AttachmentUpdate is the builder for updating Attachment entities.
type AttachmentUpdate struct {
	config
	hooks    []Hook
	mutation *AttachmentMutation
}

// Where appends a list predicates to the AttachmentUpdate builder.
func (au *AttachmentUpdate) Where(ps ...predicate.Attachment) *AttachmentUpdate {
	au.mutation.Where(ps...)
	return au
}

// SetName sets the "name" field.
func (au *AttachmentUpdate) SetName(s string) *AttachmentUpdate {
	au.mutation.SetName(s)
	return au
}

// SetNillableName sets the "name" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableName(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetName(*s)
	}
	return au
}

// ClearName clears the value of the "name" field.
func (au *AttachmentUpdate) ClearName() *AttachmentUpdate {
	au.mutation.ClearName()
	return au
}

// SetPath sets the "path" field.
func (au *AttachmentUpdate) SetPath(s string) *AttachmentUpdate {
	au.mutation.SetPath(s)
	return au
}

// SetNillablePath sets the "path" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillablePath(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetPath(*s)
	}
	return au
}

// ClearPath clears the value of the "path" field.
func (au *AttachmentUpdate) ClearPath() *AttachmentUpdate {
	au.mutation.ClearPath()
	return au
}

// SetType sets the "type" field.
func (au *AttachmentUpdate) SetType(s string) *AttachmentUpdate {
	au.mutation.SetType(s)
	return au
}

// SetNillableType sets the "type" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableType(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetType(*s)
	}
	return au
}

// ClearType clears the value of the "type" field.
func (au *AttachmentUpdate) ClearType() *AttachmentUpdate {
	au.mutation.ClearType()
	return au
}

// SetSize sets the "size" field.
func (au *AttachmentUpdate) SetSize(i int) *AttachmentUpdate {
	au.mutation.ResetSize()
	au.mutation.SetSize(i)
	return au
}

// SetNillableSize sets the "size" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableSize(i *int) *AttachmentUpdate {
	if i != nil {
		au.SetSize(*i)
	}
	return au
}

// AddSize adds i to the "size" field.
func (au *AttachmentUpdate) AddSize(i int) *AttachmentUpdate {
	au.mutation.AddSize(i)
	return au
}

// SetStorage sets the "storage" field.
func (au *AttachmentUpdate) SetStorage(s string) *AttachmentUpdate {
	au.mutation.SetStorage(s)
	return au
}

// SetNillableStorage sets the "storage" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableStorage(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetStorage(*s)
	}
	return au
}

// ClearStorage clears the value of the "storage" field.
func (au *AttachmentUpdate) ClearStorage() *AttachmentUpdate {
	au.mutation.ClearStorage()
	return au
}

// SetBucket sets the "bucket" field.
func (au *AttachmentUpdate) SetBucket(s string) *AttachmentUpdate {
	au.mutation.SetBucket(s)
	return au
}

// SetNillableBucket sets the "bucket" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableBucket(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetBucket(*s)
	}
	return au
}

// ClearBucket clears the value of the "bucket" field.
func (au *AttachmentUpdate) ClearBucket() *AttachmentUpdate {
	au.mutation.ClearBucket()
	return au
}

// SetEndpoint sets the "endpoint" field.
func (au *AttachmentUpdate) SetEndpoint(s string) *AttachmentUpdate {
	au.mutation.SetEndpoint(s)
	return au
}

// SetNillableEndpoint sets the "endpoint" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableEndpoint(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetEndpoint(*s)
	}
	return au
}

// ClearEndpoint clears the value of the "endpoint" field.
func (au *AttachmentUpdate) ClearEndpoint() *AttachmentUpdate {
	au.mutation.ClearEndpoint()
	return au
}

// SetObjectID sets the "object_id" field.
func (au *AttachmentUpdate) SetObjectID(s string) *AttachmentUpdate {
	au.mutation.SetObjectID(s)
	return au
}

// SetNillableObjectID sets the "object_id" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableObjectID(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetObjectID(*s)
	}
	return au
}

// ClearObjectID clears the value of the "object_id" field.
func (au *AttachmentUpdate) ClearObjectID() *AttachmentUpdate {
	au.mutation.ClearObjectID()
	return au
}

// SetTenantID sets the "tenant_id" field.
func (au *AttachmentUpdate) SetTenantID(s string) *AttachmentUpdate {
	au.mutation.SetTenantID(s)
	return au
}

// SetNillableTenantID sets the "tenant_id" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableTenantID(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetTenantID(*s)
	}
	return au
}

// ClearTenantID clears the value of the "tenant_id" field.
func (au *AttachmentUpdate) ClearTenantID() *AttachmentUpdate {
	au.mutation.ClearTenantID()
	return au
}

// SetExtras sets the "extras" field.
func (au *AttachmentUpdate) SetExtras(m map[string]interface{}) *AttachmentUpdate {
	au.mutation.SetExtras(m)
	return au
}

// ClearExtras clears the value of the "extras" field.
func (au *AttachmentUpdate) ClearExtras() *AttachmentUpdate {
	au.mutation.ClearExtras()
	return au
}

// SetCreatedBy sets the "created_by" field.
func (au *AttachmentUpdate) SetCreatedBy(s string) *AttachmentUpdate {
	au.mutation.SetCreatedBy(s)
	return au
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableCreatedBy(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetCreatedBy(*s)
	}
	return au
}

// ClearCreatedBy clears the value of the "created_by" field.
func (au *AttachmentUpdate) ClearCreatedBy() *AttachmentUpdate {
	au.mutation.ClearCreatedBy()
	return au
}

// SetUpdatedBy sets the "updated_by" field.
func (au *AttachmentUpdate) SetUpdatedBy(s string) *AttachmentUpdate {
	au.mutation.SetUpdatedBy(s)
	return au
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (au *AttachmentUpdate) SetNillableUpdatedBy(s *string) *AttachmentUpdate {
	if s != nil {
		au.SetUpdatedBy(*s)
	}
	return au
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (au *AttachmentUpdate) ClearUpdatedBy() *AttachmentUpdate {
	au.mutation.ClearUpdatedBy()
	return au
}

// SetUpdatedAt sets the "updated_at" field.
func (au *AttachmentUpdate) SetUpdatedAt(i int64) *AttachmentUpdate {
	au.mutation.ResetUpdatedAt()
	au.mutation.SetUpdatedAt(i)
	return au
}

// AddUpdatedAt adds i to the "updated_at" field.
func (au *AttachmentUpdate) AddUpdatedAt(i int64) *AttachmentUpdate {
	au.mutation.AddUpdatedAt(i)
	return au
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (au *AttachmentUpdate) ClearUpdatedAt() *AttachmentUpdate {
	au.mutation.ClearUpdatedAt()
	return au
}

// Mutation returns the AttachmentMutation object of the builder.
func (au *AttachmentUpdate) Mutation() *AttachmentMutation {
	return au.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (au *AttachmentUpdate) Save(ctx context.Context) (int, error) {
	au.defaults()
	return withHooks(ctx, au.sqlSave, au.mutation, au.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (au *AttachmentUpdate) SaveX(ctx context.Context) int {
	affected, err := au.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (au *AttachmentUpdate) Exec(ctx context.Context) error {
	_, err := au.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (au *AttachmentUpdate) ExecX(ctx context.Context) {
	if err := au.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (au *AttachmentUpdate) defaults() {
	if _, ok := au.mutation.UpdatedAt(); !ok && !au.mutation.UpdatedAtCleared() {
		v := attachment.UpdateDefaultUpdatedAt()
		au.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (au *AttachmentUpdate) check() error {
	if v, ok := au.mutation.Name(); ok {
		if err := attachment.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "Attachment.name": %w`, err)}
		}
	}
	if v, ok := au.mutation.ObjectID(); ok {
		if err := attachment.ObjectIDValidator(v); err != nil {
			return &ValidationError{Name: "object_id", err: fmt.Errorf(`ent: validator failed for field "Attachment.object_id": %w`, err)}
		}
	}
	if v, ok := au.mutation.TenantID(); ok {
		if err := attachment.TenantIDValidator(v); err != nil {
			return &ValidationError{Name: "tenant_id", err: fmt.Errorf(`ent: validator failed for field "Attachment.tenant_id": %w`, err)}
		}
	}
	if v, ok := au.mutation.CreatedBy(); ok {
		if err := attachment.CreatedByValidator(v); err != nil {
			return &ValidationError{Name: "created_by", err: fmt.Errorf(`ent: validator failed for field "Attachment.created_by": %w`, err)}
		}
	}
	if v, ok := au.mutation.UpdatedBy(); ok {
		if err := attachment.UpdatedByValidator(v); err != nil {
			return &ValidationError{Name: "updated_by", err: fmt.Errorf(`ent: validator failed for field "Attachment.updated_by": %w`, err)}
		}
	}
	return nil
}

func (au *AttachmentUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := au.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(attachment.Table, attachment.Columns, sqlgraph.NewFieldSpec(attachment.FieldID, field.TypeString))
	if ps := au.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := au.mutation.Name(); ok {
		_spec.SetField(attachment.FieldName, field.TypeString, value)
	}
	if au.mutation.NameCleared() {
		_spec.ClearField(attachment.FieldName, field.TypeString)
	}
	if value, ok := au.mutation.Path(); ok {
		_spec.SetField(attachment.FieldPath, field.TypeString, value)
	}
	if au.mutation.PathCleared() {
		_spec.ClearField(attachment.FieldPath, field.TypeString)
	}
	if value, ok := au.mutation.GetType(); ok {
		_spec.SetField(attachment.FieldType, field.TypeString, value)
	}
	if au.mutation.TypeCleared() {
		_spec.ClearField(attachment.FieldType, field.TypeString)
	}
	if value, ok := au.mutation.Size(); ok {
		_spec.SetField(attachment.FieldSize, field.TypeInt, value)
	}
	if value, ok := au.mutation.AddedSize(); ok {
		_spec.AddField(attachment.FieldSize, field.TypeInt, value)
	}
	if value, ok := au.mutation.Storage(); ok {
		_spec.SetField(attachment.FieldStorage, field.TypeString, value)
	}
	if au.mutation.StorageCleared() {
		_spec.ClearField(attachment.FieldStorage, field.TypeString)
	}
	if value, ok := au.mutation.Bucket(); ok {
		_spec.SetField(attachment.FieldBucket, field.TypeString, value)
	}
	if au.mutation.BucketCleared() {
		_spec.ClearField(attachment.FieldBucket, field.TypeString)
	}
	if value, ok := au.mutation.Endpoint(); ok {
		_spec.SetField(attachment.FieldEndpoint, field.TypeString, value)
	}
	if au.mutation.EndpointCleared() {
		_spec.ClearField(attachment.FieldEndpoint, field.TypeString)
	}
	if value, ok := au.mutation.ObjectID(); ok {
		_spec.SetField(attachment.FieldObjectID, field.TypeString, value)
	}
	if au.mutation.ObjectIDCleared() {
		_spec.ClearField(attachment.FieldObjectID, field.TypeString)
	}
	if value, ok := au.mutation.TenantID(); ok {
		_spec.SetField(attachment.FieldTenantID, field.TypeString, value)
	}
	if au.mutation.TenantIDCleared() {
		_spec.ClearField(attachment.FieldTenantID, field.TypeString)
	}
	if value, ok := au.mutation.Extras(); ok {
		_spec.SetField(attachment.FieldExtras, field.TypeJSON, value)
	}
	if au.mutation.ExtrasCleared() {
		_spec.ClearField(attachment.FieldExtras, field.TypeJSON)
	}
	if value, ok := au.mutation.CreatedBy(); ok {
		_spec.SetField(attachment.FieldCreatedBy, field.TypeString, value)
	}
	if au.mutation.CreatedByCleared() {
		_spec.ClearField(attachment.FieldCreatedBy, field.TypeString)
	}
	if value, ok := au.mutation.UpdatedBy(); ok {
		_spec.SetField(attachment.FieldUpdatedBy, field.TypeString, value)
	}
	if au.mutation.UpdatedByCleared() {
		_spec.ClearField(attachment.FieldUpdatedBy, field.TypeString)
	}
	if au.mutation.CreatedAtCleared() {
		_spec.ClearField(attachment.FieldCreatedAt, field.TypeInt64)
	}
	if value, ok := au.mutation.UpdatedAt(); ok {
		_spec.SetField(attachment.FieldUpdatedAt, field.TypeInt64, value)
	}
	if value, ok := au.mutation.AddedUpdatedAt(); ok {
		_spec.AddField(attachment.FieldUpdatedAt, field.TypeInt64, value)
	}
	if au.mutation.UpdatedAtCleared() {
		_spec.ClearField(attachment.FieldUpdatedAt, field.TypeInt64)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, au.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{attachment.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	au.mutation.done = true
	return n, nil
}

// AttachmentUpdateOne is the builder for updating a single Attachment entity.
type AttachmentUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *AttachmentMutation
}

// SetName sets the "name" field.
func (auo *AttachmentUpdateOne) SetName(s string) *AttachmentUpdateOne {
	auo.mutation.SetName(s)
	return auo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableName(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetName(*s)
	}
	return auo
}

// ClearName clears the value of the "name" field.
func (auo *AttachmentUpdateOne) ClearName() *AttachmentUpdateOne {
	auo.mutation.ClearName()
	return auo
}

// SetPath sets the "path" field.
func (auo *AttachmentUpdateOne) SetPath(s string) *AttachmentUpdateOne {
	auo.mutation.SetPath(s)
	return auo
}

// SetNillablePath sets the "path" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillablePath(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetPath(*s)
	}
	return auo
}

// ClearPath clears the value of the "path" field.
func (auo *AttachmentUpdateOne) ClearPath() *AttachmentUpdateOne {
	auo.mutation.ClearPath()
	return auo
}

// SetType sets the "type" field.
func (auo *AttachmentUpdateOne) SetType(s string) *AttachmentUpdateOne {
	auo.mutation.SetType(s)
	return auo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableType(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetType(*s)
	}
	return auo
}

// ClearType clears the value of the "type" field.
func (auo *AttachmentUpdateOne) ClearType() *AttachmentUpdateOne {
	auo.mutation.ClearType()
	return auo
}

// SetSize sets the "size" field.
func (auo *AttachmentUpdateOne) SetSize(i int) *AttachmentUpdateOne {
	auo.mutation.ResetSize()
	auo.mutation.SetSize(i)
	return auo
}

// SetNillableSize sets the "size" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableSize(i *int) *AttachmentUpdateOne {
	if i != nil {
		auo.SetSize(*i)
	}
	return auo
}

// AddSize adds i to the "size" field.
func (auo *AttachmentUpdateOne) AddSize(i int) *AttachmentUpdateOne {
	auo.mutation.AddSize(i)
	return auo
}

// SetStorage sets the "storage" field.
func (auo *AttachmentUpdateOne) SetStorage(s string) *AttachmentUpdateOne {
	auo.mutation.SetStorage(s)
	return auo
}

// SetNillableStorage sets the "storage" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableStorage(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetStorage(*s)
	}
	return auo
}

// ClearStorage clears the value of the "storage" field.
func (auo *AttachmentUpdateOne) ClearStorage() *AttachmentUpdateOne {
	auo.mutation.ClearStorage()
	return auo
}

// SetBucket sets the "bucket" field.
func (auo *AttachmentUpdateOne) SetBucket(s string) *AttachmentUpdateOne {
	auo.mutation.SetBucket(s)
	return auo
}

// SetNillableBucket sets the "bucket" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableBucket(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetBucket(*s)
	}
	return auo
}

// ClearBucket clears the value of the "bucket" field.
func (auo *AttachmentUpdateOne) ClearBucket() *AttachmentUpdateOne {
	auo.mutation.ClearBucket()
	return auo
}

// SetEndpoint sets the "endpoint" field.
func (auo *AttachmentUpdateOne) SetEndpoint(s string) *AttachmentUpdateOne {
	auo.mutation.SetEndpoint(s)
	return auo
}

// SetNillableEndpoint sets the "endpoint" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableEndpoint(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetEndpoint(*s)
	}
	return auo
}

// ClearEndpoint clears the value of the "endpoint" field.
func (auo *AttachmentUpdateOne) ClearEndpoint() *AttachmentUpdateOne {
	auo.mutation.ClearEndpoint()
	return auo
}

// SetObjectID sets the "object_id" field.
func (auo *AttachmentUpdateOne) SetObjectID(s string) *AttachmentUpdateOne {
	auo.mutation.SetObjectID(s)
	return auo
}

// SetNillableObjectID sets the "object_id" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableObjectID(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetObjectID(*s)
	}
	return auo
}

// ClearObjectID clears the value of the "object_id" field.
func (auo *AttachmentUpdateOne) ClearObjectID() *AttachmentUpdateOne {
	auo.mutation.ClearObjectID()
	return auo
}

// SetTenantID sets the "tenant_id" field.
func (auo *AttachmentUpdateOne) SetTenantID(s string) *AttachmentUpdateOne {
	auo.mutation.SetTenantID(s)
	return auo
}

// SetNillableTenantID sets the "tenant_id" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableTenantID(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetTenantID(*s)
	}
	return auo
}

// ClearTenantID clears the value of the "tenant_id" field.
func (auo *AttachmentUpdateOne) ClearTenantID() *AttachmentUpdateOne {
	auo.mutation.ClearTenantID()
	return auo
}

// SetExtras sets the "extras" field.
func (auo *AttachmentUpdateOne) SetExtras(m map[string]interface{}) *AttachmentUpdateOne {
	auo.mutation.SetExtras(m)
	return auo
}

// ClearExtras clears the value of the "extras" field.
func (auo *AttachmentUpdateOne) ClearExtras() *AttachmentUpdateOne {
	auo.mutation.ClearExtras()
	return auo
}

// SetCreatedBy sets the "created_by" field.
func (auo *AttachmentUpdateOne) SetCreatedBy(s string) *AttachmentUpdateOne {
	auo.mutation.SetCreatedBy(s)
	return auo
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableCreatedBy(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetCreatedBy(*s)
	}
	return auo
}

// ClearCreatedBy clears the value of the "created_by" field.
func (auo *AttachmentUpdateOne) ClearCreatedBy() *AttachmentUpdateOne {
	auo.mutation.ClearCreatedBy()
	return auo
}

// SetUpdatedBy sets the "updated_by" field.
func (auo *AttachmentUpdateOne) SetUpdatedBy(s string) *AttachmentUpdateOne {
	auo.mutation.SetUpdatedBy(s)
	return auo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (auo *AttachmentUpdateOne) SetNillableUpdatedBy(s *string) *AttachmentUpdateOne {
	if s != nil {
		auo.SetUpdatedBy(*s)
	}
	return auo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (auo *AttachmentUpdateOne) ClearUpdatedBy() *AttachmentUpdateOne {
	auo.mutation.ClearUpdatedBy()
	return auo
}

// SetUpdatedAt sets the "updated_at" field.
func (auo *AttachmentUpdateOne) SetUpdatedAt(i int64) *AttachmentUpdateOne {
	auo.mutation.ResetUpdatedAt()
	auo.mutation.SetUpdatedAt(i)
	return auo
}

// AddUpdatedAt adds i to the "updated_at" field.
func (auo *AttachmentUpdateOne) AddUpdatedAt(i int64) *AttachmentUpdateOne {
	auo.mutation.AddUpdatedAt(i)
	return auo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (auo *AttachmentUpdateOne) ClearUpdatedAt() *AttachmentUpdateOne {
	auo.mutation.ClearUpdatedAt()
	return auo
}

// Mutation returns the AttachmentMutation object of the builder.
func (auo *AttachmentUpdateOne) Mutation() *AttachmentMutation {
	return auo.mutation
}

// Where appends a list predicates to the AttachmentUpdate builder.
func (auo *AttachmentUpdateOne) Where(ps ...predicate.Attachment) *AttachmentUpdateOne {
	auo.mutation.Where(ps...)
	return auo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (auo *AttachmentUpdateOne) Select(field string, fields ...string) *AttachmentUpdateOne {
	auo.fields = append([]string{field}, fields...)
	return auo
}

// Save executes the query and returns the updated Attachment entity.
func (auo *AttachmentUpdateOne) Save(ctx context.Context) (*Attachment, error) {
	auo.defaults()
	return withHooks(ctx, auo.sqlSave, auo.mutation, auo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (auo *AttachmentUpdateOne) SaveX(ctx context.Context) *Attachment {
	node, err := auo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (auo *AttachmentUpdateOne) Exec(ctx context.Context) error {
	_, err := auo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (auo *AttachmentUpdateOne) ExecX(ctx context.Context) {
	if err := auo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (auo *AttachmentUpdateOne) defaults() {
	if _, ok := auo.mutation.UpdatedAt(); !ok && !auo.mutation.UpdatedAtCleared() {
		v := attachment.UpdateDefaultUpdatedAt()
		auo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (auo *AttachmentUpdateOne) check() error {
	if v, ok := auo.mutation.Name(); ok {
		if err := attachment.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "Attachment.name": %w`, err)}
		}
	}
	if v, ok := auo.mutation.ObjectID(); ok {
		if err := attachment.ObjectIDValidator(v); err != nil {
			return &ValidationError{Name: "object_id", err: fmt.Errorf(`ent: validator failed for field "Attachment.object_id": %w`, err)}
		}
	}
	if v, ok := auo.mutation.TenantID(); ok {
		if err := attachment.TenantIDValidator(v); err != nil {
			return &ValidationError{Name: "tenant_id", err: fmt.Errorf(`ent: validator failed for field "Attachment.tenant_id": %w`, err)}
		}
	}
	if v, ok := auo.mutation.CreatedBy(); ok {
		if err := attachment.CreatedByValidator(v); err != nil {
			return &ValidationError{Name: "created_by", err: fmt.Errorf(`ent: validator failed for field "Attachment.created_by": %w`, err)}
		}
	}
	if v, ok := auo.mutation.UpdatedBy(); ok {
		if err := attachment.UpdatedByValidator(v); err != nil {
			return &ValidationError{Name: "updated_by", err: fmt.Errorf(`ent: validator failed for field "Attachment.updated_by": %w`, err)}
		}
	}
	return nil
}

func (auo *AttachmentUpdateOne) sqlSave(ctx context.Context) (_node *Attachment, err error) {
	if err := auo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(attachment.Table, attachment.Columns, sqlgraph.NewFieldSpec(attachment.FieldID, field.TypeString))
	id, ok := auo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Attachment.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := auo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, attachment.FieldID)
		for _, f := range fields {
			if !attachment.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != attachment.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := auo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := auo.mutation.Name(); ok {
		_spec.SetField(attachment.FieldName, field.TypeString, value)
	}
	if auo.mutation.NameCleared() {
		_spec.ClearField(attachment.FieldName, field.TypeString)
	}
	if value, ok := auo.mutation.Path(); ok {
		_spec.SetField(attachment.FieldPath, field.TypeString, value)
	}
	if auo.mutation.PathCleared() {
		_spec.ClearField(attachment.FieldPath, field.TypeString)
	}
	if value, ok := auo.mutation.GetType(); ok {
		_spec.SetField(attachment.FieldType, field.TypeString, value)
	}
	if auo.mutation.TypeCleared() {
		_spec.ClearField(attachment.FieldType, field.TypeString)
	}
	if value, ok := auo.mutation.Size(); ok {
		_spec.SetField(attachment.FieldSize, field.TypeInt, value)
	}
	if value, ok := auo.mutation.AddedSize(); ok {
		_spec.AddField(attachment.FieldSize, field.TypeInt, value)
	}
	if value, ok := auo.mutation.Storage(); ok {
		_spec.SetField(attachment.FieldStorage, field.TypeString, value)
	}
	if auo.mutation.StorageCleared() {
		_spec.ClearField(attachment.FieldStorage, field.TypeString)
	}
	if value, ok := auo.mutation.Bucket(); ok {
		_spec.SetField(attachment.FieldBucket, field.TypeString, value)
	}
	if auo.mutation.BucketCleared() {
		_spec.ClearField(attachment.FieldBucket, field.TypeString)
	}
	if value, ok := auo.mutation.Endpoint(); ok {
		_spec.SetField(attachment.FieldEndpoint, field.TypeString, value)
	}
	if auo.mutation.EndpointCleared() {
		_spec.ClearField(attachment.FieldEndpoint, field.TypeString)
	}
	if value, ok := auo.mutation.ObjectID(); ok {
		_spec.SetField(attachment.FieldObjectID, field.TypeString, value)
	}
	if auo.mutation.ObjectIDCleared() {
		_spec.ClearField(attachment.FieldObjectID, field.TypeString)
	}
	if value, ok := auo.mutation.TenantID(); ok {
		_spec.SetField(attachment.FieldTenantID, field.TypeString, value)
	}
	if auo.mutation.TenantIDCleared() {
		_spec.ClearField(attachment.FieldTenantID, field.TypeString)
	}
	if value, ok := auo.mutation.Extras(); ok {
		_spec.SetField(attachment.FieldExtras, field.TypeJSON, value)
	}
	if auo.mutation.ExtrasCleared() {
		_spec.ClearField(attachment.FieldExtras, field.TypeJSON)
	}
	if value, ok := auo.mutation.CreatedBy(); ok {
		_spec.SetField(attachment.FieldCreatedBy, field.TypeString, value)
	}
	if auo.mutation.CreatedByCleared() {
		_spec.ClearField(attachment.FieldCreatedBy, field.TypeString)
	}
	if value, ok := auo.mutation.UpdatedBy(); ok {
		_spec.SetField(attachment.FieldUpdatedBy, field.TypeString, value)
	}
	if auo.mutation.UpdatedByCleared() {
		_spec.ClearField(attachment.FieldUpdatedBy, field.TypeString)
	}
	if auo.mutation.CreatedAtCleared() {
		_spec.ClearField(attachment.FieldCreatedAt, field.TypeInt64)
	}
	if value, ok := auo.mutation.UpdatedAt(); ok {
		_spec.SetField(attachment.FieldUpdatedAt, field.TypeInt64, value)
	}
	if value, ok := auo.mutation.AddedUpdatedAt(); ok {
		_spec.AddField(attachment.FieldUpdatedAt, field.TypeInt64, value)
	}
	if auo.mutation.UpdatedAtCleared() {
		_spec.ClearField(attachment.FieldUpdatedAt, field.TypeInt64)
	}
	_node = &Attachment{config: auo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, auo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{attachment.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	auo.mutation.done = true
	return _node, nil
}
