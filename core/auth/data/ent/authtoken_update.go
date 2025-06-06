// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"ncobase/auth/data/ent/authtoken"
	"ncobase/auth/data/ent/predicate"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// AuthTokenUpdate is the builder for updating AuthToken entities.
type AuthTokenUpdate struct {
	config
	hooks    []Hook
	mutation *AuthTokenMutation
}

// Where appends a list predicates to the AuthTokenUpdate builder.
func (atu *AuthTokenUpdate) Where(ps ...predicate.AuthToken) *AuthTokenUpdate {
	atu.mutation.Where(ps...)
	return atu
}

// SetDisabled sets the "disabled" field.
func (atu *AuthTokenUpdate) SetDisabled(b bool) *AuthTokenUpdate {
	atu.mutation.SetDisabled(b)
	return atu
}

// SetNillableDisabled sets the "disabled" field if the given value is not nil.
func (atu *AuthTokenUpdate) SetNillableDisabled(b *bool) *AuthTokenUpdate {
	if b != nil {
		atu.SetDisabled(*b)
	}
	return atu
}

// ClearDisabled clears the value of the "disabled" field.
func (atu *AuthTokenUpdate) ClearDisabled() *AuthTokenUpdate {
	atu.mutation.ClearDisabled()
	return atu
}

// SetUpdatedAt sets the "updated_at" field.
func (atu *AuthTokenUpdate) SetUpdatedAt(i int64) *AuthTokenUpdate {
	atu.mutation.ResetUpdatedAt()
	atu.mutation.SetUpdatedAt(i)
	return atu
}

// AddUpdatedAt adds i to the "updated_at" field.
func (atu *AuthTokenUpdate) AddUpdatedAt(i int64) *AuthTokenUpdate {
	atu.mutation.AddUpdatedAt(i)
	return atu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (atu *AuthTokenUpdate) ClearUpdatedAt() *AuthTokenUpdate {
	atu.mutation.ClearUpdatedAt()
	return atu
}

// SetUserID sets the "user_id" field.
func (atu *AuthTokenUpdate) SetUserID(s string) *AuthTokenUpdate {
	atu.mutation.SetUserID(s)
	return atu
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (atu *AuthTokenUpdate) SetNillableUserID(s *string) *AuthTokenUpdate {
	if s != nil {
		atu.SetUserID(*s)
	}
	return atu
}

// ClearUserID clears the value of the "user_id" field.
func (atu *AuthTokenUpdate) ClearUserID() *AuthTokenUpdate {
	atu.mutation.ClearUserID()
	return atu
}

// Mutation returns the AuthTokenMutation object of the builder.
func (atu *AuthTokenUpdate) Mutation() *AuthTokenMutation {
	return atu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (atu *AuthTokenUpdate) Save(ctx context.Context) (int, error) {
	atu.defaults()
	return withHooks(ctx, atu.sqlSave, atu.mutation, atu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (atu *AuthTokenUpdate) SaveX(ctx context.Context) int {
	affected, err := atu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (atu *AuthTokenUpdate) Exec(ctx context.Context) error {
	_, err := atu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atu *AuthTokenUpdate) ExecX(ctx context.Context) {
	if err := atu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (atu *AuthTokenUpdate) defaults() {
	if _, ok := atu.mutation.UpdatedAt(); !ok && !atu.mutation.UpdatedAtCleared() {
		v := authtoken.UpdateDefaultUpdatedAt()
		atu.mutation.SetUpdatedAt(v)
	}
}

func (atu *AuthTokenUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(authtoken.Table, authtoken.Columns, sqlgraph.NewFieldSpec(authtoken.FieldID, field.TypeString))
	if ps := atu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := atu.mutation.Disabled(); ok {
		_spec.SetField(authtoken.FieldDisabled, field.TypeBool, value)
	}
	if atu.mutation.DisabledCleared() {
		_spec.ClearField(authtoken.FieldDisabled, field.TypeBool)
	}
	if atu.mutation.CreatedAtCleared() {
		_spec.ClearField(authtoken.FieldCreatedAt, field.TypeInt64)
	}
	if value, ok := atu.mutation.UpdatedAt(); ok {
		_spec.SetField(authtoken.FieldUpdatedAt, field.TypeInt64, value)
	}
	if value, ok := atu.mutation.AddedUpdatedAt(); ok {
		_spec.AddField(authtoken.FieldUpdatedAt, field.TypeInt64, value)
	}
	if atu.mutation.UpdatedAtCleared() {
		_spec.ClearField(authtoken.FieldUpdatedAt, field.TypeInt64)
	}
	if value, ok := atu.mutation.UserID(); ok {
		_spec.SetField(authtoken.FieldUserID, field.TypeString, value)
	}
	if atu.mutation.UserIDCleared() {
		_spec.ClearField(authtoken.FieldUserID, field.TypeString)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, atu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{authtoken.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	atu.mutation.done = true
	return n, nil
}

// AuthTokenUpdateOne is the builder for updating a single AuthToken entity.
type AuthTokenUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *AuthTokenMutation
}

// SetDisabled sets the "disabled" field.
func (atuo *AuthTokenUpdateOne) SetDisabled(b bool) *AuthTokenUpdateOne {
	atuo.mutation.SetDisabled(b)
	return atuo
}

// SetNillableDisabled sets the "disabled" field if the given value is not nil.
func (atuo *AuthTokenUpdateOne) SetNillableDisabled(b *bool) *AuthTokenUpdateOne {
	if b != nil {
		atuo.SetDisabled(*b)
	}
	return atuo
}

// ClearDisabled clears the value of the "disabled" field.
func (atuo *AuthTokenUpdateOne) ClearDisabled() *AuthTokenUpdateOne {
	atuo.mutation.ClearDisabled()
	return atuo
}

// SetUpdatedAt sets the "updated_at" field.
func (atuo *AuthTokenUpdateOne) SetUpdatedAt(i int64) *AuthTokenUpdateOne {
	atuo.mutation.ResetUpdatedAt()
	atuo.mutation.SetUpdatedAt(i)
	return atuo
}

// AddUpdatedAt adds i to the "updated_at" field.
func (atuo *AuthTokenUpdateOne) AddUpdatedAt(i int64) *AuthTokenUpdateOne {
	atuo.mutation.AddUpdatedAt(i)
	return atuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (atuo *AuthTokenUpdateOne) ClearUpdatedAt() *AuthTokenUpdateOne {
	atuo.mutation.ClearUpdatedAt()
	return atuo
}

// SetUserID sets the "user_id" field.
func (atuo *AuthTokenUpdateOne) SetUserID(s string) *AuthTokenUpdateOne {
	atuo.mutation.SetUserID(s)
	return atuo
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (atuo *AuthTokenUpdateOne) SetNillableUserID(s *string) *AuthTokenUpdateOne {
	if s != nil {
		atuo.SetUserID(*s)
	}
	return atuo
}

// ClearUserID clears the value of the "user_id" field.
func (atuo *AuthTokenUpdateOne) ClearUserID() *AuthTokenUpdateOne {
	atuo.mutation.ClearUserID()
	return atuo
}

// Mutation returns the AuthTokenMutation object of the builder.
func (atuo *AuthTokenUpdateOne) Mutation() *AuthTokenMutation {
	return atuo.mutation
}

// Where appends a list predicates to the AuthTokenUpdate builder.
func (atuo *AuthTokenUpdateOne) Where(ps ...predicate.AuthToken) *AuthTokenUpdateOne {
	atuo.mutation.Where(ps...)
	return atuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (atuo *AuthTokenUpdateOne) Select(field string, fields ...string) *AuthTokenUpdateOne {
	atuo.fields = append([]string{field}, fields...)
	return atuo
}

// Save executes the query and returns the updated AuthToken entity.
func (atuo *AuthTokenUpdateOne) Save(ctx context.Context) (*AuthToken, error) {
	atuo.defaults()
	return withHooks(ctx, atuo.sqlSave, atuo.mutation, atuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (atuo *AuthTokenUpdateOne) SaveX(ctx context.Context) *AuthToken {
	node, err := atuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (atuo *AuthTokenUpdateOne) Exec(ctx context.Context) error {
	_, err := atuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atuo *AuthTokenUpdateOne) ExecX(ctx context.Context) {
	if err := atuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (atuo *AuthTokenUpdateOne) defaults() {
	if _, ok := atuo.mutation.UpdatedAt(); !ok && !atuo.mutation.UpdatedAtCleared() {
		v := authtoken.UpdateDefaultUpdatedAt()
		atuo.mutation.SetUpdatedAt(v)
	}
}

func (atuo *AuthTokenUpdateOne) sqlSave(ctx context.Context) (_node *AuthToken, err error) {
	_spec := sqlgraph.NewUpdateSpec(authtoken.Table, authtoken.Columns, sqlgraph.NewFieldSpec(authtoken.FieldID, field.TypeString))
	id, ok := atuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "AuthToken.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := atuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, authtoken.FieldID)
		for _, f := range fields {
			if !authtoken.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != authtoken.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := atuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := atuo.mutation.Disabled(); ok {
		_spec.SetField(authtoken.FieldDisabled, field.TypeBool, value)
	}
	if atuo.mutation.DisabledCleared() {
		_spec.ClearField(authtoken.FieldDisabled, field.TypeBool)
	}
	if atuo.mutation.CreatedAtCleared() {
		_spec.ClearField(authtoken.FieldCreatedAt, field.TypeInt64)
	}
	if value, ok := atuo.mutation.UpdatedAt(); ok {
		_spec.SetField(authtoken.FieldUpdatedAt, field.TypeInt64, value)
	}
	if value, ok := atuo.mutation.AddedUpdatedAt(); ok {
		_spec.AddField(authtoken.FieldUpdatedAt, field.TypeInt64, value)
	}
	if atuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(authtoken.FieldUpdatedAt, field.TypeInt64)
	}
	if value, ok := atuo.mutation.UserID(); ok {
		_spec.SetField(authtoken.FieldUserID, field.TypeString, value)
	}
	if atuo.mutation.UserIDCleared() {
		_spec.ClearField(authtoken.FieldUserID, field.TypeString)
	}
	_node = &AuthToken{config: atuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, atuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{authtoken.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	atuo.mutation.done = true
	return _node, nil
}
