package data

import (
	"context"
	"database/sql"
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
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Data contains the shared resources and clients.
type Data struct {
	*data.Data
	EntClient  *ent.Client
	GormClient *gorm.DB
}

// New creates a new Data instance with database connections.
func New(conf *config.Data) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	entClient, err := newEntClient(d.DB, conf.Database)
	if err != nil {
		return nil, nil, err
	}

	gormClient, err := newGormClient(d.DB, conf.Database)
	if err != nil {
		return nil, nil, err
	}

	return &Data{
		Data:       d,
		EntClient:  entClient,
		GormClient: gormClient,
	}, cleanup, nil
}

// newEntClient creates a new ent client.
func newEntClient(db *sql.DB, conf *config.Database) (*ent.Client, error) {
	client := ent.NewClient(ent.Driver(dialect.DebugWithContext(
		entsql.OpenDB(conf.Driver, db),
		func(ctx context.Context, i ...interface{}) {
			if conf.Logging {
				log.Infof(ctx, "%v", i)
			}
		},
	)))
	if conf.Logging {
		client = client.Debug()
	}

	if conf.Migrate {
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
func newGormClient(db *sql.DB, conf *config.Database) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch conf.Driver {
	case "postgres":
		dialector = postgres.Open(conf.Source)
	case "mysql":
		dialector = mysql.Open(conf.Source)
	case "sqlite3":
		dialector = sqlite.Open(conf.Source)
	default:
		log.Fatalf(context.Background(), "Dialect %v not supported", conf.Driver)
		return nil, nil
	}

	client, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if conf.Logging {
		client = client.Debug()
	}

	if conf.Migrate {
		// Example: err = client.AutoMigrate(&YourModel{})
		// if err != nil {
		// 	return nil, err
		// }
	}

	return client, nil
}

// GetEntClient returns the ent client.
func (d *Data) GetEntClient() *ent.Client {
	return d.EntClient
}

// GetGormClient returns the GORM client.
func (d *Data) GetGormClient() *gorm.DB {
	return d.GormClient
}

// GetDB returns the SQL database instance.
func (d *Data) GetDB() *sql.DB {
	return d.DB
}

// GetRedis returns the Redis client.
func (d *Data) GetRedis() *redis.Client {
	return d.RC
}

// GetMeilisearch returns the Meilisearch client.
func (d *Data) GetMeilisearch() *meili.Client {
	return d.MS
}

// GetElasticsearchClient returns the Elasticsearch client.
func (d *Data) GetElasticsearchClient() *elastic.Client {
	return d.ES
}

// Ping checks the database connection.
func (d *Data) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if d.EntClient != nil {
		if err := d.EntClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// CloseDB closes the SQL database connection.
func (d *Data) CloseDB() error {
	return d.DB.Close()
}
