package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/log"
	"ncobase/plugin/sample/data/ent"
	"ncobase/plugin/sample/data/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"go.mongodb.org/mongo-driver/mongo"
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
	EC         *ent.Client   // master ent client
	ECRead     *ent.Client   // slave ent client for read operations
	GormClient *gorm.DB      // master gorm client
	GormRead   *gorm.DB      // slave gorm client for read operations
	MC         *mongo.Client // master mongo client
	MCRead     *mongo.Client // slave mongo client for read operations
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

	// get mongo master connection
	mongoMaster := d.Conn.MGM.Master()
	if mongoMaster == nil {
		return nil, nil, fmt.Errorf("mongo master client is nil")
	}

	// get mongo slave connection
	mongoSlave, err := d.Conn.MGM.Slave()
	if err != nil {
		log.Warnf(context.Background(), "Failed to get read-only mongo client: %v", err)
	}

	// no slave, use master
	if mongoSlave == nil {
		mongoSlave = mongoMaster
	}

	return &Data{
		Data:       d,
		EC:         entClient,
		ECRead:     entClientRead,
		GormClient: gormClient,
		GormRead:   gormRead,
		MC:         mongoMaster,
		MCRead:     mongoSlave,
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

// GetMongoClient get master mongo client for write operations
func (d *Data) GetMongoClient() *mongo.Client {
	return d.MC
}

// GetMongoClientRead get slave mongo client for read operations
func (d *Data) GetMongoClientRead() *mongo.Client {
	if d.MCRead != nil {
		return d.MCRead
	}
	return d.MC // Downgrade, use master
}

// GetMongoCollection returns a collection from master/slave client
func (d *Data) GetMongoCollection(dbName, collName string, readOnly bool) *mongo.Collection {
	if readOnly {
		return d.MCRead.Database(dbName).Collection(collName)
	}
	return d.MC.Database(dbName).Collection(collName)
}

// GetMongoTx retrieves mongo transaction from context
func GetMongoTx(ctx context.Context) (mongo.SessionContext, error) {
	session, ok := ctx.Value("mongoTx").(mongo.SessionContext)
	if !ok {
		return nil, fmt.Errorf("mongo session not found in context")
	}
	return session, nil
}

// WithMongoTx wraps a function within a mongo transaction
func (d *Data) WithMongoTx(ctx context.Context, fn func(mongo.SessionContext) error) error {
	if d.MC == nil {
		return fmt.Errorf("mongo client is nil")
	}

	session, err := d.MC.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
}

// WithMongoTxRead wraps a function within a read-only mongo transaction
func (d *Data) WithMongoTxRead(ctx context.Context, fn func(mongo.SessionContext) error) error {
	client := d.GetMongoClientRead()
	if client == nil {
		return fmt.Errorf("mongo read client is nil")
	}

	session, err := client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// MongoDB does not support read-only transaction, so we downgrade to read-write
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})
	return err
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
