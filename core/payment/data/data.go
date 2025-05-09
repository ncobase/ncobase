package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/core/payment/data/ent"
	"ncobase/core/payment/data/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/logging/logger"
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
	entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate, conf.Enveronment) // master support migration
	if err != nil {
		return nil, nil, err
	}

	// get slave connection, create ent client
	var entClientRead *ent.Client
	if readDB, err := d.DBRead(); err == nil && readDB != nil {
		entClientRead, err = newEntClient(readDB, conf.Database.Master, false, conf.Enveronment) // slave does not support migration
		if err != nil {
			logger.Warnf(context.Background(), "Failed to create read-only ent client: %v", err)
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
func newEntClient(db *sql.DB, conf *config.DBNode, enableMigrate bool, env ...string) (*ent.Client, error) {
	client := ent.NewClient(ent.Driver(dialect.DebugWithContext(
		entsql.OpenDB(conf.Driver, db),
		func(ctx context.Context, i ...any) {
			if conf.Logging {
				logger.Infof(ctx, "%v", i)
			}
		},
	)))

	// Enable SQL logging
	if conf.Logging {
		client = client.Debug()
	}

	// Auto migrate (only for master)
	if enableMigrate {
		migrateOpts := []schema.MigrateOption{
			migrate.WithForeignKeys(false),
			// migrate.WithGlobalUniqueID(true),
		}
		// Production does not support drop index and drop column
		if len(env) == 0 || (len(env) > 0 && env[0] != "production") {
			migrateOpts = append(migrateOpts, migrate.WithDropIndex(true), migrate.WithDropColumn(true))
		}
		if err := client.Schema.Create(context.Background(), migrateOpts...); err != nil {
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
			errs = append(errs, fmt.Errorf("failed to close master ent client: %w", err))
		}
	}

	if d.ECRead != nil && d.ECRead != d.EC {
		if err := d.ECRead.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close read ent client: %w", err))
		}
	}

	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		for _, err := range baseErrs {
			errs = append(errs, fmt.Errorf("base data error: %w", err))
		}
	}

	if len(errs) > 0 {
		// Log errors before returning them
		for _, err := range errs {
			logger.Error(context.Background(), err)
		}
	}

	return errs
}

// GetEntTx retrieves ent transaction from context
func (d *Data) GetEntTx(ctx context.Context) (*ent.Tx, error) {
	tx, ok := ctx.Value("entTx").(*ent.Tx)
	if !ok {
		return nil, fmt.Errorf("ent transaction not found in context")
	}
	return tx, nil
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
