// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"
	"ncobase/space/data/ent/predicate"
	"ncobase/space/data/ent/spacebilling"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// SpaceBillingQuery is the builder for querying SpaceBilling entities.
type SpaceBillingQuery struct {
	config
	ctx        *QueryContext
	order      []spacebilling.OrderOption
	inters     []Interceptor
	predicates []predicate.SpaceBilling
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the SpaceBillingQuery builder.
func (sbq *SpaceBillingQuery) Where(ps ...predicate.SpaceBilling) *SpaceBillingQuery {
	sbq.predicates = append(sbq.predicates, ps...)
	return sbq
}

// Limit the number of records to be returned by this query.
func (sbq *SpaceBillingQuery) Limit(limit int) *SpaceBillingQuery {
	sbq.ctx.Limit = &limit
	return sbq
}

// Offset to start from.
func (sbq *SpaceBillingQuery) Offset(offset int) *SpaceBillingQuery {
	sbq.ctx.Offset = &offset
	return sbq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (sbq *SpaceBillingQuery) Unique(unique bool) *SpaceBillingQuery {
	sbq.ctx.Unique = &unique
	return sbq
}

// Order specifies how the records should be ordered.
func (sbq *SpaceBillingQuery) Order(o ...spacebilling.OrderOption) *SpaceBillingQuery {
	sbq.order = append(sbq.order, o...)
	return sbq
}

// First returns the first SpaceBilling entity from the query.
// Returns a *NotFoundError when no SpaceBilling was found.
func (sbq *SpaceBillingQuery) First(ctx context.Context) (*SpaceBilling, error) {
	nodes, err := sbq.Limit(1).All(setContextOp(ctx, sbq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{spacebilling.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (sbq *SpaceBillingQuery) FirstX(ctx context.Context) *SpaceBilling {
	node, err := sbq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first SpaceBilling ID from the query.
// Returns a *NotFoundError when no SpaceBilling ID was found.
func (sbq *SpaceBillingQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = sbq.Limit(1).IDs(setContextOp(ctx, sbq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{spacebilling.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (sbq *SpaceBillingQuery) FirstIDX(ctx context.Context) string {
	id, err := sbq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single SpaceBilling entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one SpaceBilling entity is found.
// Returns a *NotFoundError when no SpaceBilling entities are found.
func (sbq *SpaceBillingQuery) Only(ctx context.Context) (*SpaceBilling, error) {
	nodes, err := sbq.Limit(2).All(setContextOp(ctx, sbq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{spacebilling.Label}
	default:
		return nil, &NotSingularError{spacebilling.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (sbq *SpaceBillingQuery) OnlyX(ctx context.Context) *SpaceBilling {
	node, err := sbq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only SpaceBilling ID in the query.
// Returns a *NotSingularError when more than one SpaceBilling ID is found.
// Returns a *NotFoundError when no entities are found.
func (sbq *SpaceBillingQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = sbq.Limit(2).IDs(setContextOp(ctx, sbq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{spacebilling.Label}
	default:
		err = &NotSingularError{spacebilling.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (sbq *SpaceBillingQuery) OnlyIDX(ctx context.Context) string {
	id, err := sbq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of SpaceBillings.
func (sbq *SpaceBillingQuery) All(ctx context.Context) ([]*SpaceBilling, error) {
	ctx = setContextOp(ctx, sbq.ctx, ent.OpQueryAll)
	if err := sbq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*SpaceBilling, *SpaceBillingQuery]()
	return withInterceptors[[]*SpaceBilling](ctx, sbq, qr, sbq.inters)
}

// AllX is like All, but panics if an error occurs.
func (sbq *SpaceBillingQuery) AllX(ctx context.Context) []*SpaceBilling {
	nodes, err := sbq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of SpaceBilling IDs.
func (sbq *SpaceBillingQuery) IDs(ctx context.Context) (ids []string, err error) {
	if sbq.ctx.Unique == nil && sbq.path != nil {
		sbq.Unique(true)
	}
	ctx = setContextOp(ctx, sbq.ctx, ent.OpQueryIDs)
	if err = sbq.Select(spacebilling.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (sbq *SpaceBillingQuery) IDsX(ctx context.Context) []string {
	ids, err := sbq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (sbq *SpaceBillingQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, sbq.ctx, ent.OpQueryCount)
	if err := sbq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, sbq, querierCount[*SpaceBillingQuery](), sbq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (sbq *SpaceBillingQuery) CountX(ctx context.Context) int {
	count, err := sbq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (sbq *SpaceBillingQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, sbq.ctx, ent.OpQueryExist)
	switch _, err := sbq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (sbq *SpaceBillingQuery) ExistX(ctx context.Context) bool {
	exist, err := sbq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the SpaceBillingQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (sbq *SpaceBillingQuery) Clone() *SpaceBillingQuery {
	if sbq == nil {
		return nil
	}
	return &SpaceBillingQuery{
		config:     sbq.config,
		ctx:        sbq.ctx.Clone(),
		order:      append([]spacebilling.OrderOption{}, sbq.order...),
		inters:     append([]Interceptor{}, sbq.inters...),
		predicates: append([]predicate.SpaceBilling{}, sbq.predicates...),
		// clone intermediate query.
		sql:  sbq.sql.Clone(),
		path: sbq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		SpaceID string `json:"space_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.SpaceBilling.Query().
//		GroupBy(spacebilling.FieldSpaceID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (sbq *SpaceBillingQuery) GroupBy(field string, fields ...string) *SpaceBillingGroupBy {
	sbq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &SpaceBillingGroupBy{build: sbq}
	grbuild.flds = &sbq.ctx.Fields
	grbuild.label = spacebilling.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		SpaceID string `json:"space_id,omitempty"`
//	}
//
//	client.SpaceBilling.Query().
//		Select(spacebilling.FieldSpaceID).
//		Scan(ctx, &v)
func (sbq *SpaceBillingQuery) Select(fields ...string) *SpaceBillingSelect {
	sbq.ctx.Fields = append(sbq.ctx.Fields, fields...)
	sbuild := &SpaceBillingSelect{SpaceBillingQuery: sbq}
	sbuild.label = spacebilling.Label
	sbuild.flds, sbuild.scan = &sbq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a SpaceBillingSelect configured with the given aggregations.
func (sbq *SpaceBillingQuery) Aggregate(fns ...AggregateFunc) *SpaceBillingSelect {
	return sbq.Select().Aggregate(fns...)
}

func (sbq *SpaceBillingQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range sbq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, sbq); err != nil {
				return err
			}
		}
	}
	for _, f := range sbq.ctx.Fields {
		if !spacebilling.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if sbq.path != nil {
		prev, err := sbq.path(ctx)
		if err != nil {
			return err
		}
		sbq.sql = prev
	}
	return nil
}

func (sbq *SpaceBillingQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*SpaceBilling, error) {
	var (
		nodes = []*SpaceBilling{}
		_spec = sbq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*SpaceBilling).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &SpaceBilling{config: sbq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, sbq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (sbq *SpaceBillingQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := sbq.querySpec()
	_spec.Node.Columns = sbq.ctx.Fields
	if len(sbq.ctx.Fields) > 0 {
		_spec.Unique = sbq.ctx.Unique != nil && *sbq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, sbq.driver, _spec)
}

func (sbq *SpaceBillingQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(spacebilling.Table, spacebilling.Columns, sqlgraph.NewFieldSpec(spacebilling.FieldID, field.TypeString))
	_spec.From = sbq.sql
	if unique := sbq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if sbq.path != nil {
		_spec.Unique = true
	}
	if fields := sbq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, spacebilling.FieldID)
		for i := range fields {
			if fields[i] != spacebilling.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := sbq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := sbq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := sbq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := sbq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (sbq *SpaceBillingQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(sbq.driver.Dialect())
	t1 := builder.Table(spacebilling.Table)
	columns := sbq.ctx.Fields
	if len(columns) == 0 {
		columns = spacebilling.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if sbq.sql != nil {
		selector = sbq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if sbq.ctx.Unique != nil && *sbq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range sbq.predicates {
		p(selector)
	}
	for _, p := range sbq.order {
		p(selector)
	}
	if offset := sbq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := sbq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// SpaceBillingGroupBy is the group-by builder for SpaceBilling entities.
type SpaceBillingGroupBy struct {
	selector
	build *SpaceBillingQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (sbgb *SpaceBillingGroupBy) Aggregate(fns ...AggregateFunc) *SpaceBillingGroupBy {
	sbgb.fns = append(sbgb.fns, fns...)
	return sbgb
}

// Scan applies the selector query and scans the result into the given value.
func (sbgb *SpaceBillingGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sbgb.build.ctx, ent.OpQueryGroupBy)
	if err := sbgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SpaceBillingQuery, *SpaceBillingGroupBy](ctx, sbgb.build, sbgb, sbgb.build.inters, v)
}

func (sbgb *SpaceBillingGroupBy) sqlScan(ctx context.Context, root *SpaceBillingQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(sbgb.fns))
	for _, fn := range sbgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*sbgb.flds)+len(sbgb.fns))
		for _, f := range *sbgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*sbgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sbgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// SpaceBillingSelect is the builder for selecting fields of SpaceBilling entities.
type SpaceBillingSelect struct {
	*SpaceBillingQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (sbs *SpaceBillingSelect) Aggregate(fns ...AggregateFunc) *SpaceBillingSelect {
	sbs.fns = append(sbs.fns, fns...)
	return sbs
}

// Scan applies the selector query and scans the result into the given value.
func (sbs *SpaceBillingSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sbs.ctx, ent.OpQuerySelect)
	if err := sbs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SpaceBillingQuery, *SpaceBillingSelect](ctx, sbs.SpaceBillingQuery, sbs, sbs.inters, v)
}

func (sbs *SpaceBillingSelect) sqlScan(ctx context.Context, root *SpaceBillingQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(sbs.fns))
	for _, fn := range sbs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*sbs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sbs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
