// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"ncobase/user/data/ent/userprofile"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// UserProfileCreate is the builder for creating a UserProfile entity.
type UserProfileCreate struct {
	config
	mutation *UserProfileMutation
	hooks    []Hook
}

// SetDisplayName sets the "display_name" field.
func (upc *UserProfileCreate) SetDisplayName(s string) *UserProfileCreate {
	upc.mutation.SetDisplayName(s)
	return upc
}

// SetNillableDisplayName sets the "display_name" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableDisplayName(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetDisplayName(*s)
	}
	return upc
}

// SetFirstName sets the "first_name" field.
func (upc *UserProfileCreate) SetFirstName(s string) *UserProfileCreate {
	upc.mutation.SetFirstName(s)
	return upc
}

// SetNillableFirstName sets the "first_name" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableFirstName(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetFirstName(*s)
	}
	return upc
}

// SetLastName sets the "last_name" field.
func (upc *UserProfileCreate) SetLastName(s string) *UserProfileCreate {
	upc.mutation.SetLastName(s)
	return upc
}

// SetNillableLastName sets the "last_name" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableLastName(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetLastName(*s)
	}
	return upc
}

// SetTitle sets the "title" field.
func (upc *UserProfileCreate) SetTitle(s string) *UserProfileCreate {
	upc.mutation.SetTitle(s)
	return upc
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableTitle(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetTitle(*s)
	}
	return upc
}

// SetShortBio sets the "short_bio" field.
func (upc *UserProfileCreate) SetShortBio(s string) *UserProfileCreate {
	upc.mutation.SetShortBio(s)
	return upc
}

// SetNillableShortBio sets the "short_bio" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableShortBio(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetShortBio(*s)
	}
	return upc
}

// SetAbout sets the "about" field.
func (upc *UserProfileCreate) SetAbout(s string) *UserProfileCreate {
	upc.mutation.SetAbout(s)
	return upc
}

// SetNillableAbout sets the "about" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableAbout(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetAbout(*s)
	}
	return upc
}

// SetLinks sets the "links" field.
func (upc *UserProfileCreate) SetLinks(m []map[string]interface{}) *UserProfileCreate {
	upc.mutation.SetLinks(m)
	return upc
}

// SetThumbnail sets the "thumbnail" field.
func (upc *UserProfileCreate) SetThumbnail(s string) *UserProfileCreate {
	upc.mutation.SetThumbnail(s)
	return upc
}

// SetNillableThumbnail sets the "thumbnail" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableThumbnail(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetThumbnail(*s)
	}
	return upc
}

// SetExtras sets the "extras" field.
func (upc *UserProfileCreate) SetExtras(m map[string]interface{}) *UserProfileCreate {
	upc.mutation.SetExtras(m)
	return upc
}

// SetID sets the "id" field.
func (upc *UserProfileCreate) SetID(s string) *UserProfileCreate {
	upc.mutation.SetID(s)
	return upc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (upc *UserProfileCreate) SetNillableID(s *string) *UserProfileCreate {
	if s != nil {
		upc.SetID(*s)
	}
	return upc
}

// Mutation returns the UserProfileMutation object of the builder.
func (upc *UserProfileCreate) Mutation() *UserProfileMutation {
	return upc.mutation
}

// Save creates the UserProfile in the database.
func (upc *UserProfileCreate) Save(ctx context.Context) (*UserProfile, error) {
	upc.defaults()
	return withHooks(ctx, upc.sqlSave, upc.mutation, upc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (upc *UserProfileCreate) SaveX(ctx context.Context) *UserProfile {
	v, err := upc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (upc *UserProfileCreate) Exec(ctx context.Context) error {
	_, err := upc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (upc *UserProfileCreate) ExecX(ctx context.Context) {
	if err := upc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (upc *UserProfileCreate) defaults() {
	if _, ok := upc.mutation.Links(); !ok {
		v := userprofile.DefaultLinks
		upc.mutation.SetLinks(v)
	}
	if _, ok := upc.mutation.Extras(); !ok {
		v := userprofile.DefaultExtras
		upc.mutation.SetExtras(v)
	}
	if _, ok := upc.mutation.ID(); !ok {
		v := userprofile.DefaultID()
		upc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (upc *UserProfileCreate) check() error {
	return nil
}

func (upc *UserProfileCreate) sqlSave(ctx context.Context) (*UserProfile, error) {
	if err := upc.check(); err != nil {
		return nil, err
	}
	_node, _spec := upc.createSpec()
	if err := sqlgraph.CreateNode(ctx, upc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected UserProfile.ID type: %T", _spec.ID.Value)
		}
	}
	upc.mutation.id = &_node.ID
	upc.mutation.done = true
	return _node, nil
}

func (upc *UserProfileCreate) createSpec() (*UserProfile, *sqlgraph.CreateSpec) {
	var (
		_node = &UserProfile{config: upc.config}
		_spec = sqlgraph.NewCreateSpec(userprofile.Table, sqlgraph.NewFieldSpec(userprofile.FieldID, field.TypeString))
	)
	if id, ok := upc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := upc.mutation.DisplayName(); ok {
		_spec.SetField(userprofile.FieldDisplayName, field.TypeString, value)
		_node.DisplayName = value
	}
	if value, ok := upc.mutation.FirstName(); ok {
		_spec.SetField(userprofile.FieldFirstName, field.TypeString, value)
		_node.FirstName = value
	}
	if value, ok := upc.mutation.LastName(); ok {
		_spec.SetField(userprofile.FieldLastName, field.TypeString, value)
		_node.LastName = value
	}
	if value, ok := upc.mutation.Title(); ok {
		_spec.SetField(userprofile.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := upc.mutation.ShortBio(); ok {
		_spec.SetField(userprofile.FieldShortBio, field.TypeString, value)
		_node.ShortBio = value
	}
	if value, ok := upc.mutation.About(); ok {
		_spec.SetField(userprofile.FieldAbout, field.TypeString, value)
		_node.About = value
	}
	if value, ok := upc.mutation.Links(); ok {
		_spec.SetField(userprofile.FieldLinks, field.TypeJSON, value)
		_node.Links = value
	}
	if value, ok := upc.mutation.Thumbnail(); ok {
		_spec.SetField(userprofile.FieldThumbnail, field.TypeString, value)
		_node.Thumbnail = value
	}
	if value, ok := upc.mutation.Extras(); ok {
		_spec.SetField(userprofile.FieldExtras, field.TypeJSON, value)
		_node.Extras = value
	}
	return _node, _spec
}

// UserProfileCreateBulk is the builder for creating many UserProfile entities in bulk.
type UserProfileCreateBulk struct {
	config
	err      error
	builders []*UserProfileCreate
}

// Save creates the UserProfile entities in the database.
func (upcb *UserProfileCreateBulk) Save(ctx context.Context) ([]*UserProfile, error) {
	if upcb.err != nil {
		return nil, upcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(upcb.builders))
	nodes := make([]*UserProfile, len(upcb.builders))
	mutators := make([]Mutator, len(upcb.builders))
	for i := range upcb.builders {
		func(i int, root context.Context) {
			builder := upcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*UserProfileMutation)
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
					_, err = mutators[i+1].Mutate(root, upcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, upcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, upcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (upcb *UserProfileCreateBulk) SaveX(ctx context.Context) []*UserProfile {
	v, err := upcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (upcb *UserProfileCreateBulk) Exec(ctx context.Context) error {
	_, err := upcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (upcb *UserProfileCreateBulk) ExecX(ctx context.Context) {
	if err := upcb.Exec(ctx); err != nil {
		panic(err)
	}
}
