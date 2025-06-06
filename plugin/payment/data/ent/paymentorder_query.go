// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"
	"ncobase/payment/data/ent/paymentlog"
	"ncobase/payment/data/ent/paymentorder"
	"ncobase/payment/data/ent/predicate"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// PaymentOrderQuery is the builder for querying PaymentOrder entities.
type PaymentOrderQuery struct {
	config
	ctx        *QueryContext
	order      []paymentorder.OrderOption
	inters     []Interceptor
	predicates []predicate.PaymentOrder
	withLogs   *PaymentLogQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the PaymentOrderQuery builder.
func (poq *PaymentOrderQuery) Where(ps ...predicate.PaymentOrder) *PaymentOrderQuery {
	poq.predicates = append(poq.predicates, ps...)
	return poq
}

// Limit the number of records to be returned by this query.
func (poq *PaymentOrderQuery) Limit(limit int) *PaymentOrderQuery {
	poq.ctx.Limit = &limit
	return poq
}

// Offset to start from.
func (poq *PaymentOrderQuery) Offset(offset int) *PaymentOrderQuery {
	poq.ctx.Offset = &offset
	return poq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (poq *PaymentOrderQuery) Unique(unique bool) *PaymentOrderQuery {
	poq.ctx.Unique = &unique
	return poq
}

// Order specifies how the records should be ordered.
func (poq *PaymentOrderQuery) Order(o ...paymentorder.OrderOption) *PaymentOrderQuery {
	poq.order = append(poq.order, o...)
	return poq
}

// QueryLogs chains the current query on the "logs" edge.
func (poq *PaymentOrderQuery) QueryLogs() *PaymentLogQuery {
	query := (&PaymentLogClient{config: poq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := poq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := poq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(paymentorder.Table, paymentorder.FieldID, selector),
			sqlgraph.To(paymentlog.Table, paymentlog.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, paymentorder.LogsTable, paymentorder.LogsColumn),
		)
		fromU = sqlgraph.SetNeighbors(poq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first PaymentOrder entity from the query.
// Returns a *NotFoundError when no PaymentOrder was found.
func (poq *PaymentOrderQuery) First(ctx context.Context) (*PaymentOrder, error) {
	nodes, err := poq.Limit(1).All(setContextOp(ctx, poq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{paymentorder.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (poq *PaymentOrderQuery) FirstX(ctx context.Context) *PaymentOrder {
	node, err := poq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first PaymentOrder ID from the query.
// Returns a *NotFoundError when no PaymentOrder ID was found.
func (poq *PaymentOrderQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = poq.Limit(1).IDs(setContextOp(ctx, poq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{paymentorder.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (poq *PaymentOrderQuery) FirstIDX(ctx context.Context) string {
	id, err := poq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single PaymentOrder entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one PaymentOrder entity is found.
// Returns a *NotFoundError when no PaymentOrder entities are found.
func (poq *PaymentOrderQuery) Only(ctx context.Context) (*PaymentOrder, error) {
	nodes, err := poq.Limit(2).All(setContextOp(ctx, poq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{paymentorder.Label}
	default:
		return nil, &NotSingularError{paymentorder.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (poq *PaymentOrderQuery) OnlyX(ctx context.Context) *PaymentOrder {
	node, err := poq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only PaymentOrder ID in the query.
// Returns a *NotSingularError when more than one PaymentOrder ID is found.
// Returns a *NotFoundError when no entities are found.
func (poq *PaymentOrderQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = poq.Limit(2).IDs(setContextOp(ctx, poq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{paymentorder.Label}
	default:
		err = &NotSingularError{paymentorder.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (poq *PaymentOrderQuery) OnlyIDX(ctx context.Context) string {
	id, err := poq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of PaymentOrders.
func (poq *PaymentOrderQuery) All(ctx context.Context) ([]*PaymentOrder, error) {
	ctx = setContextOp(ctx, poq.ctx, ent.OpQueryAll)
	if err := poq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*PaymentOrder, *PaymentOrderQuery]()
	return withInterceptors[[]*PaymentOrder](ctx, poq, qr, poq.inters)
}

// AllX is like All, but panics if an error occurs.
func (poq *PaymentOrderQuery) AllX(ctx context.Context) []*PaymentOrder {
	nodes, err := poq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of PaymentOrder IDs.
func (poq *PaymentOrderQuery) IDs(ctx context.Context) (ids []string, err error) {
	if poq.ctx.Unique == nil && poq.path != nil {
		poq.Unique(true)
	}
	ctx = setContextOp(ctx, poq.ctx, ent.OpQueryIDs)
	if err = poq.Select(paymentorder.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (poq *PaymentOrderQuery) IDsX(ctx context.Context) []string {
	ids, err := poq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (poq *PaymentOrderQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, poq.ctx, ent.OpQueryCount)
	if err := poq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, poq, querierCount[*PaymentOrderQuery](), poq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (poq *PaymentOrderQuery) CountX(ctx context.Context) int {
	count, err := poq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (poq *PaymentOrderQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, poq.ctx, ent.OpQueryExist)
	switch _, err := poq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (poq *PaymentOrderQuery) ExistX(ctx context.Context) bool {
	exist, err := poq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the PaymentOrderQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (poq *PaymentOrderQuery) Clone() *PaymentOrderQuery {
	if poq == nil {
		return nil
	}
	return &PaymentOrderQuery{
		config:     poq.config,
		ctx:        poq.ctx.Clone(),
		order:      append([]paymentorder.OrderOption{}, poq.order...),
		inters:     append([]Interceptor{}, poq.inters...),
		predicates: append([]predicate.PaymentOrder{}, poq.predicates...),
		withLogs:   poq.withLogs.Clone(),
		// clone intermediate query.
		sql:  poq.sql.Clone(),
		path: poq.path,
	}
}

// WithLogs tells the query-builder to eager-load the nodes that are connected to
// the "logs" edge. The optional arguments are used to configure the query builder of the edge.
func (poq *PaymentOrderQuery) WithLogs(opts ...func(*PaymentLogQuery)) *PaymentOrderQuery {
	query := (&PaymentLogClient{config: poq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	poq.withLogs = query
	return poq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Extras map[string]interface {} `json:"extras,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.PaymentOrder.Query().
//		GroupBy(paymentorder.FieldExtras).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (poq *PaymentOrderQuery) GroupBy(field string, fields ...string) *PaymentOrderGroupBy {
	poq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &PaymentOrderGroupBy{build: poq}
	grbuild.flds = &poq.ctx.Fields
	grbuild.label = paymentorder.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Extras map[string]interface {} `json:"extras,omitempty"`
//	}
//
//	client.PaymentOrder.Query().
//		Select(paymentorder.FieldExtras).
//		Scan(ctx, &v)
func (poq *PaymentOrderQuery) Select(fields ...string) *PaymentOrderSelect {
	poq.ctx.Fields = append(poq.ctx.Fields, fields...)
	sbuild := &PaymentOrderSelect{PaymentOrderQuery: poq}
	sbuild.label = paymentorder.Label
	sbuild.flds, sbuild.scan = &poq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a PaymentOrderSelect configured with the given aggregations.
func (poq *PaymentOrderQuery) Aggregate(fns ...AggregateFunc) *PaymentOrderSelect {
	return poq.Select().Aggregate(fns...)
}

func (poq *PaymentOrderQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range poq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, poq); err != nil {
				return err
			}
		}
	}
	for _, f := range poq.ctx.Fields {
		if !paymentorder.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if poq.path != nil {
		prev, err := poq.path(ctx)
		if err != nil {
			return err
		}
		poq.sql = prev
	}
	return nil
}

func (poq *PaymentOrderQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*PaymentOrder, error) {
	var (
		nodes       = []*PaymentOrder{}
		_spec       = poq.querySpec()
		loadedTypes = [1]bool{
			poq.withLogs != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*PaymentOrder).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &PaymentOrder{config: poq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, poq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := poq.withLogs; query != nil {
		if err := poq.loadLogs(ctx, query, nodes,
			func(n *PaymentOrder) { n.Edges.Logs = []*PaymentLog{} },
			func(n *PaymentOrder, e *PaymentLog) { n.Edges.Logs = append(n.Edges.Logs, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (poq *PaymentOrderQuery) loadLogs(ctx context.Context, query *PaymentLogQuery, nodes []*PaymentOrder, init func(*PaymentOrder), assign func(*PaymentOrder, *PaymentLog)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[string]*PaymentOrder)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(paymentlog.FieldOrderID)
	}
	query.Where(predicate.PaymentLog(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(paymentorder.LogsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.OrderID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "order_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (poq *PaymentOrderQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := poq.querySpec()
	_spec.Node.Columns = poq.ctx.Fields
	if len(poq.ctx.Fields) > 0 {
		_spec.Unique = poq.ctx.Unique != nil && *poq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, poq.driver, _spec)
}

func (poq *PaymentOrderQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(paymentorder.Table, paymentorder.Columns, sqlgraph.NewFieldSpec(paymentorder.FieldID, field.TypeString))
	_spec.From = poq.sql
	if unique := poq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if poq.path != nil {
		_spec.Unique = true
	}
	if fields := poq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, paymentorder.FieldID)
		for i := range fields {
			if fields[i] != paymentorder.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := poq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := poq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := poq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := poq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (poq *PaymentOrderQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(poq.driver.Dialect())
	t1 := builder.Table(paymentorder.Table)
	columns := poq.ctx.Fields
	if len(columns) == 0 {
		columns = paymentorder.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if poq.sql != nil {
		selector = poq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if poq.ctx.Unique != nil && *poq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range poq.predicates {
		p(selector)
	}
	for _, p := range poq.order {
		p(selector)
	}
	if offset := poq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := poq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// PaymentOrderGroupBy is the group-by builder for PaymentOrder entities.
type PaymentOrderGroupBy struct {
	selector
	build *PaymentOrderQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (pogb *PaymentOrderGroupBy) Aggregate(fns ...AggregateFunc) *PaymentOrderGroupBy {
	pogb.fns = append(pogb.fns, fns...)
	return pogb
}

// Scan applies the selector query and scans the result into the given value.
func (pogb *PaymentOrderGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, pogb.build.ctx, ent.OpQueryGroupBy)
	if err := pogb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*PaymentOrderQuery, *PaymentOrderGroupBy](ctx, pogb.build, pogb, pogb.build.inters, v)
}

func (pogb *PaymentOrderGroupBy) sqlScan(ctx context.Context, root *PaymentOrderQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(pogb.fns))
	for _, fn := range pogb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*pogb.flds)+len(pogb.fns))
		for _, f := range *pogb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*pogb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := pogb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// PaymentOrderSelect is the builder for selecting fields of PaymentOrder entities.
type PaymentOrderSelect struct {
	*PaymentOrderQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (pos *PaymentOrderSelect) Aggregate(fns ...AggregateFunc) *PaymentOrderSelect {
	pos.fns = append(pos.fns, fns...)
	return pos
}

// Scan applies the selector query and scans the result into the given value.
func (pos *PaymentOrderSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, pos.ctx, ent.OpQuerySelect)
	if err := pos.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*PaymentOrderQuery, *PaymentOrderSelect](ctx, pos.PaymentOrderQuery, pos, pos.inters, v)
}

func (pos *PaymentOrderSelect) sqlScan(ctx context.Context, root *PaymentOrderQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(pos.fns))
	for _, fn := range pos.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*pos.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := pos.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
