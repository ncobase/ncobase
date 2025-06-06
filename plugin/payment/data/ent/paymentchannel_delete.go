// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"ncobase/payment/data/ent/paymentchannel"
	"ncobase/payment/data/ent/predicate"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// PaymentChannelDelete is the builder for deleting a PaymentChannel entity.
type PaymentChannelDelete struct {
	config
	hooks    []Hook
	mutation *PaymentChannelMutation
}

// Where appends a list predicates to the PaymentChannelDelete builder.
func (pcd *PaymentChannelDelete) Where(ps ...predicate.PaymentChannel) *PaymentChannelDelete {
	pcd.mutation.Where(ps...)
	return pcd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (pcd *PaymentChannelDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, pcd.sqlExec, pcd.mutation, pcd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (pcd *PaymentChannelDelete) ExecX(ctx context.Context) int {
	n, err := pcd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (pcd *PaymentChannelDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(paymentchannel.Table, sqlgraph.NewFieldSpec(paymentchannel.FieldID, field.TypeString))
	if ps := pcd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, pcd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	pcd.mutation.done = true
	return affected, err
}

// PaymentChannelDeleteOne is the builder for deleting a single PaymentChannel entity.
type PaymentChannelDeleteOne struct {
	pcd *PaymentChannelDelete
}

// Where appends a list predicates to the PaymentChannelDelete builder.
func (pcdo *PaymentChannelDeleteOne) Where(ps ...predicate.PaymentChannel) *PaymentChannelDeleteOne {
	pcdo.pcd.mutation.Where(ps...)
	return pcdo
}

// Exec executes the deletion query.
func (pcdo *PaymentChannelDeleteOne) Exec(ctx context.Context) error {
	n, err := pcdo.pcd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{paymentchannel.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (pcdo *PaymentChannelDeleteOne) ExecX(ctx context.Context) {
	if err := pcdo.Exec(ctx); err != nil {
		panic(err)
	}
}
