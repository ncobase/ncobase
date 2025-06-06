// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"
	"ncobase/workflow/data/ent/predicate"
	"ncobase/workflow/data/ent/processdesign"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// ProcessDesignQuery is the builder for querying ProcessDesign entities.
type ProcessDesignQuery struct {
	config
	ctx        *QueryContext
	order      []processdesign.OrderOption
	inters     []Interceptor
	predicates []predicate.ProcessDesign
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ProcessDesignQuery builder.
func (pdq *ProcessDesignQuery) Where(ps ...predicate.ProcessDesign) *ProcessDesignQuery {
	pdq.predicates = append(pdq.predicates, ps...)
	return pdq
}

// Limit the number of records to be returned by this query.
func (pdq *ProcessDesignQuery) Limit(limit int) *ProcessDesignQuery {
	pdq.ctx.Limit = &limit
	return pdq
}

// Offset to start from.
func (pdq *ProcessDesignQuery) Offset(offset int) *ProcessDesignQuery {
	pdq.ctx.Offset = &offset
	return pdq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (pdq *ProcessDesignQuery) Unique(unique bool) *ProcessDesignQuery {
	pdq.ctx.Unique = &unique
	return pdq
}

// Order specifies how the records should be ordered.
func (pdq *ProcessDesignQuery) Order(o ...processdesign.OrderOption) *ProcessDesignQuery {
	pdq.order = append(pdq.order, o...)
	return pdq
}

// First returns the first ProcessDesign entity from the query.
// Returns a *NotFoundError when no ProcessDesign was found.
func (pdq *ProcessDesignQuery) First(ctx context.Context) (*ProcessDesign, error) {
	nodes, err := pdq.Limit(1).All(setContextOp(ctx, pdq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{processdesign.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (pdq *ProcessDesignQuery) FirstX(ctx context.Context) *ProcessDesign {
	node, err := pdq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first ProcessDesign ID from the query.
// Returns a *NotFoundError when no ProcessDesign ID was found.
func (pdq *ProcessDesignQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = pdq.Limit(1).IDs(setContextOp(ctx, pdq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{processdesign.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (pdq *ProcessDesignQuery) FirstIDX(ctx context.Context) string {
	id, err := pdq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single ProcessDesign entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one ProcessDesign entity is found.
// Returns a *NotFoundError when no ProcessDesign entities are found.
func (pdq *ProcessDesignQuery) Only(ctx context.Context) (*ProcessDesign, error) {
	nodes, err := pdq.Limit(2).All(setContextOp(ctx, pdq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{processdesign.Label}
	default:
		return nil, &NotSingularError{processdesign.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (pdq *ProcessDesignQuery) OnlyX(ctx context.Context) *ProcessDesign {
	node, err := pdq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only ProcessDesign ID in the query.
// Returns a *NotSingularError when more than one ProcessDesign ID is found.
// Returns a *NotFoundError when no entities are found.
func (pdq *ProcessDesignQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = pdq.Limit(2).IDs(setContextOp(ctx, pdq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{processdesign.Label}
	default:
		err = &NotSingularError{processdesign.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (pdq *ProcessDesignQuery) OnlyIDX(ctx context.Context) string {
	id, err := pdq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of ProcessDesigns.
func (pdq *ProcessDesignQuery) All(ctx context.Context) ([]*ProcessDesign, error) {
	ctx = setContextOp(ctx, pdq.ctx, ent.OpQueryAll)
	if err := pdq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*ProcessDesign, *ProcessDesignQuery]()
	return withInterceptors[[]*ProcessDesign](ctx, pdq, qr, pdq.inters)
}

// AllX is like All, but panics if an error occurs.
func (pdq *ProcessDesignQuery) AllX(ctx context.Context) []*ProcessDesign {
	nodes, err := pdq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of ProcessDesign IDs.
func (pdq *ProcessDesignQuery) IDs(ctx context.Context) (ids []string, err error) {
	if pdq.ctx.Unique == nil && pdq.path != nil {
		pdq.Unique(true)
	}
	ctx = setContextOp(ctx, pdq.ctx, ent.OpQueryIDs)
	if err = pdq.Select(processdesign.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (pdq *ProcessDesignQuery) IDsX(ctx context.Context) []string {
	ids, err := pdq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (pdq *ProcessDesignQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, pdq.ctx, ent.OpQueryCount)
	if err := pdq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, pdq, querierCount[*ProcessDesignQuery](), pdq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (pdq *ProcessDesignQuery) CountX(ctx context.Context) int {
	count, err := pdq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (pdq *ProcessDesignQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, pdq.ctx, ent.OpQueryExist)
	switch _, err := pdq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (pdq *ProcessDesignQuery) ExistX(ctx context.Context) bool {
	exist, err := pdq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ProcessDesignQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (pdq *ProcessDesignQuery) Clone() *ProcessDesignQuery {
	if pdq == nil {
		return nil
	}
	return &ProcessDesignQuery{
		config:     pdq.config,
		ctx:        pdq.ctx.Clone(),
		order:      append([]processdesign.OrderOption{}, pdq.order...),
		inters:     append([]Interceptor{}, pdq.inters...),
		predicates: append([]predicate.ProcessDesign{}, pdq.predicates...),
		// clone intermediate query.
		sql:  pdq.sql.Clone(),
		path: pdq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Version string `json:"version,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.ProcessDesign.Query().
//		GroupBy(processdesign.FieldVersion).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (pdq *ProcessDesignQuery) GroupBy(field string, fields ...string) *ProcessDesignGroupBy {
	pdq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ProcessDesignGroupBy{build: pdq}
	grbuild.flds = &pdq.ctx.Fields
	grbuild.label = processdesign.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Version string `json:"version,omitempty"`
//	}
//
//	client.ProcessDesign.Query().
//		Select(processdesign.FieldVersion).
//		Scan(ctx, &v)
func (pdq *ProcessDesignQuery) Select(fields ...string) *ProcessDesignSelect {
	pdq.ctx.Fields = append(pdq.ctx.Fields, fields...)
	sbuild := &ProcessDesignSelect{ProcessDesignQuery: pdq}
	sbuild.label = processdesign.Label
	sbuild.flds, sbuild.scan = &pdq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ProcessDesignSelect configured with the given aggregations.
func (pdq *ProcessDesignQuery) Aggregate(fns ...AggregateFunc) *ProcessDesignSelect {
	return pdq.Select().Aggregate(fns...)
}

func (pdq *ProcessDesignQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range pdq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, pdq); err != nil {
				return err
			}
		}
	}
	for _, f := range pdq.ctx.Fields {
		if !processdesign.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if pdq.path != nil {
		prev, err := pdq.path(ctx)
		if err != nil {
			return err
		}
		pdq.sql = prev
	}
	return nil
}

func (pdq *ProcessDesignQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*ProcessDesign, error) {
	var (
		nodes = []*ProcessDesign{}
		_spec = pdq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*ProcessDesign).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &ProcessDesign{config: pdq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, pdq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (pdq *ProcessDesignQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := pdq.querySpec()
	_spec.Node.Columns = pdq.ctx.Fields
	if len(pdq.ctx.Fields) > 0 {
		_spec.Unique = pdq.ctx.Unique != nil && *pdq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, pdq.driver, _spec)
}

func (pdq *ProcessDesignQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(processdesign.Table, processdesign.Columns, sqlgraph.NewFieldSpec(processdesign.FieldID, field.TypeString))
	_spec.From = pdq.sql
	if unique := pdq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if pdq.path != nil {
		_spec.Unique = true
	}
	if fields := pdq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, processdesign.FieldID)
		for i := range fields {
			if fields[i] != processdesign.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := pdq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := pdq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := pdq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := pdq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (pdq *ProcessDesignQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(pdq.driver.Dialect())
	t1 := builder.Table(processdesign.Table)
	columns := pdq.ctx.Fields
	if len(columns) == 0 {
		columns = processdesign.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if pdq.sql != nil {
		selector = pdq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if pdq.ctx.Unique != nil && *pdq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range pdq.predicates {
		p(selector)
	}
	for _, p := range pdq.order {
		p(selector)
	}
	if offset := pdq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := pdq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// ProcessDesignGroupBy is the group-by builder for ProcessDesign entities.
type ProcessDesignGroupBy struct {
	selector
	build *ProcessDesignQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (pdgb *ProcessDesignGroupBy) Aggregate(fns ...AggregateFunc) *ProcessDesignGroupBy {
	pdgb.fns = append(pdgb.fns, fns...)
	return pdgb
}

// Scan applies the selector query and scans the result into the given value.
func (pdgb *ProcessDesignGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, pdgb.build.ctx, ent.OpQueryGroupBy)
	if err := pdgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ProcessDesignQuery, *ProcessDesignGroupBy](ctx, pdgb.build, pdgb, pdgb.build.inters, v)
}

func (pdgb *ProcessDesignGroupBy) sqlScan(ctx context.Context, root *ProcessDesignQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(pdgb.fns))
	for _, fn := range pdgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*pdgb.flds)+len(pdgb.fns))
		for _, f := range *pdgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*pdgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := pdgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ProcessDesignSelect is the builder for selecting fields of ProcessDesign entities.
type ProcessDesignSelect struct {
	*ProcessDesignQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (pds *ProcessDesignSelect) Aggregate(fns ...AggregateFunc) *ProcessDesignSelect {
	pds.fns = append(pds.fns, fns...)
	return pds
}

// Scan applies the selector query and scans the result into the given value.
func (pds *ProcessDesignSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, pds.ctx, ent.OpQuerySelect)
	if err := pds.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ProcessDesignQuery, *ProcessDesignSelect](ctx, pds.ProcessDesignQuery, pds, pds.inters, v)
}

func (pds *ProcessDesignSelect) sqlScan(ctx context.Context, root *ProcessDesignQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(pds.fns))
	for _, fn := range pds.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*pds.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := pds.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
