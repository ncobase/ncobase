// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"ncobase/space/data/ent/migrate"

	"ncobase/space/data/ent/group"
	"ncobase/space/data/ent/grouprole"
	"ncobase/space/data/ent/usergroup"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// Group is the client for interacting with the Group builders.
	Group *GroupClient
	// GroupRole is the client for interacting with the GroupRole builders.
	GroupRole *GroupRoleClient
	// UserGroup is the client for interacting with the UserGroup builders.
	UserGroup *UserGroupClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.Group = NewGroupClient(c.config)
	c.GroupRole = NewGroupRoleClient(c.config)
	c.UserGroup = NewUserGroupClient(c.config)
}

type (
	// config is the configuration for the client and its builder.
	config struct {
		// driver used for executing database requests.
		driver dialect.Driver
		// debug enable a debug logging.
		debug bool
		// log used for logging on debug mode.
		log func(...any)
		// hooks to execute on mutations.
		hooks *hooks
		// interceptors to execute on queries.
		inters *inters
	}
	// Option function to configure the client.
	Option func(*config)
)

// newConfig creates a new config for the client.
func newConfig(opts ...Option) config {
	cfg := config{log: log.Println, hooks: &hooks{}, inters: &inters{}}
	cfg.options(opts...)
	return cfg
}

// options applies the options on the config object.
func (c *config) options(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.debug {
		c.driver = dialect.Debug(c.driver, c.log)
	}
}

// Debug enables debug logging on the ent.Driver.
func Debug() Option {
	return func(c *config) {
		c.debug = true
	}
}

// Log sets the logging function for debug mode.
func Log(fn func(...any)) Option {
	return func(c *config) {
		c.log = fn
	}
}

// Driver configures the client driver.
func Driver(driver dialect.Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// Open opens a database/sql.DB specified by the driver name and
// the data source name, and returns a new client attached to it.
// Optional parameters can be added for configuring the client.
func Open(driverName, dataSourceName string, options ...Option) (*Client, error) {
	switch driverName {
	case dialect.MySQL, dialect.Postgres, dialect.SQLite:
		drv, err := sql.Open(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return NewClient(append(options, Driver(drv))...), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driverName)
	}
}

// ErrTxStarted is returned when trying to start a new transaction from a transactional client.
var ErrTxStarted = errors.New("ent: cannot start a transaction within a transaction")

// Tx returns a new transactional client. The provided context
// is used until the transaction is committed or rolled back.
func (c *Client) Tx(ctx context.Context) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, ErrTxStarted
	}
	tx, err := newTx(ctx, c.driver)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = tx
	return &Tx{
		ctx:       ctx,
		config:    cfg,
		Group:     NewGroupClient(cfg),
		GroupRole: NewGroupRoleClient(cfg),
		UserGroup: NewUserGroupClient(cfg),
	}, nil
}

// BeginTx returns a transactional client with specified options.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	if _, ok := c.driver.(*txDriver); ok {
		return nil, errors.New("ent: cannot start a transaction within a transaction")
	}
	tx, err := c.driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	}).BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ent: starting a transaction: %w", err)
	}
	cfg := c.config
	cfg.driver = &txDriver{tx: tx, drv: c.driver}
	return &Tx{
		ctx:       ctx,
		config:    cfg,
		Group:     NewGroupClient(cfg),
		GroupRole: NewGroupRoleClient(cfg),
		UserGroup: NewUserGroupClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		Group.
//		Query().
//		Count(ctx)
func (c *Client) Debug() *Client {
	if c.debug {
		return c
	}
	cfg := c.config
	cfg.driver = dialect.Debug(c.driver, c.log)
	client := &Client{config: cfg}
	client.init()
	return client
}

// Close closes the database connection and prevents new queries from starting.
func (c *Client) Close() error {
	return c.driver.Close()
}

// Use adds the mutation hooks to all the entity clients.
// In order to add hooks to a specific client, call: `client.Node.Use(...)`.
func (c *Client) Use(hooks ...Hook) {
	c.Group.Use(hooks...)
	c.GroupRole.Use(hooks...)
	c.UserGroup.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.Group.Intercept(interceptors...)
	c.GroupRole.Intercept(interceptors...)
	c.UserGroup.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *GroupMutation:
		return c.Group.mutate(ctx, m)
	case *GroupRoleMutation:
		return c.GroupRole.mutate(ctx, m)
	case *UserGroupMutation:
		return c.UserGroup.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// GroupClient is a client for the Group schema.
type GroupClient struct {
	config
}

// NewGroupClient returns a client for the Group from the given config.
func NewGroupClient(c config) *GroupClient {
	return &GroupClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `group.Hooks(f(g(h())))`.
func (c *GroupClient) Use(hooks ...Hook) {
	c.hooks.Group = append(c.hooks.Group, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `group.Intercept(f(g(h())))`.
func (c *GroupClient) Intercept(interceptors ...Interceptor) {
	c.inters.Group = append(c.inters.Group, interceptors...)
}

// Create returns a builder for creating a Group entity.
func (c *GroupClient) Create() *GroupCreate {
	mutation := newGroupMutation(c.config, OpCreate)
	return &GroupCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Group entities.
func (c *GroupClient) CreateBulk(builders ...*GroupCreate) *GroupCreateBulk {
	return &GroupCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *GroupClient) MapCreateBulk(slice any, setFunc func(*GroupCreate, int)) *GroupCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &GroupCreateBulk{err: fmt.Errorf("calling to GroupClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*GroupCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &GroupCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Group.
func (c *GroupClient) Update() *GroupUpdate {
	mutation := newGroupMutation(c.config, OpUpdate)
	return &GroupUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *GroupClient) UpdateOne(gr *Group) *GroupUpdateOne {
	mutation := newGroupMutation(c.config, OpUpdateOne, withGroup(gr))
	return &GroupUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *GroupClient) UpdateOneID(id string) *GroupUpdateOne {
	mutation := newGroupMutation(c.config, OpUpdateOne, withGroupID(id))
	return &GroupUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Group.
func (c *GroupClient) Delete() *GroupDelete {
	mutation := newGroupMutation(c.config, OpDelete)
	return &GroupDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *GroupClient) DeleteOne(gr *Group) *GroupDeleteOne {
	return c.DeleteOneID(gr.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *GroupClient) DeleteOneID(id string) *GroupDeleteOne {
	builder := c.Delete().Where(group.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &GroupDeleteOne{builder}
}

// Query returns a query builder for Group.
func (c *GroupClient) Query() *GroupQuery {
	return &GroupQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeGroup},
		inters: c.Interceptors(),
	}
}

// Get returns a Group entity by its id.
func (c *GroupClient) Get(ctx context.Context, id string) (*Group, error) {
	return c.Query().Where(group.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *GroupClient) GetX(ctx context.Context, id string) *Group {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *GroupClient) Hooks() []Hook {
	return c.hooks.Group
}

// Interceptors returns the client interceptors.
func (c *GroupClient) Interceptors() []Interceptor {
	return c.inters.Group
}

func (c *GroupClient) mutate(ctx context.Context, m *GroupMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&GroupCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&GroupUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&GroupUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&GroupDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Group mutation op: %q", m.Op())
	}
}

// GroupRoleClient is a client for the GroupRole schema.
type GroupRoleClient struct {
	config
}

// NewGroupRoleClient returns a client for the GroupRole from the given config.
func NewGroupRoleClient(c config) *GroupRoleClient {
	return &GroupRoleClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `grouprole.Hooks(f(g(h())))`.
func (c *GroupRoleClient) Use(hooks ...Hook) {
	c.hooks.GroupRole = append(c.hooks.GroupRole, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `grouprole.Intercept(f(g(h())))`.
func (c *GroupRoleClient) Intercept(interceptors ...Interceptor) {
	c.inters.GroupRole = append(c.inters.GroupRole, interceptors...)
}

// Create returns a builder for creating a GroupRole entity.
func (c *GroupRoleClient) Create() *GroupRoleCreate {
	mutation := newGroupRoleMutation(c.config, OpCreate)
	return &GroupRoleCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of GroupRole entities.
func (c *GroupRoleClient) CreateBulk(builders ...*GroupRoleCreate) *GroupRoleCreateBulk {
	return &GroupRoleCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *GroupRoleClient) MapCreateBulk(slice any, setFunc func(*GroupRoleCreate, int)) *GroupRoleCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &GroupRoleCreateBulk{err: fmt.Errorf("calling to GroupRoleClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*GroupRoleCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &GroupRoleCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for GroupRole.
func (c *GroupRoleClient) Update() *GroupRoleUpdate {
	mutation := newGroupRoleMutation(c.config, OpUpdate)
	return &GroupRoleUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *GroupRoleClient) UpdateOne(gr *GroupRole) *GroupRoleUpdateOne {
	mutation := newGroupRoleMutation(c.config, OpUpdateOne, withGroupRole(gr))
	return &GroupRoleUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *GroupRoleClient) UpdateOneID(id string) *GroupRoleUpdateOne {
	mutation := newGroupRoleMutation(c.config, OpUpdateOne, withGroupRoleID(id))
	return &GroupRoleUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for GroupRole.
func (c *GroupRoleClient) Delete() *GroupRoleDelete {
	mutation := newGroupRoleMutation(c.config, OpDelete)
	return &GroupRoleDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *GroupRoleClient) DeleteOne(gr *GroupRole) *GroupRoleDeleteOne {
	return c.DeleteOneID(gr.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *GroupRoleClient) DeleteOneID(id string) *GroupRoleDeleteOne {
	builder := c.Delete().Where(grouprole.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &GroupRoleDeleteOne{builder}
}

// Query returns a query builder for GroupRole.
func (c *GroupRoleClient) Query() *GroupRoleQuery {
	return &GroupRoleQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeGroupRole},
		inters: c.Interceptors(),
	}
}

// Get returns a GroupRole entity by its id.
func (c *GroupRoleClient) Get(ctx context.Context, id string) (*GroupRole, error) {
	return c.Query().Where(grouprole.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *GroupRoleClient) GetX(ctx context.Context, id string) *GroupRole {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *GroupRoleClient) Hooks() []Hook {
	return c.hooks.GroupRole
}

// Interceptors returns the client interceptors.
func (c *GroupRoleClient) Interceptors() []Interceptor {
	return c.inters.GroupRole
}

func (c *GroupRoleClient) mutate(ctx context.Context, m *GroupRoleMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&GroupRoleCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&GroupRoleUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&GroupRoleUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&GroupRoleDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown GroupRole mutation op: %q", m.Op())
	}
}

// UserGroupClient is a client for the UserGroup schema.
type UserGroupClient struct {
	config
}

// NewUserGroupClient returns a client for the UserGroup from the given config.
func NewUserGroupClient(c config) *UserGroupClient {
	return &UserGroupClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `usergroup.Hooks(f(g(h())))`.
func (c *UserGroupClient) Use(hooks ...Hook) {
	c.hooks.UserGroup = append(c.hooks.UserGroup, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `usergroup.Intercept(f(g(h())))`.
func (c *UserGroupClient) Intercept(interceptors ...Interceptor) {
	c.inters.UserGroup = append(c.inters.UserGroup, interceptors...)
}

// Create returns a builder for creating a UserGroup entity.
func (c *UserGroupClient) Create() *UserGroupCreate {
	mutation := newUserGroupMutation(c.config, OpCreate)
	return &UserGroupCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of UserGroup entities.
func (c *UserGroupClient) CreateBulk(builders ...*UserGroupCreate) *UserGroupCreateBulk {
	return &UserGroupCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *UserGroupClient) MapCreateBulk(slice any, setFunc func(*UserGroupCreate, int)) *UserGroupCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &UserGroupCreateBulk{err: fmt.Errorf("calling to UserGroupClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*UserGroupCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &UserGroupCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for UserGroup.
func (c *UserGroupClient) Update() *UserGroupUpdate {
	mutation := newUserGroupMutation(c.config, OpUpdate)
	return &UserGroupUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *UserGroupClient) UpdateOne(ug *UserGroup) *UserGroupUpdateOne {
	mutation := newUserGroupMutation(c.config, OpUpdateOne, withUserGroup(ug))
	return &UserGroupUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *UserGroupClient) UpdateOneID(id string) *UserGroupUpdateOne {
	mutation := newUserGroupMutation(c.config, OpUpdateOne, withUserGroupID(id))
	return &UserGroupUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for UserGroup.
func (c *UserGroupClient) Delete() *UserGroupDelete {
	mutation := newUserGroupMutation(c.config, OpDelete)
	return &UserGroupDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *UserGroupClient) DeleteOne(ug *UserGroup) *UserGroupDeleteOne {
	return c.DeleteOneID(ug.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *UserGroupClient) DeleteOneID(id string) *UserGroupDeleteOne {
	builder := c.Delete().Where(usergroup.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &UserGroupDeleteOne{builder}
}

// Query returns a query builder for UserGroup.
func (c *UserGroupClient) Query() *UserGroupQuery {
	return &UserGroupQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeUserGroup},
		inters: c.Interceptors(),
	}
}

// Get returns a UserGroup entity by its id.
func (c *UserGroupClient) Get(ctx context.Context, id string) (*UserGroup, error) {
	return c.Query().Where(usergroup.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *UserGroupClient) GetX(ctx context.Context, id string) *UserGroup {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *UserGroupClient) Hooks() []Hook {
	return c.hooks.UserGroup
}

// Interceptors returns the client interceptors.
func (c *UserGroupClient) Interceptors() []Interceptor {
	return c.inters.UserGroup
}

func (c *UserGroupClient) mutate(ctx context.Context, m *UserGroupMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&UserGroupCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&UserGroupUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&UserGroupUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&UserGroupDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown UserGroup mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		Group, GroupRole, UserGroup []ent.Hook
	}
	inters struct {
		Group, GroupRole, UserGroup []ent.Interceptor
	}
)
