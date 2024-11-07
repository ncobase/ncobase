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
	"ncobase/plugin/sample/data/ent"
	"ncobase/plugin/sample/data/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Data contains the shared resources and clients.
type Data struct {
	*data.Data
	EC         *ent.Client // master ent client
	ECRead     *ent.Client // slave ent client for read operations
	GormClient *gorm.DB    // master gorm client
	GormRead   *gorm.DB    // slave gorm client for read operations
}

// New creates a new Data instance with database connections.
func New(conf *config.Data) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	// get master connection,
	masterDB := d.DB()
	if masterDB == nil {
		return nil, nil, err
	}

	// create ent client
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

	// create gorm client
	gormClient, err := newGormClient(masterDB, conf.Database.Master)
	if err != nil {
		return nil, nil, err
	}

	var gormRead *gorm.DB
	if readDB, err := d.DBRead(); err == nil && readDB != nil {
		gormRead, err = newGormClient(readDB, conf.Database.Master)
		if err != nil {
			log.Warnf(context.Background(), "Failed to create read-only gorm client: %v", err)
		}
	}

	// no slave, use master
	if gormRead == nil {
		gormRead = gormClient
	}

	return &Data{
		Data:       d,
		EC:         entClient,
		ECRead:     entClientRead,
		GormClient: gormClient,
		GormRead:   gormRead,
	}, cleanup, nil
}

// GetDB get master database for write operations
func (d *Data) GetDB() *sql.DB {
	return d.DB()
}

// GetDBRead get slave database for read operations
func (d *Data) GetDBRead() (*sql.DB, error) {
	return d.DBRead()
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

// GetEntTx retrieves ent transaction from context
func GetEntTx(ctx context.Context) (*ent.Tx, error) {
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

// newGormClient creates a new GORM client.
func newGormClient(db *sql.DB, conf *config.DBNode) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch conf.Driver {
	case "postgres":
		dialector = postgres.New(postgres.Config{
			Conn: db,
		})
	case "mysql":
		dialector = mysql.New(mysql.Config{
			Conn: db,
		})
	case "sqlite3":
		dialector = sqlite.Open(conf.Source)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", conf.Driver)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	if conf.Logging {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	return gorm.Open(dialector, gormConfig)
}

// GetGormClient returns the master GORM client for write operations
func (d *Data) GetGormClient() *gorm.DB {
	return d.GormClient
}

// GetGormClientRead returns the slave GORM client for read operations
func (d *Data) GetGormClientRead() *gorm.DB {
	if d.GormRead != nil {
		return d.GormRead
	}
	return d.GormClient // Downgrade, use master
}

// GetGormTx retrieves gorm transaction from context
func GetGormTx(ctx context.Context) (*gorm.DB, error) {
	tx, ok := ctx.Value("gormTx").(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("gorm transaction not found in context")
	}
	return tx, nil
}

// WithGormTx wraps a function within a gorm transaction
func (d *Data) WithGormTx(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	if d.GormClient == nil {
		return fmt.Errorf("gorm client is nil")
	}

	return d.GormClient.Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, "gormTx", tx), tx)
	})
}

// WithGormTxRead wraps a function within a read-only gorm transaction
func (d *Data) WithGormTxRead(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	client := d.GetGormClientRead()
	if client == nil {
		return fmt.Errorf("gorm read client is nil")
	}

	// use sql.DB to create a read-only transaction
	sqlDB, err := client.DB()
	if err != nil {
		return err
	}

	// create a read-only transaction
	sqlTx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}

	// use sql.Tx to create a gorm.DB
	tx := client.Session(&gorm.Session{
		SkipHooks: true, // skip hooks
	}).WithContext(ctx)
	tx.Statement.ConnPool = sqlTx

	err = fn(context.WithValue(ctx, "gormTx", tx), tx)
	if err != nil {
		if rbErr := sqlTx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return sqlTx.Commit()
}

// Close closes all the resources in Data
func (d *Data) Close() (errs []error) {
	// Close ent clients
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

	// Close gorm clients
	if d.GormClient != nil {
		if db, err := d.GormClient.DB(); err == nil {
			if err := db.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if d.GormRead != nil && d.GormRead != d.GormClient {
		if db, err := d.GormRead.DB(); err == nil {
			if err := db.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// Close base resources
	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
	}

	return errs
}
