// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"
	"ncobase/tenant/data/ent/predicate"
	"ncobase/tenant/data/ent/tenantbilling"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// TenantBillingQuery is the builder for querying TenantBilling entities.
type TenantBillingQuery struct {
	config
	ctx        *QueryContext
	order      []tenantbilling.OrderOption
	inters     []Interceptor
	predicates []predicate.TenantBilling
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the TenantBillingQuery builder.
func (tbq *TenantBillingQuery) Where(ps ...predicate.TenantBilling) *TenantBillingQuery {
	tbq.predicates = append(tbq.predicates, ps...)
	return tbq
}

// Limit the number of records to be returned by this query.
func (tbq *TenantBillingQuery) Limit(limit int) *TenantBillingQuery {
	tbq.ctx.Limit = &limit
	return tbq
}

// Offset to start from.
func (tbq *TenantBillingQuery) Offset(offset int) *TenantBillingQuery {
	tbq.ctx.Offset = &offset
	return tbq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (tbq *TenantBillingQuery) Unique(unique bool) *TenantBillingQuery {
	tbq.ctx.Unique = &unique
	return tbq
}

// Order specifies how the records should be ordered.
func (tbq *TenantBillingQuery) Order(o ...tenantbilling.OrderOption) *TenantBillingQuery {
	tbq.order = append(tbq.order, o...)
	return tbq
}

// First returns the first TenantBilling entity from the query.
// Returns a *NotFoundError when no TenantBilling was found.
func (tbq *TenantBillingQuery) First(ctx context.Context) (*TenantBilling, error) {
	nodes, err := tbq.Limit(1).All(setContextOp(ctx, tbq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{tenantbilling.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (tbq *TenantBillingQuery) FirstX(ctx context.Context) *TenantBilling {
	node, err := tbq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first TenantBilling ID from the query.
// Returns a *NotFoundError when no TenantBilling ID was found.
func (tbq *TenantBillingQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = tbq.Limit(1).IDs(setContextOp(ctx, tbq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{tenantbilling.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (tbq *TenantBillingQuery) FirstIDX(ctx context.Context) string {
	id, err := tbq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single TenantBilling entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one TenantBilling entity is found.
// Returns a *NotFoundError when no TenantBilling entities are found.
func (tbq *TenantBillingQuery) Only(ctx context.Context) (*TenantBilling, error) {
	nodes, err := tbq.Limit(2).All(setContextOp(ctx, tbq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{tenantbilling.Label}
	default:
		return nil, &NotSingularError{tenantbilling.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (tbq *TenantBillingQuery) OnlyX(ctx context.Context) *TenantBilling {
	node, err := tbq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only TenantBilling ID in the query.
// Returns a *NotSingularError when more than one TenantBilling ID is found.
// Returns a *NotFoundError when no entities are found.
func (tbq *TenantBillingQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = tbq.Limit(2).IDs(setContextOp(ctx, tbq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{tenantbilling.Label}
	default:
		err = &NotSingularError{tenantbilling.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (tbq *TenantBillingQuery) OnlyIDX(ctx context.Context) string {
	id, err := tbq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of TenantBillings.
func (tbq *TenantBillingQuery) All(ctx context.Context) ([]*TenantBilling, error) {
	ctx = setContextOp(ctx, tbq.ctx, ent.OpQueryAll)
	if err := tbq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*TenantBilling, *TenantBillingQuery]()
	return withInterceptors[[]*TenantBilling](ctx, tbq, qr, tbq.inters)
}

// AllX is like All, but panics if an error occurs.
func (tbq *TenantBillingQuery) AllX(ctx context.Context) []*TenantBilling {
	nodes, err := tbq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of TenantBilling IDs.
func (tbq *TenantBillingQuery) IDs(ctx context.Context) (ids []string, err error) {
	if tbq.ctx.Unique == nil && tbq.path != nil {
		tbq.Unique(true)
	}
	ctx = setContextOp(ctx, tbq.ctx, ent.OpQueryIDs)
	if err = tbq.Select(tenantbilling.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (tbq *TenantBillingQuery) IDsX(ctx context.Context) []string {
	ids, err := tbq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (tbq *TenantBillingQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, tbq.ctx, ent.OpQueryCount)
	if err := tbq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, tbq, querierCount[*TenantBillingQuery](), tbq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (tbq *TenantBillingQuery) CountX(ctx context.Context) int {
	count, err := tbq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (tbq *TenantBillingQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, tbq.ctx, ent.OpQueryExist)
	switch _, err := tbq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (tbq *TenantBillingQuery) ExistX(ctx context.Context) bool {
	exist, err := tbq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the TenantBillingQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (tbq *TenantBillingQuery) Clone() *TenantBillingQuery {
	if tbq == nil {
		return nil
	}
	return &TenantBillingQuery{
		config:     tbq.config,
		ctx:        tbq.ctx.Clone(),
		order:      append([]tenantbilling.OrderOption{}, tbq.order...),
		inters:     append([]Interceptor{}, tbq.inters...),
		predicates: append([]predicate.TenantBilling{}, tbq.predicates...),
		// clone intermediate query.
		sql:  tbq.sql.Clone(),
		path: tbq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		TenantID string `json:"tenant_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.TenantBilling.Query().
//		GroupBy(tenantbilling.FieldTenantID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (tbq *TenantBillingQuery) GroupBy(field string, fields ...string) *TenantBillingGroupBy {
	tbq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &TenantBillingGroupBy{build: tbq}
	grbuild.flds = &tbq.ctx.Fields
	grbuild.label = tenantbilling.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		TenantID string `json:"tenant_id,omitempty"`
//	}
//
//	client.TenantBilling.Query().
//		Select(tenantbilling.FieldTenantID).
//		Scan(ctx, &v)
func (tbq *TenantBillingQuery) Select(fields ...string) *TenantBillingSelect {
	tbq.ctx.Fields = append(tbq.ctx.Fields, fields...)
	sbuild := &TenantBillingSelect{TenantBillingQuery: tbq}
	sbuild.label = tenantbilling.Label
	sbuild.flds, sbuild.scan = &tbq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a TenantBillingSelect configured with the given aggregations.
func (tbq *TenantBillingQuery) Aggregate(fns ...AggregateFunc) *TenantBillingSelect {
	return tbq.Select().Aggregate(fns...)
}

func (tbq *TenantBillingQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range tbq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, tbq); err != nil {
				return err
			}
		}
	}
	for _, f := range tbq.ctx.Fields {
		if !tenantbilling.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if tbq.path != nil {
		prev, err := tbq.path(ctx)
		if err != nil {
			return err
		}
		tbq.sql = prev
	}
	return nil
}

func (tbq *TenantBillingQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*TenantBilling, error) {
	var (
		nodes = []*TenantBilling{}
		_spec = tbq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*TenantBilling).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &TenantBilling{config: tbq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, tbq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (tbq *TenantBillingQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := tbq.querySpec()
	_spec.Node.Columns = tbq.ctx.Fields
	if len(tbq.ctx.Fields) > 0 {
		_spec.Unique = tbq.ctx.Unique != nil && *tbq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, tbq.driver, _spec)
}

func (tbq *TenantBillingQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(tenantbilling.Table, tenantbilling.Columns, sqlgraph.NewFieldSpec(tenantbilling.FieldID, field.TypeString))
	_spec.From = tbq.sql
	if unique := tbq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if tbq.path != nil {
		_spec.Unique = true
	}
	if fields := tbq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, tenantbilling.FieldID)
		for i := range fields {
			if fields[i] != tenantbilling.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := tbq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := tbq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := tbq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := tbq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (tbq *TenantBillingQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(tbq.driver.Dialect())
	t1 := builder.Table(tenantbilling.Table)
	columns := tbq.ctx.Fields
	if len(columns) == 0 {
		columns = tenantbilling.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if tbq.sql != nil {
		selector = tbq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if tbq.ctx.Unique != nil && *tbq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range tbq.predicates {
		p(selector)
	}
	for _, p := range tbq.order {
		p(selector)
	}
	if offset := tbq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := tbq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// TenantBillingGroupBy is the group-by builder for TenantBilling entities.
type TenantBillingGroupBy struct {
	selector
	build *TenantBillingQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (tbgb *TenantBillingGroupBy) Aggregate(fns ...AggregateFunc) *TenantBillingGroupBy {
	tbgb.fns = append(tbgb.fns, fns...)
	return tbgb
}

// Scan applies the selector query and scans the result into the given value.
func (tbgb *TenantBillingGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tbgb.build.ctx, ent.OpQueryGroupBy)
	if err := tbgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TenantBillingQuery, *TenantBillingGroupBy](ctx, tbgb.build, tbgb, tbgb.build.inters, v)
}

func (tbgb *TenantBillingGroupBy) sqlScan(ctx context.Context, root *TenantBillingQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(tbgb.fns))
	for _, fn := range tbgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*tbgb.flds)+len(tbgb.fns))
		for _, f := range *tbgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*tbgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tbgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// TenantBillingSelect is the builder for selecting fields of TenantBilling entities.
type TenantBillingSelect struct {
	*TenantBillingQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (tbs *TenantBillingSelect) Aggregate(fns ...AggregateFunc) *TenantBillingSelect {
	tbs.fns = append(tbs.fns, fns...)
	return tbs
}

// Scan applies the selector query and scans the result into the given value.
func (tbs *TenantBillingSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tbs.ctx, ent.OpQuerySelect)
	if err := tbs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TenantBillingQuery, *TenantBillingSelect](ctx, tbs.TenantBillingQuery, tbs, tbs.inters, v)
}

func (tbs *TenantBillingSelect) sqlScan(ctx context.Context, root *TenantBillingQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(tbs.fns))
	for _, fn := range tbs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*tbs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tbs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
