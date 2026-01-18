package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/plugin/sample/data/ent"
	"ncobase/plugin/sample/data/ent/migrate"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/logging/logger"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gl "gorm.io/gorm/logger"
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
func New(conf *config.Data, env ...string) (*Data, func(name ...string), error) {
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
	entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate, env...)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create master ent client: %v", err)
	}

	// get read connection
	var entClientRead *ent.Client
	if readDB, err := d.GetSlaveDB(); err == nil && readDB != nil {
		if readDB != masterDB {
			entClientRead, err = newEntClient(readDB, conf.Database.Master, false, env...) // slave does not support migration
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

	// create gorm client
	gormClient, err := newGormClient(masterDB, conf.Database.Master)
	if err != nil {
		return nil, nil, err
	}

	var gormRead *gorm.DB
	if readDB, err := d.GetSlaveDB(); err == nil && readDB != nil {
		if readDB != masterDB {
			gormRead, err = newGormClient(readDB, conf.Database.Master)
			if err != nil {
				logger.Warnf(ctx, "Failed to create read-only gorm client, will use master for reads: %v", err)
				gormRead = gormClient // fallback to master
			}
		} else {
			// Read DB is the same as master (no slaves available)
			gormRead = gormClient
		}
	} else {
		// Failed to get read DB, use master
		gormRead = gormClient
	}

	// MongoDB is optional - check if it's configured
	var mongoMaster *mongo.Client
	var mongoSlave *mongo.Client

	mongoManager := d.GetMongoManager()
	if mongoManager != nil {
		// Try to get master client through interface
		if mgmInterface, ok := mongoManager.(interface{ Master() *mongo.Client }); ok {
			mongoMaster = mgmInterface.Master()

			// Try to get slave client
			if slaveInterface, ok := mongoManager.(interface{ Slave() (*mongo.Client, error) }); ok {
				mongoSlave, _ = slaveInterface.Slave()
			}
		} else if client, ok := mongoManager.(*mongo.Client); ok {
			// If mongo manager is already a *mongo.Client, use it directly
			mongoMaster = client
		}

		// Fallback slave to master if not available
		if mongoSlave == nil {
			mongoSlave = mongoMaster
		}
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
		Logger: gl.Default.LogMode(1),
	}

	if conf.Logging {
		gormConfig.Logger = gl.Default.LogMode(4)
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

// GetMongoDatabase returns a MongoDB database using ncore v0.2 API
func (d *Data) GetMongoDatabase(dbName string, readOnly bool) (*mongo.Database, error) {
	if d.Data == nil {
		return nil, fmt.Errorf("base data layer is nil")
	}

	// Use ncore v0.2 unified API
	db, err := d.Data.GetMongoDatabase(dbName, readOnly)
	if err != nil {
		return nil, err
	}

	// Type assert to *mongo.Database
	mongoDb, ok := db.(*mongo.Database)
	if !ok {
		return nil, fmt.Errorf("unexpected mongo database type: %T", db)
	}

	return mongoDb, nil
}

// GetMongoCollection returns a collection using ncore v0.2 API
func (d *Data) GetMongoCollection(dbName, collName string, readOnly bool) (*mongo.Collection, error) {
	if d.Data == nil {
		return nil, fmt.Errorf("base data layer is nil")
	}

	// Use ncore v0.2 unified API
	coll, err := d.Data.GetMongoCollection(dbName, collName, readOnly)
	if err != nil {
		return nil, err
	}

	// Type assert to *mongo.Collection
	mongoColl, ok := coll.(*mongo.Collection)
	if !ok {
		return nil, fmt.Errorf("unexpected mongo collection type: %T", coll)
	}

	return mongoColl, nil
}

// GetMongoCollectionDirect returns a collection from cached client (for backward compatibility)
// Deprecated: Use GetMongoCollection instead for better error handling
func (d *Data) GetMongoCollectionDirect(dbName, collName string, readOnly bool) *mongo.Collection {
	if readOnly && d.MCRead != nil {
		return d.MCRead.Database(dbName).Collection(collName)
	}
	if d.MC != nil {
		return d.MC.Database(dbName).Collection(collName)
	}
	return nil
}

// GetMongoTx retrieves mongo transaction from context
func GetMongoTx(ctx context.Context) (mongo.SessionContext, error) {
	session, ok := ctx.Value("mongoTx").(mongo.SessionContext)
	if !ok {
		return nil, fmt.Errorf("mongo session not found in context")
	}
	return session, nil
}

// WithMongoTx wraps a function within a mongo transaction using ncore v0.2 API
func (d *Data) WithMongoTx(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	if d.Data == nil {
		return fmt.Errorf("base data layer is nil")
	}

	// Use ncore v0.2 unified transaction API
	return d.Data.WithMongoTransaction(ctx, func(sessCtx any) error {
		mongoCtx, ok := sessCtx.(mongo.SessionContext)
		if !ok {
			return fmt.Errorf("unexpected session context type: %T", sessCtx)
		}
		return fn(mongoCtx)
	})
}

// WithMongoTxDirect wraps a function within a mongo transaction using cached client
// Deprecated: Use WithMongoTx for ncore v0.2 compatibility
func (d *Data) WithMongoTxDirect(ctx context.Context, fn func(mongo.SessionContext) error) error {
	if d.MC == nil {
		return fmt.Errorf("mongo client is nil")
	}

	session, err := d.MC.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (any, error) {
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
