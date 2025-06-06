// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"ncobase/system/data/ent/dictionary"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// DictionaryCreate is the builder for creating a Dictionary entity.
type DictionaryCreate struct {
	config
	mutation *DictionaryMutation
	hooks    []Hook
}

// SetName sets the "name" field.
func (dc *DictionaryCreate) SetName(s string) *DictionaryCreate {
	dc.mutation.SetName(s)
	return dc
}

// SetNillableName sets the "name" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableName(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetName(*s)
	}
	return dc
}

// SetSlug sets the "slug" field.
func (dc *DictionaryCreate) SetSlug(s string) *DictionaryCreate {
	dc.mutation.SetSlug(s)
	return dc
}

// SetNillableSlug sets the "slug" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableSlug(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetSlug(*s)
	}
	return dc
}

// SetType sets the "type" field.
func (dc *DictionaryCreate) SetType(s string) *DictionaryCreate {
	dc.mutation.SetType(s)
	return dc
}

// SetNillableType sets the "type" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableType(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetType(*s)
	}
	return dc
}

// SetValue sets the "value" field.
func (dc *DictionaryCreate) SetValue(s string) *DictionaryCreate {
	dc.mutation.SetValue(s)
	return dc
}

// SetNillableValue sets the "value" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableValue(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetValue(*s)
	}
	return dc
}

// SetDescription sets the "description" field.
func (dc *DictionaryCreate) SetDescription(s string) *DictionaryCreate {
	dc.mutation.SetDescription(s)
	return dc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableDescription(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetDescription(*s)
	}
	return dc
}

// SetCreatedBy sets the "created_by" field.
func (dc *DictionaryCreate) SetCreatedBy(s string) *DictionaryCreate {
	dc.mutation.SetCreatedBy(s)
	return dc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableCreatedBy(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetCreatedBy(*s)
	}
	return dc
}

// SetUpdatedBy sets the "updated_by" field.
func (dc *DictionaryCreate) SetUpdatedBy(s string) *DictionaryCreate {
	dc.mutation.SetUpdatedBy(s)
	return dc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableUpdatedBy(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetUpdatedBy(*s)
	}
	return dc
}

// SetCreatedAt sets the "created_at" field.
func (dc *DictionaryCreate) SetCreatedAt(i int64) *DictionaryCreate {
	dc.mutation.SetCreatedAt(i)
	return dc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableCreatedAt(i *int64) *DictionaryCreate {
	if i != nil {
		dc.SetCreatedAt(*i)
	}
	return dc
}

// SetUpdatedAt sets the "updated_at" field.
func (dc *DictionaryCreate) SetUpdatedAt(i int64) *DictionaryCreate {
	dc.mutation.SetUpdatedAt(i)
	return dc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableUpdatedAt(i *int64) *DictionaryCreate {
	if i != nil {
		dc.SetUpdatedAt(*i)
	}
	return dc
}

// SetID sets the "id" field.
func (dc *DictionaryCreate) SetID(s string) *DictionaryCreate {
	dc.mutation.SetID(s)
	return dc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (dc *DictionaryCreate) SetNillableID(s *string) *DictionaryCreate {
	if s != nil {
		dc.SetID(*s)
	}
	return dc
}

// Mutation returns the DictionaryMutation object of the builder.
func (dc *DictionaryCreate) Mutation() *DictionaryMutation {
	return dc.mutation
}

// Save creates the Dictionary in the database.
func (dc *DictionaryCreate) Save(ctx context.Context) (*Dictionary, error) {
	dc.defaults()
	return withHooks(ctx, dc.sqlSave, dc.mutation, dc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (dc *DictionaryCreate) SaveX(ctx context.Context) *Dictionary {
	v, err := dc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dc *DictionaryCreate) Exec(ctx context.Context) error {
	_, err := dc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dc *DictionaryCreate) ExecX(ctx context.Context) {
	if err := dc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (dc *DictionaryCreate) defaults() {
	if _, ok := dc.mutation.CreatedAt(); !ok {
		v := dictionary.DefaultCreatedAt()
		dc.mutation.SetCreatedAt(v)
	}
	if _, ok := dc.mutation.UpdatedAt(); !ok {
		v := dictionary.DefaultUpdatedAt()
		dc.mutation.SetUpdatedAt(v)
	}
	if _, ok := dc.mutation.ID(); !ok {
		v := dictionary.DefaultID()
		dc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (dc *DictionaryCreate) check() error {
	if v, ok := dc.mutation.ID(); ok {
		if err := dictionary.IDValidator(v); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`ent: validator failed for field "Dictionary.id": %w`, err)}
		}
	}
	return nil
}

func (dc *DictionaryCreate) sqlSave(ctx context.Context) (*Dictionary, error) {
	if err := dc.check(); err != nil {
		return nil, err
	}
	_node, _spec := dc.createSpec()
	if err := sqlgraph.CreateNode(ctx, dc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Dictionary.ID type: %T", _spec.ID.Value)
		}
	}
	dc.mutation.id = &_node.ID
	dc.mutation.done = true
	return _node, nil
}

func (dc *DictionaryCreate) createSpec() (*Dictionary, *sqlgraph.CreateSpec) {
	var (
		_node = &Dictionary{config: dc.config}
		_spec = sqlgraph.NewCreateSpec(dictionary.Table, sqlgraph.NewFieldSpec(dictionary.FieldID, field.TypeString))
	)
	if id, ok := dc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := dc.mutation.Name(); ok {
		_spec.SetField(dictionary.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := dc.mutation.Slug(); ok {
		_spec.SetField(dictionary.FieldSlug, field.TypeString, value)
		_node.Slug = value
	}
	if value, ok := dc.mutation.GetType(); ok {
		_spec.SetField(dictionary.FieldType, field.TypeString, value)
		_node.Type = value
	}
	if value, ok := dc.mutation.Value(); ok {
		_spec.SetField(dictionary.FieldValue, field.TypeString, value)
		_node.Value = value
	}
	if value, ok := dc.mutation.Description(); ok {
		_spec.SetField(dictionary.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := dc.mutation.CreatedBy(); ok {
		_spec.SetField(dictionary.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := dc.mutation.UpdatedBy(); ok {
		_spec.SetField(dictionary.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := dc.mutation.CreatedAt(); ok {
		_spec.SetField(dictionary.FieldCreatedAt, field.TypeInt64, value)
		_node.CreatedAt = value
	}
	if value, ok := dc.mutation.UpdatedAt(); ok {
		_spec.SetField(dictionary.FieldUpdatedAt, field.TypeInt64, value)
		_node.UpdatedAt = value
	}
	return _node, _spec
}

// DictionaryCreateBulk is the builder for creating many Dictionary entities in bulk.
type DictionaryCreateBulk struct {
	config
	err      error
	builders []*DictionaryCreate
}

// Save creates the Dictionary entities in the database.
func (dcb *DictionaryCreateBulk) Save(ctx context.Context) ([]*Dictionary, error) {
	if dcb.err != nil {
		return nil, dcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(dcb.builders))
	nodes := make([]*Dictionary, len(dcb.builders))
	mutators := make([]Mutator, len(dcb.builders))
	for i := range dcb.builders {
		func(i int, root context.Context) {
			builder := dcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*DictionaryMutation)
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
					_, err = mutators[i+1].Mutate(root, dcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, dcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, dcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (dcb *DictionaryCreateBulk) SaveX(ctx context.Context) []*Dictionary {
	v, err := dcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (dcb *DictionaryCreateBulk) Exec(ctx context.Context) error {
	_, err := dcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dcb *DictionaryCreateBulk) ExecX(ctx context.Context) {
	if err := dcb.Exec(ctx); err != nil {
		panic(err)
	}
}
