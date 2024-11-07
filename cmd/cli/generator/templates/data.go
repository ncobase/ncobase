package templates

import "fmt"

func DataTemplate(name, extType string) string {
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

func DataTemplateWithEnt(name, extType string) string {
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
`, extType, name, extType, name)
}

func DataTemplateWithGorm(name, extType string) string {
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

    "github.com/redis/go-redis/v9"
    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// Data .
type Data struct {
    *data.Data
    GormClient *gorm.DB    // master gorm client
    GormRead   *gorm.DB    // slave gorm client for read operations
}

// New creates a new Database Connection.
func New(conf *config.Data) (*Data, func(name ...string), error) {
    d, cleanup, err := data.New(conf)
    if err != nil {
        return nil, nil, err
    }

    // get master connection
    masterDB := d.DB()
    if masterDB == nil {
        return nil, nil, err
    }

    // create gorm master client
    gormClient, err := newGormClient(masterDB, conf.Database.Master)
    if err != nil {
        return nil, nil, err
    }

    // create gorm read client
    var gormRead *gorm.DB
    if readDB, err := d.DBRead(); err == nil && readDB != nil {
        gormRead, err = newGormClient(readDB, conf.Database.Master)
        if err != nil {
            log.Warnf(context.Background(), "Failed to create read-only gorm client: %v", err)
        }
    }

    // if no slave available, use master
    if gormRead == nil {
        gormRead = gormClient
    }

    return &Data{
        Data:       d,
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

// GetElasticsearchClient get elasticsearch client
func (d *Data) GetElasticsearchClient() *elastic.Client {
    return d.Conn.ES
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

// WithGormTxRead wraps a function within a transaction using read replica
func (d *Data) WithGormTxRead(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
    client := d.GetGormClientRead()
    if client == nil {
        return fmt.Errorf("gorm read client is nil")
    }

    sqlDB, err := client.DB()
    if err != nil {
        return err
    }

    sqlTx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{
        ReadOnly: true,
    })
    if err != nil {
        return err
    }

    tx := client.Session(&gorm.Session{
        SkipHooks: true,
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

/* Example usage:

// Write operations with GORM
db := d.GetGormClient()
result := db.Create(&User{Name: "test"})

// Read operations with GORM
dbRead := d.GetGormClientRead()
var users []User
result := dbRead.Where("name = ?", "test").Find(&users)

// Write transaction with GORM
err := d.WithGormTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
    // Create user
    if err := tx.Create(&User{Name: "test"}).Error; err != nil {
        return err
    }

    // Create user config
    if err := tx.Create(&Config{
        UserID: user.ID,
        Key: "theme",
        Value: "dark",
    }).Error; err != nil {
        return err
    }

    return nil
})

// Read-only transaction with GORM
err := d.WithGormTxRead(ctx, func(ctx context.Context, tx *gorm.DB) error {
    var users []User
    if err := tx.Where("status = ?", "active").Find(&users).Error; err != nil {
        return err
    }
    return nil
})
*/
`)
}

func DataTemplateComplete(name, extType string) string {
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
    "gorm.io/driver/mysql"
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

    // get master connection
    masterDB := d.DB()
    if masterDB == nil {
        return nil, nil, err
    }

    // create ent master client
    entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate)
    if err != nil {
        return nil, nil, err
    }

    // create ent read client
    var entClientRead *ent.Client
    if readDB, err := d.DBRead(); err == nil && readDB != nil {
        entClientRead, err = newEntClient(readDB, conf.Database.Master, false)
        if err != nil {
            log.Warnf(context.Background(), "Failed to create read-only ent client: %v", err)
        }
    }

    // create gorm master client
    gormClient, err := newGormClient(masterDB, conf.Database.Master)
    if err != nil {
        return nil, nil, err
    }

    // create gorm read client
    var gormRead *gorm.DB
    if readDB, err := d.DBRead(); err == nil && readDB != nil {
        gormRead, err = newGormClient(readDB, conf.Database.Master)
        if err != nil {
            log.Warnf(context.Background(), "Failed to create read-only gorm client: %v", err)
        }
    }

    // if no slave available, use master
    if entClientRead == nil {
        entClientRead = entClient
    }
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

// newEntClient creates a new ent client.
func newEntClient(db *sql.DB, conf *config.DBNode, enableMigrate bool) (*ent.Client, error) {
    client := ent.NewClient(ent.Driver(dialect.DebugWithContext(
        entsql.OpenDB(conf.Driver, db),
        func(ctx context.Context, i ...any) {
            if conf.Logging {
                log.Infof(ctx, "%%v", i)
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
        return nil, fmt.Errorf("unsupported database driver: %%s", conf.Driver)
    }

    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    }

    if conf.Logging {
        gormConfig.Logger = logger.Default.LogMode(logger.Info)
    }

    return gorm.Open(dialector, gormConfig)
}

// GetDB get master database for write operations
func (d *Data) GetDB() *sql.DB {
    return d.DB()
}

// GetDBRead get slave database for read operations
func (d *Data) GetDBRead() (*sql.DB, error) {
    return d.DBRead()
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

// GetEntTx retrieves ent transaction from context
func GetEntTx(ctx context.Context) (*ent.Tx, error) {
    tx, ok := ctx.Value("entTx").(*ent.Tx)
    if !ok {
        return nil, fmt.Errorf("ent transaction not found in context")
    }
    return tx, nil
}

// GetGormTx retrieves gorm transaction from context
func GetGormTx(ctx context.Context) (*gorm.DB, error) {
    tx, ok := ctx.Value("gormTx").(*gorm.DB)
    if !ok {
        return nil, fmt.Errorf("gorm transaction not found in context")
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
            return fmt.Errorf("tx err: %%v, rb err: %%v", err, rbErr)
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
            return fmt.Errorf("tx err: %%v, rb err: %%v", err, rbErr)
        }
        return err
    }

    return tx.Commit()
}

// WithEntTx wraps a function within an ent transaction
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
            return fmt.Errorf("tx err: %%v, rb err: %%v", err, rbErr)
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
            return fmt.Errorf("tx err: %%v, rb err: %%v", err, rbErr)
        }
        return err
    }

    return tx.Commit()
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

    sqlDB, err := client.DB()
    if err != nil {
        return err
    }

    sqlTx, err := sqlDB.BeginTx(ctx, &sql.TxOptions{
        ReadOnly: true,
    })
    if err != nil {
        return err
    }

    tx := client.Session(&gorm.Session{
        SkipHooks: true,
    }).WithContext(ctx)
    tx.Statement.ConnPool = sqlTx

    err = fn(context.WithValue(ctx, "gormTx", tx), tx)
    if err != nil {
        if rbErr := sqlTx.Rollback(); rbErr != nil {
            return fmt.Errorf("tx err: %%v, rb err: %%v", err, rbErr)
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

/* Example usage (continued):

// GORM operations
// Write
err := d.WithGormTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
    // Create user
    user := &User{Name: "test"}
    if err := tx.Create(user).Error; err != nil {
        return err
    }

    // Create related config
    config := &Config{
        UserID: user.ID,
        Key:    "theme",
        Value:  "dark",
    }
    return tx.Create(config).Error
})

// Read
err := d.WithGormTxRead(ctx, func(ctx context.Context, tx *gorm.DB) error {
    var users []User
    if err := tx.Where("status = ?", "active").Find(&users).Error; err != nil {
        return err
    }

    // Process users...
    return nil
})

// Mixed usage example (using different ORMs in same flow)
err := d.WithTx(ctx, func(ctx context.Context) error {
    // 1. First do something with native SQL
    tx, err := GetTx(ctx)
    if err != nil {
        return err
    }
    result, err := tx.Exec("UPDATE users SET status = ? WHERE id = ?", "active", 1)
    if err != nil {
        return err
    }

    // 2. Then use Ent for some operations
    err = d.WithEntTx(ctx, func(ctx context.Context, etx *ent.Tx) error {
        return etx.User.Create().
            SetName("test").
            SetStatus("active").
            Exec(ctx)
    })
    if err != nil {
        return err
    }

    // 3. Finally use GORM for bulk operations
    return d.WithGormTx(ctx, func(ctx context.Context, gtx *gorm.DB) error {
        users := []User{{Name: "user1"}, {Name: "user2"}}
        return gtx.CreateInBatches(users, 100).Error
    })
})

// Service layer usage example:
type UserService struct {
    data *Data
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    return s.data.WithGormTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
        // Create user
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        // Create default settings
        settings := []UserSetting{
            {UserID: user.ID, Key: "theme", Value: "light"},
            {UserID: user.ID, Key: "language", Value: "en"},
        }
        return tx.Create(&settings).Error
    })
}

func (s *UserService) GetUserWithSettings(ctx context.Context, userID int64) (*User, error) {
    var user User
    err := s.data.WithGormTxRead(ctx, func(ctx context.Context, tx *gorm.DB) error {
        return tx.Preload("Settings").First(&user, userID).Error
    })
    return &user, err
}

func (s *UserService) UpdateUserStatus(ctx context.Context, userID int64, status string) error {
    return s.data.WithEntTx(ctx, func(ctx context.Context, tx *ent.Tx) error {
        return tx.User.UpdateOneID(userID).
            SetStatus(status).
            Exec(ctx)
    })
}

// Repository layer usage example:
type UserRepository struct {
    data *Data
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    return r.data.WithGormTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
        return tx.Create(user).Error
    })
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
    var user User
    err := r.data.WithGormTxRead(ctx, func(ctx context.Context, tx *gorm.DB) error {
        return tx.First(&user, id).Error
    })
    return &user, err
}

func (r *UserRepository) List(ctx context.Context, condition map[string]interface{}) ([]User, error) {
    var users []User
    err := r.data.WithGormTxRead(ctx, func(ctx context.Context, tx *gorm.DB) error {
        return tx.Where(condition).Find(&users).Error
    })
    return users, err
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
    return r.data.WithGormTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
        return tx.Save(user).Error
    })
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
    return r.data.WithGormTx(ctx, func(ctx context.Context, tx *gorm.DB) error {
        return tx.Delete(&User{}, id).Error
    })
}
*/
`, extType, name, extType, name)
}
