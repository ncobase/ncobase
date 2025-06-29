// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"ncobase/content/data/ent/taxonomy"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// TaxonomyCreate is the builder for creating a Taxonomy entity.
type TaxonomyCreate struct {
	config
	mutation *TaxonomyMutation
	hooks    []Hook
}

// SetName sets the "name" field.
func (tc *TaxonomyCreate) SetName(s string) *TaxonomyCreate {
	tc.mutation.SetName(s)
	return tc
}

// SetNillableName sets the "name" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableName(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetName(*s)
	}
	return tc
}

// SetType sets the "type" field.
func (tc *TaxonomyCreate) SetType(s string) *TaxonomyCreate {
	tc.mutation.SetType(s)
	return tc
}

// SetNillableType sets the "type" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableType(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetType(*s)
	}
	return tc
}

// SetSlug sets the "slug" field.
func (tc *TaxonomyCreate) SetSlug(s string) *TaxonomyCreate {
	tc.mutation.SetSlug(s)
	return tc
}

// SetNillableSlug sets the "slug" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableSlug(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetSlug(*s)
	}
	return tc
}

// SetCover sets the "cover" field.
func (tc *TaxonomyCreate) SetCover(s string) *TaxonomyCreate {
	tc.mutation.SetCover(s)
	return tc
}

// SetNillableCover sets the "cover" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableCover(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetCover(*s)
	}
	return tc
}

// SetThumbnail sets the "thumbnail" field.
func (tc *TaxonomyCreate) SetThumbnail(s string) *TaxonomyCreate {
	tc.mutation.SetThumbnail(s)
	return tc
}

// SetNillableThumbnail sets the "thumbnail" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableThumbnail(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetThumbnail(*s)
	}
	return tc
}

// SetColor sets the "color" field.
func (tc *TaxonomyCreate) SetColor(s string) *TaxonomyCreate {
	tc.mutation.SetColor(s)
	return tc
}

// SetNillableColor sets the "color" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableColor(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetColor(*s)
	}
	return tc
}

// SetIcon sets the "icon" field.
func (tc *TaxonomyCreate) SetIcon(s string) *TaxonomyCreate {
	tc.mutation.SetIcon(s)
	return tc
}

// SetNillableIcon sets the "icon" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableIcon(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetIcon(*s)
	}
	return tc
}

// SetURL sets the "url" field.
func (tc *TaxonomyCreate) SetURL(s string) *TaxonomyCreate {
	tc.mutation.SetURL(s)
	return tc
}

// SetNillableURL sets the "url" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableURL(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetURL(*s)
	}
	return tc
}

// SetKeywords sets the "keywords" field.
func (tc *TaxonomyCreate) SetKeywords(s string) *TaxonomyCreate {
	tc.mutation.SetKeywords(s)
	return tc
}

// SetNillableKeywords sets the "keywords" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableKeywords(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetKeywords(*s)
	}
	return tc
}

// SetDescription sets the "description" field.
func (tc *TaxonomyCreate) SetDescription(s string) *TaxonomyCreate {
	tc.mutation.SetDescription(s)
	return tc
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableDescription(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetDescription(*s)
	}
	return tc
}

// SetStatus sets the "status" field.
func (tc *TaxonomyCreate) SetStatus(i int) *TaxonomyCreate {
	tc.mutation.SetStatus(i)
	return tc
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableStatus(i *int) *TaxonomyCreate {
	if i != nil {
		tc.SetStatus(*i)
	}
	return tc
}

// SetExtras sets the "extras" field.
func (tc *TaxonomyCreate) SetExtras(m map[string]interface{}) *TaxonomyCreate {
	tc.mutation.SetExtras(m)
	return tc
}

// SetParentID sets the "parent_id" field.
func (tc *TaxonomyCreate) SetParentID(s string) *TaxonomyCreate {
	tc.mutation.SetParentID(s)
	return tc
}

// SetNillableParentID sets the "parent_id" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableParentID(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetParentID(*s)
	}
	return tc
}

// SetSpaceID sets the "space_id" field.
func (tc *TaxonomyCreate) SetSpaceID(s string) *TaxonomyCreate {
	tc.mutation.SetSpaceID(s)
	return tc
}

// SetNillableSpaceID sets the "space_id" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableSpaceID(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetSpaceID(*s)
	}
	return tc
}

// SetCreatedBy sets the "created_by" field.
func (tc *TaxonomyCreate) SetCreatedBy(s string) *TaxonomyCreate {
	tc.mutation.SetCreatedBy(s)
	return tc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableCreatedBy(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetCreatedBy(*s)
	}
	return tc
}

// SetUpdatedBy sets the "updated_by" field.
func (tc *TaxonomyCreate) SetUpdatedBy(s string) *TaxonomyCreate {
	tc.mutation.SetUpdatedBy(s)
	return tc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableUpdatedBy(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetUpdatedBy(*s)
	}
	return tc
}

// SetCreatedAt sets the "created_at" field.
func (tc *TaxonomyCreate) SetCreatedAt(i int64) *TaxonomyCreate {
	tc.mutation.SetCreatedAt(i)
	return tc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableCreatedAt(i *int64) *TaxonomyCreate {
	if i != nil {
		tc.SetCreatedAt(*i)
	}
	return tc
}

// SetUpdatedAt sets the "updated_at" field.
func (tc *TaxonomyCreate) SetUpdatedAt(i int64) *TaxonomyCreate {
	tc.mutation.SetUpdatedAt(i)
	return tc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableUpdatedAt(i *int64) *TaxonomyCreate {
	if i != nil {
		tc.SetUpdatedAt(*i)
	}
	return tc
}

// SetID sets the "id" field.
func (tc *TaxonomyCreate) SetID(s string) *TaxonomyCreate {
	tc.mutation.SetID(s)
	return tc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tc *TaxonomyCreate) SetNillableID(s *string) *TaxonomyCreate {
	if s != nil {
		tc.SetID(*s)
	}
	return tc
}

// Mutation returns the TaxonomyMutation object of the builder.
func (tc *TaxonomyCreate) Mutation() *TaxonomyMutation {
	return tc.mutation
}

// Save creates the Taxonomy in the database.
func (tc *TaxonomyCreate) Save(ctx context.Context) (*Taxonomy, error) {
	tc.defaults()
	return withHooks(ctx, tc.sqlSave, tc.mutation, tc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tc *TaxonomyCreate) SaveX(ctx context.Context) *Taxonomy {
	v, err := tc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tc *TaxonomyCreate) Exec(ctx context.Context) error {
	_, err := tc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tc *TaxonomyCreate) ExecX(ctx context.Context) {
	if err := tc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tc *TaxonomyCreate) defaults() {
	if _, ok := tc.mutation.Status(); !ok {
		v := taxonomy.DefaultStatus
		tc.mutation.SetStatus(v)
	}
	if _, ok := tc.mutation.Extras(); !ok {
		v := taxonomy.DefaultExtras
		tc.mutation.SetExtras(v)
	}
	if _, ok := tc.mutation.CreatedAt(); !ok {
		v := taxonomy.DefaultCreatedAt()
		tc.mutation.SetCreatedAt(v)
	}
	if _, ok := tc.mutation.UpdatedAt(); !ok {
		v := taxonomy.DefaultUpdatedAt()
		tc.mutation.SetUpdatedAt(v)
	}
	if _, ok := tc.mutation.ID(); !ok {
		v := taxonomy.DefaultID()
		tc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tc *TaxonomyCreate) check() error {
	if _, ok := tc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "Taxonomy.status"`)}
	}
	if v, ok := tc.mutation.ID(); ok {
		if err := taxonomy.IDValidator(v); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`ent: validator failed for field "Taxonomy.id": %w`, err)}
		}
	}
	return nil
}

func (tc *TaxonomyCreate) sqlSave(ctx context.Context) (*Taxonomy, error) {
	if err := tc.check(); err != nil {
		return nil, err
	}
	_node, _spec := tc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected Taxonomy.ID type: %T", _spec.ID.Value)
		}
	}
	tc.mutation.id = &_node.ID
	tc.mutation.done = true
	return _node, nil
}

func (tc *TaxonomyCreate) createSpec() (*Taxonomy, *sqlgraph.CreateSpec) {
	var (
		_node = &Taxonomy{config: tc.config}
		_spec = sqlgraph.NewCreateSpec(taxonomy.Table, sqlgraph.NewFieldSpec(taxonomy.FieldID, field.TypeString))
	)
	if id, ok := tc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := tc.mutation.Name(); ok {
		_spec.SetField(taxonomy.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := tc.mutation.GetType(); ok {
		_spec.SetField(taxonomy.FieldType, field.TypeString, value)
		_node.Type = value
	}
	if value, ok := tc.mutation.Slug(); ok {
		_spec.SetField(taxonomy.FieldSlug, field.TypeString, value)
		_node.Slug = value
	}
	if value, ok := tc.mutation.Cover(); ok {
		_spec.SetField(taxonomy.FieldCover, field.TypeString, value)
		_node.Cover = value
	}
	if value, ok := tc.mutation.Thumbnail(); ok {
		_spec.SetField(taxonomy.FieldThumbnail, field.TypeString, value)
		_node.Thumbnail = value
	}
	if value, ok := tc.mutation.Color(); ok {
		_spec.SetField(taxonomy.FieldColor, field.TypeString, value)
		_node.Color = value
	}
	if value, ok := tc.mutation.Icon(); ok {
		_spec.SetField(taxonomy.FieldIcon, field.TypeString, value)
		_node.Icon = value
	}
	if value, ok := tc.mutation.URL(); ok {
		_spec.SetField(taxonomy.FieldURL, field.TypeString, value)
		_node.URL = value
	}
	if value, ok := tc.mutation.Keywords(); ok {
		_spec.SetField(taxonomy.FieldKeywords, field.TypeString, value)
		_node.Keywords = value
	}
	if value, ok := tc.mutation.Description(); ok {
		_spec.SetField(taxonomy.FieldDescription, field.TypeString, value)
		_node.Description = value
	}
	if value, ok := tc.mutation.Status(); ok {
		_spec.SetField(taxonomy.FieldStatus, field.TypeInt, value)
		_node.Status = value
	}
	if value, ok := tc.mutation.Extras(); ok {
		_spec.SetField(taxonomy.FieldExtras, field.TypeJSON, value)
		_node.Extras = value
	}
	if value, ok := tc.mutation.ParentID(); ok {
		_spec.SetField(taxonomy.FieldParentID, field.TypeString, value)
		_node.ParentID = value
	}
	if value, ok := tc.mutation.SpaceID(); ok {
		_spec.SetField(taxonomy.FieldSpaceID, field.TypeString, value)
		_node.SpaceID = value
	}
	if value, ok := tc.mutation.CreatedBy(); ok {
		_spec.SetField(taxonomy.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := tc.mutation.UpdatedBy(); ok {
		_spec.SetField(taxonomy.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := tc.mutation.CreatedAt(); ok {
		_spec.SetField(taxonomy.FieldCreatedAt, field.TypeInt64, value)
		_node.CreatedAt = value
	}
	if value, ok := tc.mutation.UpdatedAt(); ok {
		_spec.SetField(taxonomy.FieldUpdatedAt, field.TypeInt64, value)
		_node.UpdatedAt = value
	}
	return _node, _spec
}

// TaxonomyCreateBulk is the builder for creating many Taxonomy entities in bulk.
type TaxonomyCreateBulk struct {
	config
	err      error
	builders []*TaxonomyCreate
}

// Save creates the Taxonomy entities in the database.
func (tcb *TaxonomyCreateBulk) Save(ctx context.Context) ([]*Taxonomy, error) {
	if tcb.err != nil {
		return nil, tcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tcb.builders))
	nodes := make([]*Taxonomy, len(tcb.builders))
	mutators := make([]Mutator, len(tcb.builders))
	for i := range tcb.builders {
		func(i int, root context.Context) {
			builder := tcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TaxonomyMutation)
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
					_, err = mutators[i+1].Mutate(root, tcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, tcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tcb *TaxonomyCreateBulk) SaveX(ctx context.Context) []*Taxonomy {
	v, err := tcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcb *TaxonomyCreateBulk) Exec(ctx context.Context) error {
	_, err := tcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcb *TaxonomyCreateBulk) ExecX(ctx context.Context) {
	if err := tcb.Exec(ctx); err != nil {
		panic(err)
	}
}
