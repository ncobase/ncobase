package templates

import "fmt"

func DataTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/elastic"
	"ncobase/common/meili"

	"github.com/redis/go-redis/v9"
)

// Data .
type Data struct {
	*data.Data
}

// New creates a new Database Connection.
func New(conf *config.Data) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	return &Data{
		Data: d,
	}, cleanup, nil
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
	}
	return errs
}

// GetDB get master database for write operations
func (d *Data) GetDB() *sql.DB {
	return d.DB()
}

// GetDBRead get slave database for read operations
func (d *Data) GetDBRead() (*sql.DB, error) {
	return d.DBRead()
}

// GetRedis get redis
func (d *Data) GetRedis() *redis.Client {
	return d.Conn.RC
}

// GetMeilisearch get meilisearch
func (d *Data) GetMeilisearch() *meili.Client {
	return d.Conn.MS
}

// GetElasticsearchClient get elasticsearch client
func (d *Data) GetElasticsearchClient() *elastic.Client {
	return d.Conn.ES
}

// Ping checks database connections health
func (d *Data) Ping(ctx context.Context) error {
	return d.Conn.Ping(ctx)
}

// GetTx retrieves transaction from context
func GetTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value("tx").(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("transaction not found in context")
	}
	return tx, nil
}

// WithTx wraps a function within a transaction
func (d *Data) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	db := d.GetDB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(context.WithValue(ctx, "tx", tx))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// WithTxRead wraps a function within a read-only transaction
func (d *Data) WithTxRead(ctx context.Context, fn func(ctx context.Context) error) error {
	dbRead, err := d.GetDBRead()
	if err != nil {
		return err
	}

	tx, err := dbRead.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}

	err = fn(context.WithValue(ctx, "tx", tx))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

/* Example usage:

// Write operations
db := d.GetDB()
_, err = db.Exec("INSERT INTO users (name) VALUES (?)", "test")

// Read operations
dbRead, err := d.GetDBRead()
if err != nil {
    // handle error
}
rows, err := dbRead.Query("SELECT * FROM users")

// Write transaction
err := d.WithTx(ctx, func(ctx context.Context) error {
    tx, err := GetTx(ctx)
    if err != nil {
        return err
    }
    _, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "test")
    return err
})

// Read-only transaction
err := d.WithTxRead(ctx, func(ctx context.Context) error {
    tx, err := GetTx(ctx)
    if err != nil {
        return err
    }
    rows, err := tx.Query("SELECT * FROM users")
    return err
})
*/
`)
}

func DataTemplateWithEnt(name, moduleType string) string {
	return fmt.Sprintf(`package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"
	"ncobase/%s/%s/data/ent"
	"ncobase/%s/%s/data/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/redis/go-redis/v9"
)

// Data .
type Data struct {
	*data.Data
	EC     *ent.Client // master ent client
	ECRead *ent.Client // slave ent client for read operations
}

// ... (其他基础方法与 DataTemplate 相同) ...

// GetEntClient get master ent client for write operations
func (d *Data) GetEntClient() *ent.Client {
	return d.EC
}

// GetEntClientRead get slave ent client for read operations
func (d *Data) GetEntClientRead() *ent.Client {
	if d.ECRead != nil {
		return d.ECRead
	}
	return d.EC // Downgrade, use master
}

// GetEntTx retrieves ent transaction from context
func GetEntTx(ctx context.Context) (*ent.Tx, error) {
	tx, ok := ctx.Value("entTx").(*ent.Tx)
	if !ok {
		return nil, fmt.Errorf("ent transaction not found in context")
	}
	return tx, nil
}

// WithEntTx wraps a function within an ent transaction
func (d *Data) WithEntTx(ctx context.Context, fn func(ctx context.Context, tx *ent.Tx) error) error {
	client := d.GetEntClient()
	if client == nil {
		return fmt.Errorf("ent client is nil")
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	err = fn(context.WithValue(ctx, "entTx", tx), tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// WithEntTxRead wraps a function within a read-only ent transaction
func (d *Data) WithEntTxRead(ctx context.Context, fn func(ctx context.Context, tx *ent.Tx) error) error {
	client := d.GetEntClientRead()
	if client == nil {
		return fmt.Errorf("ent read client is nil")
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}

	err = fn(context.WithValue(ctx, "entTx", tx), tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

/* Example usage:

// Ent operations
// Write
err := d.WithEntTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
    return tx.User.Create().
        SetName("test").
        Exec(ctx)
})

// Read
err := d.WithEntTxRead(ctx, func(ctx context.Context, tx *ent.Tx) error {
    users, err := tx.User.Query().
        Where(user.NameEQ("test")).
        All(ctx)
    return err
})

// Complex transaction
err := d.WithEntTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
    // Create user
    u, err := tx.User.Create().
        SetName("test").
        Save(ctx)
    if err != nil {
        return err
    }

    // Create relationship config
    _, err = tx.Config.Create().
        SetUser(u).
        SetKey("theme").
        SetValue("dark").
        Save(ctx)
    return err
})
*/
`, moduleType, name, moduleType, name)
}

func RepositoryTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package repository

import "ncobase/%s/%s/data"

// Repository represents the %s repository.
type Repository struct {
	// Add your repository fields here
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		// Initialize your repository fields here
	}
}

// Add your repository methods here
`, moduleType, name, name)
}

func SchemaTemplate() string {
	return `package schema

// Add your schema definitions here
`
}

func HandlerTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package handler

import "ncobase/%s/%s/service"

// Handler represents the %s handler.
type Handler struct {
	// Add your handler fields here
}

// New creates a new handler.
func New(s *service.Service) *Handler {
	return &Handler{
		// Initialize your handler fields here
	}
}

// Add your handler methods here
`, moduleType, name, name)
}

func ServiceTemplate(name, moduleType string) string {
	return fmt.Sprintf(`package service

import (
	"ncobase/common/config"
	"ncobase/%s/%s/data"
)

// Service represents the %s service.
type Service struct {
	// Add your service fields here
}

// New creates a new service.
func New(conf *config.Config, d *data.Data) *Service {
	return &Service{
		// Initialize your service fields here
	}
}

// Add your service methods here
`, moduleType, name, name)
}

func StructsTemplate() string {
	return `package structs

// Define your structs here
`
}
