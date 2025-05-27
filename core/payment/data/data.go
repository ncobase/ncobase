package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/payment/data/ent"
	"ncobase/payment/data/ent/migrate"

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

	ctx := context.Background()

	// get master connection
	masterDB := d.GetMasterDB()
	if masterDB == nil {
		return nil, cleanup, fmt.Errorf("master database connection is nil")
	}

	// create master ent client
	entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate, conf.Environment)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create master ent client: %v", err)
	}

	// get read connection
	var entClientRead *ent.Client
	if readDB, err := d.GetSlaveDB(); err == nil && readDB != nil {
		if readDB != masterDB {
			entClientRead, err = newEntClient(readDB, conf.Database.Master, false, conf.Environment) // slave does not support migration
			if err != nil {
				logger.Warnf(ctx, "Failed to create read-only ent client, will use master for reads: %v", err)
				entClientRead = entClient // fallback to master
			}
		} else {
			// Read DB is the same as master (no slaves available)
			entClientRead = entClient
		}
	} else {
		// Failed to get read DB, use master
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
			return nil, fmt.Errorf("failed to migrate database schema: %v", err)
		}
	}

	return client, nil
}

// GetMasterEntClient get master ent client for write operations
func (d *Data) GetMasterEntClient() *ent.Client {
	return d.EC
}

// GetSlaveEntClient get slave ent client for read operations
func (d *Data) GetSlaveEntClient() *ent.Client {
	if d.ECRead != nil {
		return d.ECRead
	}
	return d.EC // Fallback to master
}

// GetEntClientWithFallback returns the appropriate ent client based on operation type
func (d *Data) GetEntClientWithFallback(ctx context.Context, readOnly ...bool) *ent.Client {
	isReadOnly := false
	if len(readOnly) > 0 {
		isReadOnly = readOnly[0]
	}

	if !isReadOnly {
		// For write operations, always use master
		return d.GetMasterEntClient()
	}

	// For read operations, try read client first
	if d.ECRead != nil && d.ECRead != d.EC {
		// We have a separate read client, use it
		return d.ECRead
	}

	// Check if system is in read-only mode
	if d.IsReadOnlyMode(ctx) {
		logger.Warnf(ctx, "System is in read-only mode, using available read connection")
	}

	// Fallback to master
	return d.EC
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if d.EC != nil {
		if err := d.EC.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close master ent client: %v", err))
		}
	}

	if d.ECRead != nil && d.ECRead != d.EC {
		if err := d.ECRead.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close read ent client: %v", err))
		}
	}

	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
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
	client := d.GetEntClientWithFallback(ctx)
	if client == nil {
		return fmt.Errorf("ent client is nil")
	}

	tx, err := d.EC.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
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
	client := d.GetEntClientWithFallback(ctx, true)
	if client == nil {
		return fmt.Errorf("ent read client is nil")
	}

	tx, err := client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin read transaction: %v", err)
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
