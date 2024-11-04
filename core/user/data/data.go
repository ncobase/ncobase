package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"

	"ncobase/core/user/data/ent"
	"ncobase/core/user/data/ent/migrate"

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

// New creates a new Database Connection.
func New(conf *config.Data) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	// get master connection, create ent client
	masterDB := d.DB()
	if masterDB == nil {
		return nil, nil, err
	}
	entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate) // master support migration
	if err != nil {
		return nil, nil, err
	}

	// get slave connection, create ent client
	var entClientRead *ent.Client
	if readDB, err := d.DBRead(); err == nil && readDB != nil {
		entClientRead, err = newEntClient(readDB, conf.Database.Master, false) // slave does not support migration
		if err != nil {
			log.Warnf(context.Background(), "Failed to create read-only ent client: %v", err)
		}
	}

	// no slave, use master
	if entClientRead == nil {
		entClientRead = entClient
	}

	return &Data{
		Data:   d,
		EC:     entClient,
		ECRead: entClientRead,
	}, cleanup, nil
}

// newEntClient creates a new ent client.
func newEntClient(db *sql.DB, conf *config.DBNode, enableMigrate bool) (*ent.Client, error) {
	client := ent.NewClient(ent.Driver(dialect.DebugWithContext(
		entsql.OpenDB(conf.Driver, db),
		func(ctx context.Context, i ...any) {
			if conf.Logging {
				log.Infof(ctx, "%v", i)
			}
		},
	)))

	// Enable SQL logging
	if conf.Logging {
		client = client.Debug()
	}

	// Auto migrate (only for master)
	if enableMigrate {
		if err := client.Schema.Create(context.Background(),
			migrate.WithForeignKeys(false),
			migrate.WithDropIndex(true),
			migrate.WithDropColumn(true),
		); err != nil {
			return nil, err
		}
	}

	return client, nil
}

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

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if d.EC != nil {
		if err := d.EC.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if d.ECRead != nil && d.ECRead != d.EC {
		if err := d.ECRead.Close(); err != nil {
			errs = append(errs, err)
		}
	}

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

func (d *Data) GetElasticsearchClient() *elastic.Client {
	return d.Conn.ES
}

// Ping checks the database connection
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

// GetEntTx retrieves ent transaction from context
func GetEntTx(ctx context.Context) (*ent.Tx, error) {
	tx, ok := ctx.Value("entTx").(*ent.Tx)
	if !ok {
		return nil, fmt.Errorf("ent transaction not found in context")
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

// WithEntTx wraps a function within an ent transaction for write operations
func (d *Data) WithEntTx(ctx context.Context, fn func(ctx context.Context, tx *ent.Tx) error) error {
	if d.EC == nil {
		return fmt.Errorf("ent client is nil")
	}

	tx, err := d.EC.Tx(ctx)
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

// WithEntTxRead wraps a function within an ent transaction for read-only operations
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
