// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"ncobase/auth/data/ent/authtoken"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// AuthTokenCreate is the builder for creating a AuthToken entity.
type AuthTokenCreate struct {
	config
	mutation *AuthTokenMutation
	hooks    []Hook
}

// SetDisabled sets the "disabled" field.
func (atc *AuthTokenCreate) SetDisabled(b bool) *AuthTokenCreate {
	atc.mutation.SetDisabled(b)
	return atc
}

// SetNillableDisabled sets the "disabled" field if the given value is not nil.
func (atc *AuthTokenCreate) SetNillableDisabled(b *bool) *AuthTokenCreate {
	if b != nil {
		atc.SetDisabled(*b)
	}
	return atc
}

// SetCreatedAt sets the "created_at" field.
func (atc *AuthTokenCreate) SetCreatedAt(i int64) *AuthTokenCreate {
	atc.mutation.SetCreatedAt(i)
	return atc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (atc *AuthTokenCreate) SetNillableCreatedAt(i *int64) *AuthTokenCreate {
	if i != nil {
		atc.SetCreatedAt(*i)
	}
	return atc
}

// SetUpdatedAt sets the "updated_at" field.
func (atc *AuthTokenCreate) SetUpdatedAt(i int64) *AuthTokenCreate {
	atc.mutation.SetUpdatedAt(i)
	return atc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (atc *AuthTokenCreate) SetNillableUpdatedAt(i *int64) *AuthTokenCreate {
	if i != nil {
		atc.SetUpdatedAt(*i)
	}
	return atc
}

// SetUserID sets the "user_id" field.
func (atc *AuthTokenCreate) SetUserID(s string) *AuthTokenCreate {
	atc.mutation.SetUserID(s)
	return atc
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (atc *AuthTokenCreate) SetNillableUserID(s *string) *AuthTokenCreate {
	if s != nil {
		atc.SetUserID(*s)
	}
	return atc
}

// SetID sets the "id" field.
func (atc *AuthTokenCreate) SetID(s string) *AuthTokenCreate {
	atc.mutation.SetID(s)
	return atc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (atc *AuthTokenCreate) SetNillableID(s *string) *AuthTokenCreate {
	if s != nil {
		atc.SetID(*s)
	}
	return atc
}

// Mutation returns the AuthTokenMutation object of the builder.
func (atc *AuthTokenCreate) Mutation() *AuthTokenMutation {
	return atc.mutation
}

// Save creates the AuthToken in the database.
func (atc *AuthTokenCreate) Save(ctx context.Context) (*AuthToken, error) {
	atc.defaults()
	return withHooks(ctx, atc.sqlSave, atc.mutation, atc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (atc *AuthTokenCreate) SaveX(ctx context.Context) *AuthToken {
	v, err := atc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (atc *AuthTokenCreate) Exec(ctx context.Context) error {
	_, err := atc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atc *AuthTokenCreate) ExecX(ctx context.Context) {
	if err := atc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (atc *AuthTokenCreate) defaults() {
	if _, ok := atc.mutation.Disabled(); !ok {
		v := authtoken.DefaultDisabled
		atc.mutation.SetDisabled(v)
	}
	if _, ok := atc.mutation.CreatedAt(); !ok {
		v := authtoken.DefaultCreatedAt()
		atc.mutation.SetCreatedAt(v)
	}
	if _, ok := atc.mutation.UpdatedAt(); !ok {
		v := authtoken.DefaultUpdatedAt()
		atc.mutation.SetUpdatedAt(v)
	}
	if _, ok := atc.mutation.ID(); !ok {
		v := authtoken.DefaultID()
		atc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (atc *AuthTokenCreate) check() error {
	if v, ok := atc.mutation.ID(); ok {
		if err := authtoken.IDValidator(v); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`ent: validator failed for field "AuthToken.id": %w`, err)}
		}
	}
	return nil
}

func (atc *AuthTokenCreate) sqlSave(ctx context.Context) (*AuthToken, error) {
	if err := atc.check(); err != nil {
		return nil, err
	}
	_node, _spec := atc.createSpec()
	if err := sqlgraph.CreateNode(ctx, atc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected AuthToken.ID type: %T", _spec.ID.Value)
		}
	}
	atc.mutation.id = &_node.ID
	atc.mutation.done = true
	return _node, nil
}

func (atc *AuthTokenCreate) createSpec() (*AuthToken, *sqlgraph.CreateSpec) {
	var (
		_node = &AuthToken{config: atc.config}
		_spec = sqlgraph.NewCreateSpec(authtoken.Table, sqlgraph.NewFieldSpec(authtoken.FieldID, field.TypeString))
	)
	if id, ok := atc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := atc.mutation.Disabled(); ok {
		_spec.SetField(authtoken.FieldDisabled, field.TypeBool, value)
		_node.Disabled = value
	}
	if value, ok := atc.mutation.CreatedAt(); ok {
		_spec.SetField(authtoken.FieldCreatedAt, field.TypeInt64, value)
		_node.CreatedAt = value
	}
	if value, ok := atc.mutation.UpdatedAt(); ok {
		_spec.SetField(authtoken.FieldUpdatedAt, field.TypeInt64, value)
		_node.UpdatedAt = value
	}
	if value, ok := atc.mutation.UserID(); ok {
		_spec.SetField(authtoken.FieldUserID, field.TypeString, value)
		_node.UserID = value
	}
	return _node, _spec
}

// AuthTokenCreateBulk is the builder for creating many AuthToken entities in bulk.
type AuthTokenCreateBulk struct {
	config
	err      error
	builders []*AuthTokenCreate
}

// Save creates the AuthToken entities in the database.
func (atcb *AuthTokenCreateBulk) Save(ctx context.Context) ([]*AuthToken, error) {
	if atcb.err != nil {
		return nil, atcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(atcb.builders))
	nodes := make([]*AuthToken, len(atcb.builders))
	mutators := make([]Mutator, len(atcb.builders))
	for i := range atcb.builders {
		func(i int, root context.Context) {
			builder := atcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*AuthTokenMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, atcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, atcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, atcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (atcb *AuthTokenCreateBulk) SaveX(ctx context.Context) []*AuthToken {
	v, err := atcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (atcb *AuthTokenCreateBulk) Exec(ctx context.Context) error {
	_, err := atcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (atcb *AuthTokenCreateBulk) ExecX(ctx context.Context) {
	if err := atcb.Exec(ctx); err != nil {
		panic(err)
	}
}
