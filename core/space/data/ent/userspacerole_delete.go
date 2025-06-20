// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"ncobase/space/data/ent/predicate"
	"ncobase/space/data/ent/userspacerole"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// UserSpaceRoleDelete is the builder for deleting a UserSpaceRole entity.
type UserSpaceRoleDelete struct {
	config
	hooks    []Hook
	mutation *UserSpaceRoleMutation
}

// Where appends a list predicates to the UserSpaceRoleDelete builder.
func (usrd *UserSpaceRoleDelete) Where(ps ...predicate.UserSpaceRole) *UserSpaceRoleDelete {
	usrd.mutation.Where(ps...)
	return usrd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (usrd *UserSpaceRoleDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, usrd.sqlExec, usrd.mutation, usrd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (usrd *UserSpaceRoleDelete) ExecX(ctx context.Context) int {
	n, err := usrd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (usrd *UserSpaceRoleDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(userspacerole.Table, sqlgraph.NewFieldSpec(userspacerole.FieldID, field.TypeString))
	if ps := usrd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, usrd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	usrd.mutation.done = true
	return affected, err
}

// UserSpaceRoleDeleteOne is the builder for deleting a single UserSpaceRole entity.
type UserSpaceRoleDeleteOne struct {
	usrd *UserSpaceRoleDelete
}

// Where appends a list predicates to the UserSpaceRoleDelete builder.
func (usrdo *UserSpaceRoleDeleteOne) Where(ps ...predicate.UserSpaceRole) *UserSpaceRoleDeleteOne {
	usrdo.usrd.mutation.Where(ps...)
	return usrdo
}

// Exec executes the deletion query.
func (usrdo *UserSpaceRoleDeleteOne) Exec(ctx context.Context) error {
	n, err := usrdo.usrd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{userspacerole.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (usrdo *UserSpaceRoleDeleteOne) ExecX(ctx context.Context) {
	if err := usrdo.Exec(ctx); err != nil {
		panic(err)
	}
}
