// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"ncobase/auth/data/ent/migrate"

	"ncobase/auth/data/ent/authtoken"
	"ncobase/auth/data/ent/codeauth"
	"ncobase/auth/data/ent/oauthuser"
	"ncobase/auth/data/ent/session"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

// Client is the client that holds all ent builders.
type Client struct {
	config
	// Schema is the client for creating, migrating and dropping schema.
	Schema *migrate.Schema
	// AuthToken is the client for interacting with the AuthToken builders.
	AuthToken *AuthTokenClient
	// CodeAuth is the client for interacting with the CodeAuth builders.
	CodeAuth *CodeAuthClient
	// OAuthUser is the client for interacting with the OAuthUser builders.
	OAuthUser *OAuthUserClient
	// Session is the client for interacting with the Session builders.
	Session *SessionClient
}

// NewClient creates a new client configured with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{config: newConfig(opts...)}
	client.init()
	return client
}

func (c *Client) init() {
	c.Schema = migrate.NewSchema(c.driver)
	c.AuthToken = NewAuthTokenClient(c.config)
	c.CodeAuth = NewCodeAuthClient(c.config)
	c.OAuthUser = NewOAuthUserClient(c.config)
	c.Session = NewSessionClient(c.config)
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
		AuthToken: NewAuthTokenClient(cfg),
		CodeAuth:  NewCodeAuthClient(cfg),
		OAuthUser: NewOAuthUserClient(cfg),
		Session:   NewSessionClient(cfg),
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
		AuthToken: NewAuthTokenClient(cfg),
		CodeAuth:  NewCodeAuthClient(cfg),
		OAuthUser: NewOAuthUserClient(cfg),
		Session:   NewSessionClient(cfg),
	}, nil
}

// Debug returns a new debug-client. It's used to get verbose logging on specific operations.
//
//	client.Debug().
//		AuthToken.
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
	c.AuthToken.Use(hooks...)
	c.CodeAuth.Use(hooks...)
	c.OAuthUser.Use(hooks...)
	c.Session.Use(hooks...)
}

// Intercept adds the query interceptors to all the entity clients.
// In order to add interceptors to a specific client, call: `client.Node.Intercept(...)`.
func (c *Client) Intercept(interceptors ...Interceptor) {
	c.AuthToken.Intercept(interceptors...)
	c.CodeAuth.Intercept(interceptors...)
	c.OAuthUser.Intercept(interceptors...)
	c.Session.Intercept(interceptors...)
}

// Mutate implements the ent.Mutator interface.
func (c *Client) Mutate(ctx context.Context, m Mutation) (Value, error) {
	switch m := m.(type) {
	case *AuthTokenMutation:
		return c.AuthToken.mutate(ctx, m)
	case *CodeAuthMutation:
		return c.CodeAuth.mutate(ctx, m)
	case *OAuthUserMutation:
		return c.OAuthUser.mutate(ctx, m)
	case *SessionMutation:
		return c.Session.mutate(ctx, m)
	default:
		return nil, fmt.Errorf("ent: unknown mutation type %T", m)
	}
}

// AuthTokenClient is a client for the AuthToken schema.
type AuthTokenClient struct {
	config
}

// NewAuthTokenClient returns a client for the AuthToken from the given config.
func NewAuthTokenClient(c config) *AuthTokenClient {
	return &AuthTokenClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `authtoken.Hooks(f(g(h())))`.
func (c *AuthTokenClient) Use(hooks ...Hook) {
	c.hooks.AuthToken = append(c.hooks.AuthToken, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `authtoken.Intercept(f(g(h())))`.
func (c *AuthTokenClient) Intercept(interceptors ...Interceptor) {
	c.inters.AuthToken = append(c.inters.AuthToken, interceptors...)
}

// Create returns a builder for creating a AuthToken entity.
func (c *AuthTokenClient) Create() *AuthTokenCreate {
	mutation := newAuthTokenMutation(c.config, OpCreate)
	return &AuthTokenCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of AuthToken entities.
func (c *AuthTokenClient) CreateBulk(builders ...*AuthTokenCreate) *AuthTokenCreateBulk {
	return &AuthTokenCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *AuthTokenClient) MapCreateBulk(slice any, setFunc func(*AuthTokenCreate, int)) *AuthTokenCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &AuthTokenCreateBulk{err: fmt.Errorf("calling to AuthTokenClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*AuthTokenCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &AuthTokenCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for AuthToken.
func (c *AuthTokenClient) Update() *AuthTokenUpdate {
	mutation := newAuthTokenMutation(c.config, OpUpdate)
	return &AuthTokenUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *AuthTokenClient) UpdateOne(at *AuthToken) *AuthTokenUpdateOne {
	mutation := newAuthTokenMutation(c.config, OpUpdateOne, withAuthToken(at))
	return &AuthTokenUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *AuthTokenClient) UpdateOneID(id string) *AuthTokenUpdateOne {
	mutation := newAuthTokenMutation(c.config, OpUpdateOne, withAuthTokenID(id))
	return &AuthTokenUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for AuthToken.
func (c *AuthTokenClient) Delete() *AuthTokenDelete {
	mutation := newAuthTokenMutation(c.config, OpDelete)
	return &AuthTokenDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *AuthTokenClient) DeleteOne(at *AuthToken) *AuthTokenDeleteOne {
	return c.DeleteOneID(at.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *AuthTokenClient) DeleteOneID(id string) *AuthTokenDeleteOne {
	builder := c.Delete().Where(authtoken.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &AuthTokenDeleteOne{builder}
}

// Query returns a query builder for AuthToken.
func (c *AuthTokenClient) Query() *AuthTokenQuery {
	return &AuthTokenQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeAuthToken},
		inters: c.Interceptors(),
	}
}

// Get returns a AuthToken entity by its id.
func (c *AuthTokenClient) Get(ctx context.Context, id string) (*AuthToken, error) {
	return c.Query().Where(authtoken.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *AuthTokenClient) GetX(ctx context.Context, id string) *AuthToken {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *AuthTokenClient) Hooks() []Hook {
	return c.hooks.AuthToken
}

// Interceptors returns the client interceptors.
func (c *AuthTokenClient) Interceptors() []Interceptor {
	return c.inters.AuthToken
}

func (c *AuthTokenClient) mutate(ctx context.Context, m *AuthTokenMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&AuthTokenCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&AuthTokenUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&AuthTokenUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&AuthTokenDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown AuthToken mutation op: %q", m.Op())
	}
}

// CodeAuthClient is a client for the CodeAuth schema.
type CodeAuthClient struct {
	config
}

// NewCodeAuthClient returns a client for the CodeAuth from the given config.
func NewCodeAuthClient(c config) *CodeAuthClient {
	return &CodeAuthClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `codeauth.Hooks(f(g(h())))`.
func (c *CodeAuthClient) Use(hooks ...Hook) {
	c.hooks.CodeAuth = append(c.hooks.CodeAuth, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `codeauth.Intercept(f(g(h())))`.
func (c *CodeAuthClient) Intercept(interceptors ...Interceptor) {
	c.inters.CodeAuth = append(c.inters.CodeAuth, interceptors...)
}

// Create returns a builder for creating a CodeAuth entity.
func (c *CodeAuthClient) Create() *CodeAuthCreate {
	mutation := newCodeAuthMutation(c.config, OpCreate)
	return &CodeAuthCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of CodeAuth entities.
func (c *CodeAuthClient) CreateBulk(builders ...*CodeAuthCreate) *CodeAuthCreateBulk {
	return &CodeAuthCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *CodeAuthClient) MapCreateBulk(slice any, setFunc func(*CodeAuthCreate, int)) *CodeAuthCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &CodeAuthCreateBulk{err: fmt.Errorf("calling to CodeAuthClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*CodeAuthCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &CodeAuthCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for CodeAuth.
func (c *CodeAuthClient) Update() *CodeAuthUpdate {
	mutation := newCodeAuthMutation(c.config, OpUpdate)
	return &CodeAuthUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *CodeAuthClient) UpdateOne(ca *CodeAuth) *CodeAuthUpdateOne {
	mutation := newCodeAuthMutation(c.config, OpUpdateOne, withCodeAuth(ca))
	return &CodeAuthUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *CodeAuthClient) UpdateOneID(id string) *CodeAuthUpdateOne {
	mutation := newCodeAuthMutation(c.config, OpUpdateOne, withCodeAuthID(id))
	return &CodeAuthUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for CodeAuth.
func (c *CodeAuthClient) Delete() *CodeAuthDelete {
	mutation := newCodeAuthMutation(c.config, OpDelete)
	return &CodeAuthDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *CodeAuthClient) DeleteOne(ca *CodeAuth) *CodeAuthDeleteOne {
	return c.DeleteOneID(ca.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *CodeAuthClient) DeleteOneID(id string) *CodeAuthDeleteOne {
	builder := c.Delete().Where(codeauth.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &CodeAuthDeleteOne{builder}
}

// Query returns a query builder for CodeAuth.
func (c *CodeAuthClient) Query() *CodeAuthQuery {
	return &CodeAuthQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeCodeAuth},
		inters: c.Interceptors(),
	}
}

// Get returns a CodeAuth entity by its id.
func (c *CodeAuthClient) Get(ctx context.Context, id string) (*CodeAuth, error) {
	return c.Query().Where(codeauth.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *CodeAuthClient) GetX(ctx context.Context, id string) *CodeAuth {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *CodeAuthClient) Hooks() []Hook {
	return c.hooks.CodeAuth
}

// Interceptors returns the client interceptors.
func (c *CodeAuthClient) Interceptors() []Interceptor {
	return c.inters.CodeAuth
}

func (c *CodeAuthClient) mutate(ctx context.Context, m *CodeAuthMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&CodeAuthCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&CodeAuthUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&CodeAuthUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&CodeAuthDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown CodeAuth mutation op: %q", m.Op())
	}
}

// OAuthUserClient is a client for the OAuthUser schema.
type OAuthUserClient struct {
	config
}

// NewOAuthUserClient returns a client for the OAuthUser from the given config.
func NewOAuthUserClient(c config) *OAuthUserClient {
	return &OAuthUserClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `oauthuser.Hooks(f(g(h())))`.
func (c *OAuthUserClient) Use(hooks ...Hook) {
	c.hooks.OAuthUser = append(c.hooks.OAuthUser, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `oauthuser.Intercept(f(g(h())))`.
func (c *OAuthUserClient) Intercept(interceptors ...Interceptor) {
	c.inters.OAuthUser = append(c.inters.OAuthUser, interceptors...)
}

// Create returns a builder for creating a OAuthUser entity.
func (c *OAuthUserClient) Create() *OAuthUserCreate {
	mutation := newOAuthUserMutation(c.config, OpCreate)
	return &OAuthUserCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of OAuthUser entities.
func (c *OAuthUserClient) CreateBulk(builders ...*OAuthUserCreate) *OAuthUserCreateBulk {
	return &OAuthUserCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *OAuthUserClient) MapCreateBulk(slice any, setFunc func(*OAuthUserCreate, int)) *OAuthUserCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &OAuthUserCreateBulk{err: fmt.Errorf("calling to OAuthUserClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*OAuthUserCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &OAuthUserCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for OAuthUser.
func (c *OAuthUserClient) Update() *OAuthUserUpdate {
	mutation := newOAuthUserMutation(c.config, OpUpdate)
	return &OAuthUserUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *OAuthUserClient) UpdateOne(ou *OAuthUser) *OAuthUserUpdateOne {
	mutation := newOAuthUserMutation(c.config, OpUpdateOne, withOAuthUser(ou))
	return &OAuthUserUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *OAuthUserClient) UpdateOneID(id string) *OAuthUserUpdateOne {
	mutation := newOAuthUserMutation(c.config, OpUpdateOne, withOAuthUserID(id))
	return &OAuthUserUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for OAuthUser.
func (c *OAuthUserClient) Delete() *OAuthUserDelete {
	mutation := newOAuthUserMutation(c.config, OpDelete)
	return &OAuthUserDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *OAuthUserClient) DeleteOne(ou *OAuthUser) *OAuthUserDeleteOne {
	return c.DeleteOneID(ou.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *OAuthUserClient) DeleteOneID(id string) *OAuthUserDeleteOne {
	builder := c.Delete().Where(oauthuser.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &OAuthUserDeleteOne{builder}
}

// Query returns a query builder for OAuthUser.
func (c *OAuthUserClient) Query() *OAuthUserQuery {
	return &OAuthUserQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeOAuthUser},
		inters: c.Interceptors(),
	}
}

// Get returns a OAuthUser entity by its id.
func (c *OAuthUserClient) Get(ctx context.Context, id string) (*OAuthUser, error) {
	return c.Query().Where(oauthuser.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *OAuthUserClient) GetX(ctx context.Context, id string) *OAuthUser {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *OAuthUserClient) Hooks() []Hook {
	return c.hooks.OAuthUser
}

// Interceptors returns the client interceptors.
func (c *OAuthUserClient) Interceptors() []Interceptor {
	return c.inters.OAuthUser
}

func (c *OAuthUserClient) mutate(ctx context.Context, m *OAuthUserMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&OAuthUserCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&OAuthUserUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&OAuthUserUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&OAuthUserDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown OAuthUser mutation op: %q", m.Op())
	}
}

// SessionClient is a client for the Session schema.
type SessionClient struct {
	config
}

// NewSessionClient returns a client for the Session from the given config.
func NewSessionClient(c config) *SessionClient {
	return &SessionClient{config: c}
}

// Use adds a list of mutation hooks to the hooks stack.
// A call to `Use(f, g, h)` equals to `session.Hooks(f(g(h())))`.
func (c *SessionClient) Use(hooks ...Hook) {
	c.hooks.Session = append(c.hooks.Session, hooks...)
}

// Intercept adds a list of query interceptors to the interceptors stack.
// A call to `Intercept(f, g, h)` equals to `session.Intercept(f(g(h())))`.
func (c *SessionClient) Intercept(interceptors ...Interceptor) {
	c.inters.Session = append(c.inters.Session, interceptors...)
}

// Create returns a builder for creating a Session entity.
func (c *SessionClient) Create() *SessionCreate {
	mutation := newSessionMutation(c.config, OpCreate)
	return &SessionCreate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// CreateBulk returns a builder for creating a bulk of Session entities.
func (c *SessionClient) CreateBulk(builders ...*SessionCreate) *SessionCreateBulk {
	return &SessionCreateBulk{config: c.config, builders: builders}
}

// MapCreateBulk creates a bulk creation builder from the given slice. For each item in the slice, the function creates
// a builder and applies setFunc on it.
func (c *SessionClient) MapCreateBulk(slice any, setFunc func(*SessionCreate, int)) *SessionCreateBulk {
	rv := reflect.ValueOf(slice)
	if rv.Kind() != reflect.Slice {
		return &SessionCreateBulk{err: fmt.Errorf("calling to SessionClient.MapCreateBulk with wrong type %T, need slice", slice)}
	}
	builders := make([]*SessionCreate, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		builders[i] = c.Create()
		setFunc(builders[i], i)
	}
	return &SessionCreateBulk{config: c.config, builders: builders}
}

// Update returns an update builder for Session.
func (c *SessionClient) Update() *SessionUpdate {
	mutation := newSessionMutation(c.config, OpUpdate)
	return &SessionUpdate{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOne returns an update builder for the given entity.
func (c *SessionClient) UpdateOne(s *Session) *SessionUpdateOne {
	mutation := newSessionMutation(c.config, OpUpdateOne, withSession(s))
	return &SessionUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// UpdateOneID returns an update builder for the given id.
func (c *SessionClient) UpdateOneID(id string) *SessionUpdateOne {
	mutation := newSessionMutation(c.config, OpUpdateOne, withSessionID(id))
	return &SessionUpdateOne{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// Delete returns a delete builder for Session.
func (c *SessionClient) Delete() *SessionDelete {
	mutation := newSessionMutation(c.config, OpDelete)
	return &SessionDelete{config: c.config, hooks: c.Hooks(), mutation: mutation}
}

// DeleteOne returns a builder for deleting the given entity.
func (c *SessionClient) DeleteOne(s *Session) *SessionDeleteOne {
	return c.DeleteOneID(s.ID)
}

// DeleteOneID returns a builder for deleting the given entity by its id.
func (c *SessionClient) DeleteOneID(id string) *SessionDeleteOne {
	builder := c.Delete().Where(session.ID(id))
	builder.mutation.id = &id
	builder.mutation.op = OpDeleteOne
	return &SessionDeleteOne{builder}
}

// Query returns a query builder for Session.
func (c *SessionClient) Query() *SessionQuery {
	return &SessionQuery{
		config: c.config,
		ctx:    &QueryContext{Type: TypeSession},
		inters: c.Interceptors(),
	}
}

// Get returns a Session entity by its id.
func (c *SessionClient) Get(ctx context.Context, id string) (*Session, error) {
	return c.Query().Where(session.ID(id)).Only(ctx)
}

// GetX is like Get, but panics if an error occurs.
func (c *SessionClient) GetX(ctx context.Context, id string) *Session {
	obj, err := c.Get(ctx, id)
	if err != nil {
		panic(err)
	}
	return obj
}

// Hooks returns the client hooks.
func (c *SessionClient) Hooks() []Hook {
	return c.hooks.Session
}

// Interceptors returns the client interceptors.
func (c *SessionClient) Interceptors() []Interceptor {
	return c.inters.Session
}

func (c *SessionClient) mutate(ctx context.Context, m *SessionMutation) (Value, error) {
	switch m.Op() {
	case OpCreate:
		return (&SessionCreate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdate:
		return (&SessionUpdate{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpUpdateOne:
		return (&SessionUpdateOne{config: c.config, hooks: c.Hooks(), mutation: m}).Save(ctx)
	case OpDelete, OpDeleteOne:
		return (&SessionDelete{config: c.config, hooks: c.Hooks(), mutation: m}).Exec(ctx)
	default:
		return nil, fmt.Errorf("ent: unknown Session mutation op: %q", m.Op())
	}
}

// hooks and interceptors per client, for fast access.
type (
	hooks struct {
		AuthToken, CodeAuth, OAuthUser, Session []ent.Hook
	}
	inters struct {
		AuthToken, CodeAuth, OAuthUser, Session []ent.Interceptor
	}
)
