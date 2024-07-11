package data

import (
	"context"
	"database/sql"
	"ncobase/common/config"
	"ncobase/common/data"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"

	"ncobase/feature/auth/data/ent"
	"ncobase/feature/auth/data/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/redis/go-redis/v9"
)

// Data .
type Data struct {
	*data.Data
	EC *ent.Client
}

// New creates a new Database Connection.
func New(conf *config.Data) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	entClient, err := newEntClient(d.DB, conf.Database)
	if err != nil {
		return nil, nil, err
	}

	return &Data{
		Data: d,
		EC:   entClient,
	}, cleanup, nil
}

// newEntClient creates a new ent client.
func newEntClient(db *sql.DB, conf *config.Database) (*ent.Client, error) {
	client := ent.NewClient(ent.Driver(dialect.DebugWithContext(
		entsql.OpenDB(conf.Driver, db),
		func(ctx context.Context, i ...any) {
			log.Infof(ctx, "%v", i)
		},
	)))
	// Enable SQL logging
	if conf.Logging {
		client = client.Debug()
	}

	// Auto migrate
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

// GetEntClient get ent client
func (d *Data) GetEntClient() *ent.Client {
	return d.EC
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if d.EC != nil {
		if err := d.EC.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// GetDB get database
func (d *Data) GetDB() *sql.DB {
	return d.DB
}

// GetRedis get redis
func (d *Data) GetRedis() *redis.Client {
	return d.RC
}

// GetMeilisearch get meilisearch
func (d *Data) GetMeilisearch() *meili.Client {
	return d.MS
}

func (d *Data) GetElasticsearchClient() *elastic.Client {
	return d.ES
}

// Ping .
func (d *Data) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}

// CloseDB .
func (d *Data) CloseDB() error {
	return d.DB.Close()
}
