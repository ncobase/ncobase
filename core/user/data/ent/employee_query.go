// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"
	"ncobase/user/data/ent/employee"
	"ncobase/user/data/ent/predicate"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// EmployeeQuery is the builder for querying Employee entities.
type EmployeeQuery struct {
	config
	ctx        *QueryContext
	order      []employee.OrderOption
	inters     []Interceptor
	predicates []predicate.Employee
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the EmployeeQuery builder.
func (eq *EmployeeQuery) Where(ps ...predicate.Employee) *EmployeeQuery {
	eq.predicates = append(eq.predicates, ps...)
	return eq
}

// Limit the number of records to be returned by this query.
func (eq *EmployeeQuery) Limit(limit int) *EmployeeQuery {
	eq.ctx.Limit = &limit
	return eq
}

// Offset to start from.
func (eq *EmployeeQuery) Offset(offset int) *EmployeeQuery {
	eq.ctx.Offset = &offset
	return eq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (eq *EmployeeQuery) Unique(unique bool) *EmployeeQuery {
	eq.ctx.Unique = &unique
	return eq
}

// Order specifies how the records should be ordered.
func (eq *EmployeeQuery) Order(o ...employee.OrderOption) *EmployeeQuery {
	eq.order = append(eq.order, o...)
	return eq
}

// First returns the first Employee entity from the query.
// Returns a *NotFoundError when no Employee was found.
func (eq *EmployeeQuery) First(ctx context.Context) (*Employee, error) {
	nodes, err := eq.Limit(1).All(setContextOp(ctx, eq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{employee.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (eq *EmployeeQuery) FirstX(ctx context.Context) *Employee {
	node, err := eq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Employee ID from the query.
// Returns a *NotFoundError when no Employee ID was found.
func (eq *EmployeeQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = eq.Limit(1).IDs(setContextOp(ctx, eq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{employee.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (eq *EmployeeQuery) FirstIDX(ctx context.Context) string {
	id, err := eq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Employee entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Employee entity is found.
// Returns a *NotFoundError when no Employee entities are found.
func (eq *EmployeeQuery) Only(ctx context.Context) (*Employee, error) {
	nodes, err := eq.Limit(2).All(setContextOp(ctx, eq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{employee.Label}
	default:
		return nil, &NotSingularError{employee.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (eq *EmployeeQuery) OnlyX(ctx context.Context) *Employee {
	node, err := eq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Employee ID in the query.
// Returns a *NotSingularError when more than one Employee ID is found.
// Returns a *NotFoundError when no entities are found.
func (eq *EmployeeQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = eq.Limit(2).IDs(setContextOp(ctx, eq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{employee.Label}
	default:
		err = &NotSingularError{employee.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (eq *EmployeeQuery) OnlyIDX(ctx context.Context) string {
	id, err := eq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Employees.
func (eq *EmployeeQuery) All(ctx context.Context) ([]*Employee, error) {
	ctx = setContextOp(ctx, eq.ctx, ent.OpQueryAll)
	if err := eq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Employee, *EmployeeQuery]()
	return withInterceptors[[]*Employee](ctx, eq, qr, eq.inters)
}

// AllX is like All, but panics if an error occurs.
func (eq *EmployeeQuery) AllX(ctx context.Context) []*Employee {
	nodes, err := eq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Employee IDs.
func (eq *EmployeeQuery) IDs(ctx context.Context) (ids []string, err error) {
	if eq.ctx.Unique == nil && eq.path != nil {
		eq.Unique(true)
	}
	ctx = setContextOp(ctx, eq.ctx, ent.OpQueryIDs)
	if err = eq.Select(employee.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (eq *EmployeeQuery) IDsX(ctx context.Context) []string {
	ids, err := eq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (eq *EmployeeQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, eq.ctx, ent.OpQueryCount)
	if err := eq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, eq, querierCount[*EmployeeQuery](), eq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (eq *EmployeeQuery) CountX(ctx context.Context) int {
	count, err := eq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (eq *EmployeeQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, eq.ctx, ent.OpQueryExist)
	switch _, err := eq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (eq *EmployeeQuery) ExistX(ctx context.Context) bool {
	exist, err := eq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the EmployeeQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (eq *EmployeeQuery) Clone() *EmployeeQuery {
	if eq == nil {
		return nil
	}
	return &EmployeeQuery{
		config:     eq.config,
		ctx:        eq.ctx.Clone(),
		order:      append([]employee.OrderOption{}, eq.order...),
		inters:     append([]Interceptor{}, eq.inters...),
		predicates: append([]predicate.Employee{}, eq.predicates...),
		// clone intermediate query.
		sql:  eq.sql.Clone(),
		path: eq.path,
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
//	client.Employee.Query().
//		GroupBy(employee.FieldSpaceID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (eq *EmployeeQuery) GroupBy(field string, fields ...string) *EmployeeGroupBy {
	eq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &EmployeeGroupBy{build: eq}
	grbuild.flds = &eq.ctx.Fields
	grbuild.label = employee.Label
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
//	client.Employee.Query().
//		Select(employee.FieldSpaceID).
//		Scan(ctx, &v)
func (eq *EmployeeQuery) Select(fields ...string) *EmployeeSelect {
	eq.ctx.Fields = append(eq.ctx.Fields, fields...)
	sbuild := &EmployeeSelect{EmployeeQuery: eq}
	sbuild.label = employee.Label
	sbuild.flds, sbuild.scan = &eq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a EmployeeSelect configured with the given aggregations.
func (eq *EmployeeQuery) Aggregate(fns ...AggregateFunc) *EmployeeSelect {
	return eq.Select().Aggregate(fns...)
}

func (eq *EmployeeQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range eq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, eq); err != nil {
				return err
			}
		}
	}
	for _, f := range eq.ctx.Fields {
		if !employee.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if eq.path != nil {
		prev, err := eq.path(ctx)
		if err != nil {
			return err
		}
		eq.sql = prev
	}
	return nil
}

func (eq *EmployeeQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Employee, error) {
	var (
		nodes = []*Employee{}
		_spec = eq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Employee).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Employee{config: eq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, eq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (eq *EmployeeQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := eq.querySpec()
	_spec.Node.Columns = eq.ctx.Fields
	if len(eq.ctx.Fields) > 0 {
		_spec.Unique = eq.ctx.Unique != nil && *eq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, eq.driver, _spec)
}

func (eq *EmployeeQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(employee.Table, employee.Columns, sqlgraph.NewFieldSpec(employee.FieldID, field.TypeString))
	_spec.From = eq.sql
	if unique := eq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if eq.path != nil {
		_spec.Unique = true
	}
	if fields := eq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, employee.FieldID)
		for i := range fields {
			if fields[i] != employee.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := eq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := eq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := eq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := eq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (eq *EmployeeQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(eq.driver.Dialect())
	t1 := builder.Table(employee.Table)
	columns := eq.ctx.Fields
	if len(columns) == 0 {
		columns = employee.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if eq.sql != nil {
		selector = eq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if eq.ctx.Unique != nil && *eq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range eq.predicates {
		p(selector)
	}
	for _, p := range eq.order {
		p(selector)
	}
	if offset := eq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := eq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// EmployeeGroupBy is the group-by builder for Employee entities.
type EmployeeGroupBy struct {
	selector
	build *EmployeeQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (egb *EmployeeGroupBy) Aggregate(fns ...AggregateFunc) *EmployeeGroupBy {
	egb.fns = append(egb.fns, fns...)
	return egb
}

// Scan applies the selector query and scans the result into the given value.
func (egb *EmployeeGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, egb.build.ctx, ent.OpQueryGroupBy)
	if err := egb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*EmployeeQuery, *EmployeeGroupBy](ctx, egb.build, egb, egb.build.inters, v)
}

func (egb *EmployeeGroupBy) sqlScan(ctx context.Context, root *EmployeeQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(egb.fns))
	for _, fn := range egb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*egb.flds)+len(egb.fns))
		for _, f := range *egb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*egb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := egb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// EmployeeSelect is the builder for selecting fields of Employee entities.
type EmployeeSelect struct {
	*EmployeeQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (es *EmployeeSelect) Aggregate(fns ...AggregateFunc) *EmployeeSelect {
	es.fns = append(es.fns, fns...)
	return es
}

// Scan applies the selector query and scans the result into the given value.
func (es *EmployeeSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, es.ctx, ent.OpQuerySelect)
	if err := es.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*EmployeeQuery, *EmployeeSelect](ctx, es.EmployeeQuery, es, es.inters, v)
}

func (es *EmployeeSelect) sqlScan(ctx context.Context, root *EmployeeQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(es.fns))
	for _, fn := range es.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*es.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := es.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
