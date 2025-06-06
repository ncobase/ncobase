// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"ncobase/content/data/ent/predicate"
	"ncobase/content/data/ent/taxonomyrelation"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// TaxonomyRelationDelete is the builder for deleting a TaxonomyRelation entity.
type TaxonomyRelationDelete struct {
	config
	hooks    []Hook
	mutation *TaxonomyRelationMutation
}

// Where appends a list predicates to the TaxonomyRelationDelete builder.
func (trd *TaxonomyRelationDelete) Where(ps ...predicate.TaxonomyRelation) *TaxonomyRelationDelete {
	trd.mutation.Where(ps...)
	return trd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (trd *TaxonomyRelationDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, trd.sqlExec, trd.mutation, trd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (trd *TaxonomyRelationDelete) ExecX(ctx context.Context) int {
	n, err := trd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (trd *TaxonomyRelationDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(taxonomyrelation.Table, sqlgraph.NewFieldSpec(taxonomyrelation.FieldID, field.TypeString))
	if ps := trd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, trd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	trd.mutation.done = true
	return affected, err
}

// TaxonomyRelationDeleteOne is the builder for deleting a single TaxonomyRelation entity.
type TaxonomyRelationDeleteOne struct {
	trd *TaxonomyRelationDelete
}

// Where appends a list predicates to the TaxonomyRelationDelete builder.
func (trdo *TaxonomyRelationDeleteOne) Where(ps ...predicate.TaxonomyRelation) *TaxonomyRelationDeleteOne {
	trdo.trd.mutation.Where(ps...)
	return trdo
}

// Exec executes the deletion query.
func (trdo *TaxonomyRelationDeleteOne) Exec(ctx context.Context) error {
	n, err := trdo.trd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{taxonomyrelation.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (trdo *TaxonomyRelationDeleteOne) ExecX(ctx context.Context) {
	if err := trdo.Exec(ctx); err != nil {
		panic(err)
	}
}
