// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"
	"ncobase/realtime/data/ent/predicate"
	"ncobase/realtime/data/ent/rtchannel"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// RTChannelQuery is the builder for querying RTChannel entities.
type RTChannelQuery struct {
	config
	ctx        *QueryContext
	order      []rtchannel.OrderOption
	inters     []Interceptor
	predicates []predicate.RTChannel
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the RTChannelQuery builder.
func (rcq *RTChannelQuery) Where(ps ...predicate.RTChannel) *RTChannelQuery {
	rcq.predicates = append(rcq.predicates, ps...)
	return rcq
}

// Limit the number of records to be returned by this query.
func (rcq *RTChannelQuery) Limit(limit int) *RTChannelQuery {
	rcq.ctx.Limit = &limit
	return rcq
}

// Offset to start from.
func (rcq *RTChannelQuery) Offset(offset int) *RTChannelQuery {
	rcq.ctx.Offset = &offset
	return rcq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (rcq *RTChannelQuery) Unique(unique bool) *RTChannelQuery {
	rcq.ctx.Unique = &unique
	return rcq
}

// Order specifies how the records should be ordered.
func (rcq *RTChannelQuery) Order(o ...rtchannel.OrderOption) *RTChannelQuery {
	rcq.order = append(rcq.order, o...)
	return rcq
}

// First returns the first RTChannel entity from the query.
// Returns a *NotFoundError when no RTChannel was found.
func (rcq *RTChannelQuery) First(ctx context.Context) (*RTChannel, error) {
	nodes, err := rcq.Limit(1).All(setContextOp(ctx, rcq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{rtchannel.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (rcq *RTChannelQuery) FirstX(ctx context.Context) *RTChannel {
	node, err := rcq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first RTChannel ID from the query.
// Returns a *NotFoundError when no RTChannel ID was found.
func (rcq *RTChannelQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = rcq.Limit(1).IDs(setContextOp(ctx, rcq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{rtchannel.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (rcq *RTChannelQuery) FirstIDX(ctx context.Context) string {
	id, err := rcq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single RTChannel entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one RTChannel entity is found.
// Returns a *NotFoundError when no RTChannel entities are found.
func (rcq *RTChannelQuery) Only(ctx context.Context) (*RTChannel, error) {
	nodes, err := rcq.Limit(2).All(setContextOp(ctx, rcq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{rtchannel.Label}
	default:
		return nil, &NotSingularError{rtchannel.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (rcq *RTChannelQuery) OnlyX(ctx context.Context) *RTChannel {
	node, err := rcq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only RTChannel ID in the query.
// Returns a *NotSingularError when more than one RTChannel ID is found.
// Returns a *NotFoundError when no entities are found.
func (rcq *RTChannelQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = rcq.Limit(2).IDs(setContextOp(ctx, rcq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{rtchannel.Label}
	default:
		err = &NotSingularError{rtchannel.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (rcq *RTChannelQuery) OnlyIDX(ctx context.Context) string {
	id, err := rcq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of RTChannels.
func (rcq *RTChannelQuery) All(ctx context.Context) ([]*RTChannel, error) {
	ctx = setContextOp(ctx, rcq.ctx, ent.OpQueryAll)
	if err := rcq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*RTChannel, *RTChannelQuery]()
	return withInterceptors[[]*RTChannel](ctx, rcq, qr, rcq.inters)
}

// AllX is like All, but panics if an error occurs.
func (rcq *RTChannelQuery) AllX(ctx context.Context) []*RTChannel {
	nodes, err := rcq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of RTChannel IDs.
func (rcq *RTChannelQuery) IDs(ctx context.Context) (ids []string, err error) {
	if rcq.ctx.Unique == nil && rcq.path != nil {
		rcq.Unique(true)
	}
	ctx = setContextOp(ctx, rcq.ctx, ent.OpQueryIDs)
	if err = rcq.Select(rtchannel.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (rcq *RTChannelQuery) IDsX(ctx context.Context) []string {
	ids, err := rcq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (rcq *RTChannelQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, rcq.ctx, ent.OpQueryCount)
	if err := rcq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, rcq, querierCount[*RTChannelQuery](), rcq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (rcq *RTChannelQuery) CountX(ctx context.Context) int {
	count, err := rcq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (rcq *RTChannelQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, rcq.ctx, ent.OpQueryExist)
	switch _, err := rcq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (rcq *RTChannelQuery) ExistX(ctx context.Context) bool {
	exist, err := rcq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the RTChannelQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (rcq *RTChannelQuery) Clone() *RTChannelQuery {
	if rcq == nil {
		return nil
	}
	return &RTChannelQuery{
		config:     rcq.config,
		ctx:        rcq.ctx.Clone(),
		order:      append([]rtchannel.OrderOption{}, rcq.order...),
		inters:     append([]Interceptor{}, rcq.inters...),
		predicates: append([]predicate.RTChannel{}, rcq.predicates...),
		// clone intermediate query.
		sql:  rcq.sql.Clone(),
		path: rcq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Name string `json:"name,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.RTChannel.Query().
//		GroupBy(rtchannel.FieldName).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (rcq *RTChannelQuery) GroupBy(field string, fields ...string) *RTChannelGroupBy {
	rcq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &RTChannelGroupBy{build: rcq}
	grbuild.flds = &rcq.ctx.Fields
	grbuild.label = rtchannel.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Name string `json:"name,omitempty"`
//	}
//
//	client.RTChannel.Query().
//		Select(rtchannel.FieldName).
//		Scan(ctx, &v)
func (rcq *RTChannelQuery) Select(fields ...string) *RTChannelSelect {
	rcq.ctx.Fields = append(rcq.ctx.Fields, fields...)
	sbuild := &RTChannelSelect{RTChannelQuery: rcq}
	sbuild.label = rtchannel.Label
	sbuild.flds, sbuild.scan = &rcq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a RTChannelSelect configured with the given aggregations.
func (rcq *RTChannelQuery) Aggregate(fns ...AggregateFunc) *RTChannelSelect {
	return rcq.Select().Aggregate(fns...)
}

func (rcq *RTChannelQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range rcq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, rcq); err != nil {
				return err
			}
		}
	}
	for _, f := range rcq.ctx.Fields {
		if !rtchannel.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if rcq.path != nil {
		prev, err := rcq.path(ctx)
		if err != nil {
			return err
		}
		rcq.sql = prev
	}
	return nil
}

func (rcq *RTChannelQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*RTChannel, error) {
	var (
		nodes = []*RTChannel{}
		_spec = rcq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*RTChannel).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &RTChannel{config: rcq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, rcq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (rcq *RTChannelQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := rcq.querySpec()
	_spec.Node.Columns = rcq.ctx.Fields
	if len(rcq.ctx.Fields) > 0 {
		_spec.Unique = rcq.ctx.Unique != nil && *rcq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, rcq.driver, _spec)
}

func (rcq *RTChannelQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(rtchannel.Table, rtchannel.Columns, sqlgraph.NewFieldSpec(rtchannel.FieldID, field.TypeString))
	_spec.From = rcq.sql
	if unique := rcq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if rcq.path != nil {
		_spec.Unique = true
	}
	if fields := rcq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, rtchannel.FieldID)
		for i := range fields {
			if fields[i] != rtchannel.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := rcq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := rcq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := rcq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := rcq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (rcq *RTChannelQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(rcq.driver.Dialect())
	t1 := builder.Table(rtchannel.Table)
	columns := rcq.ctx.Fields
	if len(columns) == 0 {
		columns = rtchannel.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if rcq.sql != nil {
		selector = rcq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if rcq.ctx.Unique != nil && *rcq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range rcq.predicates {
		p(selector)
	}
	for _, p := range rcq.order {
		p(selector)
	}
	if offset := rcq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := rcq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// RTChannelGroupBy is the group-by builder for RTChannel entities.
type RTChannelGroupBy struct {
	selector
	build *RTChannelQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (rcgb *RTChannelGroupBy) Aggregate(fns ...AggregateFunc) *RTChannelGroupBy {
	rcgb.fns = append(rcgb.fns, fns...)
	return rcgb
}

// Scan applies the selector query and scans the result into the given value.
func (rcgb *RTChannelGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rcgb.build.ctx, ent.OpQueryGroupBy)
	if err := rcgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RTChannelQuery, *RTChannelGroupBy](ctx, rcgb.build, rcgb, rcgb.build.inters, v)
}

func (rcgb *RTChannelGroupBy) sqlScan(ctx context.Context, root *RTChannelQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(rcgb.fns))
	for _, fn := range rcgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*rcgb.flds)+len(rcgb.fns))
		for _, f := range *rcgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*rcgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rcgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// RTChannelSelect is the builder for selecting fields of RTChannel entities.
type RTChannelSelect struct {
	*RTChannelQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (rcs *RTChannelSelect) Aggregate(fns ...AggregateFunc) *RTChannelSelect {
	rcs.fns = append(rcs.fns, fns...)
	return rcs
}

// Scan applies the selector query and scans the result into the given value.
func (rcs *RTChannelSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rcs.ctx, ent.OpQuerySelect)
	if err := rcs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RTChannelQuery, *RTChannelSelect](ctx, rcs.RTChannelQuery, rcs, rcs.inters, v)
}

func (rcs *RTChannelSelect) sqlScan(ctx context.Context, root *RTChannelQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(rcs.fns))
	for _, fn := range rcs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*rcs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rcs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
