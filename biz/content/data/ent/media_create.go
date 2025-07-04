// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"ncobase/content/data/ent/media"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// MediaCreate is the builder for creating a Media entity.
type MediaCreate struct {
	config
	mutation *MediaMutation
	hooks    []Hook
}

// SetTitle sets the "title" field.
func (mc *MediaCreate) SetTitle(s string) *MediaCreate {
	mc.mutation.SetTitle(s)
	return mc
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (mc *MediaCreate) SetNillableTitle(s *string) *MediaCreate {
	if s != nil {
		mc.SetTitle(*s)
	}
	return mc
}

// SetType sets the "type" field.
func (mc *MediaCreate) SetType(s string) *MediaCreate {
	mc.mutation.SetType(s)
	return mc
}

// SetNillableType sets the "type" field if the given value is not nil.
func (mc *MediaCreate) SetNillableType(s *string) *MediaCreate {
	if s != nil {
		mc.SetType(*s)
	}
	return mc
}

// SetURL sets the "url" field.
func (mc *MediaCreate) SetURL(s string) *MediaCreate {
	mc.mutation.SetURL(s)
	return mc
}

// SetNillableURL sets the "url" field if the given value is not nil.
func (mc *MediaCreate) SetNillableURL(s *string) *MediaCreate {
	if s != nil {
		mc.SetURL(*s)
	}
	return mc
}

// SetExtras sets the "extras" field.
func (mc *MediaCreate) SetExtras(m map[string]interface{}) *MediaCreate {
	mc.mutation.SetExtras(m)
	return mc
}

// SetSpaceID sets the "space_id" field.
func (mc *MediaCreate) SetSpaceID(s string) *MediaCreate {
	mc.mutation.SetSpaceID(s)
	return mc
}

// SetNillableSpaceID sets the "space_id" field if the given value is not nil.
func (mc *MediaCreate) SetNillableSpaceID(s *string) *MediaCreate {
	if s != nil {
		mc.SetSpaceID(*s)
	}
	return mc
}

// SetCreatedBy sets the "created_by" field.
func (mc *MediaCreate) SetCreatedBy(s string) *MediaCreate {
	mc.mutation.SetCreatedBy(s)
	return mc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (mc *MediaCreate) SetNillableCreatedBy(s *string) *MediaCreate {
	if s != nil {
		mc.SetCreatedBy(*s)
	}
	return mc
}

// SetUpdatedBy sets the "updated_by" field.
func (mc *MediaCreate) SetUpdatedBy(s string) *MediaCreate {
	mc.mutation.SetUpdatedBy(s)
	return mc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (mc *MediaCreate) SetNillableUpdatedBy(s *string) *MediaCreate {
	if s != nil {
		mc.SetUpdatedBy(*s)
	}
	return mc
}

// SetCreatedAt sets the "created_at" field.
func (mc *MediaCreate) SetCreatedAt(i int64) *MediaCreate {
	mc.mutation.SetCreatedAt(i)
	return mc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (mc *MediaCreate) SetNillableCreatedAt(i *int64) *MediaCreate {
	if i != nil {
		mc.SetCreatedAt(*i)
	}
	return mc
}

// SetUpdatedAt sets the "updated_at" field.
func (mc *MediaCreate) SetUpdatedAt(i int64) *MediaCreate {
	mc.mutation.SetUpdatedAt(i)
	return mc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (mc *MediaCreate) SetNillableUpdatedAt(i *int64) *MediaCreate {
	if i != nil {
		mc.SetUpdatedAt(*i)
	}
	return mc
}

// SetOwnerID sets the "owner_id" field.
func (mc *MediaCreate) SetOwnerID(s string) *MediaCreate {
	mc.mutation.SetOwnerID(s)
	return mc
}

// SetResourceID sets the "resource_id" field.
func (mc *MediaCreate) SetResourceID(s string) *MediaCreate {
	mc.mutation.SetResourceID(s)
	return mc
}

// SetNillableResourceID sets the "resource_id" field if the given value is not nil.
func (mc *MediaCreate) SetNillableResourceID(s *string) *MediaCreate {
	if s != nil {
		mc.SetResourceID(*s)
	}
	return mc
}

// SetDescription sets the "description" field.
func (mc *MediaCreate) SetDescription(s string) *MediaCreate {
	mc.mutation.SetDescription(s)
	return mc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (mc *MediaCreate) SetNillableDescription(s *string) *MediaCreate {
	if s != nil {
		mc.SetDescription(*s)
	}
	return mc
}

// SetAlt sets the "alt" field.
func (mc *MediaCreate) SetAlt(s string) *MediaCreate {
	mc.mutation.SetAlt(s)
	return mc
}

// SetNillableAlt sets the "alt" field if the given value is not nil.
func (mc *MediaCreate) SetNillableAlt(s *string) *MediaCreate {
	if s != nil {
		mc.SetAlt(*s)
	}
	return mc
}

// SetID sets the "id" field.
func (mc *MediaCreate) SetID(s string) *MediaCreate {
	mc.mutation.SetID(s)
	return mc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (mc *MediaCreate) SetNillableID(s *string) *MediaCreate {
	if s != nil {
		mc.SetID(*s)
	}
	return mc
}

// Mutation returns the MediaMutation object of the builder.
func (mc *MediaCreate) Mutation() *MediaMutation {
	return mc.mutation
}

// Save creates the Media in the database.
func (mc *MediaCreate) Save(ctx context.Context) (*Media, error) {
	mc.defaults()
	return withHooks(ctx, mc.sqlSave, mc.mutation, mc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (mc *MediaCreate) SaveX(ctx context.Context) *Media {
	v, err := mc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mc *MediaCreate) Exec(ctx context.Context) error {
	_, err := mc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mc *MediaCreate) ExecX(ctx context.Context) {
	if err := mc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (mc *MediaCreate) defaults() {
	if _, ok := mc.mutation.Extras(); !ok {
		v := media.DefaultExtras
		mc.mutation.SetExtras(v)
	}
	if _, ok := mc.mutation.CreatedAt(); !ok {
		v := media.DefaultCreatedAt()
		mc.mutation.SetCreatedAt(v)
	}
	if _, ok := mc.mutation.UpdatedAt(); !ok {
		v := media.DefaultUpdatedAt()
		mc.mutation.SetUpdatedAt(v)
	}
	if _, ok := mc.mutation.ID(); !ok {
		v := media.DefaultID()
		mc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (mc *MediaCreate) check() error {
	if _, ok := mc.mutation.OwnerID(); !ok {
		return &ValidationError{Name: "owner_id", err: errors.New(`ent: missing required field "Media.owner_id"`)}
	}
	if v, ok := mc.mutation.OwnerID(); ok {
		if err := media.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`ent: validator failed for field "Media.owner_id": %w`, err)}
		}
	}
	if v, ok := mc.mutation.ID(); ok {
		if err := media.IDValidator(v); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`ent: validator failed for field "Media.id": %w`, err)}
		}
	}
	return nil
}

func (mc *MediaCreate) sqlSave(ctx context.Context) (*Media, error) {
	if err := mc.check(); err != nil {
		return nil, err
	}
	_node, _spec := mc.createSpec()
	if err := sqlgraph.CreateNode(ctx, mc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Media.ID type: %T", _spec.ID.Value)
		}
	}
	mc.mutation.id = &_node.ID
	mc.mutation.done = true
	return _node, nil
}

func (mc *MediaCreate) createSpec() (*Media, *sqlgraph.CreateSpec) {
	var (
		_node = &Media{config: mc.config}
		_spec = sqlgraph.NewCreateSpec(media.Table, sqlgraph.NewFieldSpec(media.FieldID, field.TypeString))
	)
	if id, ok := mc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := mc.mutation.Title(); ok {
		_spec.SetField(media.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := mc.mutation.GetType(); ok {
		_spec.SetField(media.FieldType, field.TypeString, value)
		_node.Type = value
	}
	if value, ok := mc.mutation.URL(); ok {
		_spec.SetField(media.FieldURL, field.TypeString, value)
		_node.URL = value
	}
	if value, ok := mc.mutation.Extras(); ok {
		_spec.SetField(media.FieldExtras, field.TypeJSON, value)
		_node.Extras = value
	}
	if value, ok := mc.mutation.SpaceID(); ok {
		_spec.SetField(media.FieldSpaceID, field.TypeString, value)
		_node.SpaceID = value
	}
	if value, ok := mc.mutation.CreatedBy(); ok {
		_spec.SetField(media.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := mc.mutation.UpdatedBy(); ok {
		_spec.SetField(media.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := mc.mutation.CreatedAt(); ok {
		_spec.SetField(media.FieldCreatedAt, field.TypeInt64, value)
		_node.CreatedAt = value
	}
	if value, ok := mc.mutation.UpdatedAt(); ok {
		_spec.SetField(media.FieldUpdatedAt, field.TypeInt64, value)
		_node.UpdatedAt = value
	}
	if value, ok := mc.mutation.OwnerID(); ok {
		_spec.SetField(media.FieldOwnerID, field.TypeString, value)
		_node.OwnerID = value
	}
	if value, ok := mc.mutation.ResourceID(); ok {
		_spec.SetField(media.FieldResourceID, field.TypeString, value)
		_node.ResourceID = value
	}
	if value, ok := mc.mutation.Description(); ok {
		_spec.SetField(media.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := mc.mutation.Alt(); ok {
		_spec.SetField(media.FieldAlt, field.TypeString, value)
		_node.Alt = value
	}
	return _node, _spec
}

// MediaCreateBulk is the builder for creating many Media entities in bulk.
type MediaCreateBulk struct {
	config
	err      error
	builders []*MediaCreate
}

// Save creates the Media entities in the database.
func (mcb *MediaCreateBulk) Save(ctx context.Context) ([]*Media, error) {
	if mcb.err != nil {
		return nil, mcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(mcb.builders))
	nodes := make([]*Media, len(mcb.builders))
	mutators := make([]Mutator, len(mcb.builders))
	for i := range mcb.builders {
		func(i int, root context.Context) {
			builder := mcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*MediaMutation)
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
					_, err = mutators[i+1].Mutate(root, mcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, mcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, mcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (mcb *MediaCreateBulk) SaveX(ctx context.Context) []*Media {
	v, err := mcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mcb *MediaCreateBulk) Exec(ctx context.Context) error {
	_, err := mcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mcb *MediaCreateBulk) ExecX(ctx context.Context) {
	if err := mcb.Exec(ctx); err != nil {
		panic(err)
	}
}
