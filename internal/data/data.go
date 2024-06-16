package data

import (
	"context"
	"database/sql"
	"io"
	"ncobase/common/config"
	"ncobase/common/elastic"
	"ncobase/common/log"
	"ncobase/common/meili"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/ent/migrate"

	"github.com/redis/go-redis/v9"

	entsql "entgo.io/ent/dialect/sql"

	// sqlite3
	_ "github.com/mattn/go-sqlite3"
	// mysql
	_ "github.com/go-sql-driver/mysql"
	// postgres
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	err error
	db  *sql.DB
	ec  *ent.Client
	rc  *redis.Client
	ms  *meili.Client
	es  *elastic.Client
)

// Data .
type Data struct {
	db *sql.DB
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	es *elastic.Client
}

// New creates a new Database Connection.
func New(conf *config.Data) (*Data, func(), error) {
	ec, db = newClient(&conf.Database)
	es = newElasticsearch(&conf.Elasticsearch)
	d := &Data{
		db: db,
		ec: ec,
		rc: newRedis(&conf.Redis),
		ms: newMeilisearch(&conf.Meilisearch),
		es: es,
	}

	cleanup := func() {
		log.Printf(context.Background(), "execute data cleanup.")
		if errs := d.Close(); errs != nil {
			log.Fatalf(context.Background(), "cleanup errors: %v", errs)
		}
	}

	return d, cleanup, nil
}

// newRedis creates a new Redis datastore.
func newRedis(conf *config.Redis) *redis.Client {
	if conf == nil || conf.Addr == "" {
		log.Printf(context.Background(), "redis configuration is nil or empty")
		return nil
	}

	// Create client
	rc = redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Username:     conf.Username,
		Password:     conf.Password,
		DB:           conf.Db,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		DialTimeout:  conf.DialTimeout,
		PoolSize:     10,
	})

	timeout, cancelFunc := context.WithTimeout(context.Background(), conf.DialTimeout)
	defer cancelFunc()
	if err := rc.Ping(timeout).Err(); err != nil {
		log.Errorf(context.Background(), "redis connect error: %v", err)
	}

	return rc
}

// GetRedis returns the redis client
func (d *Data) GetRedis() *redis.Client {
	return d.rc
}

// newClient creates a new ent client.
func newClient(conf *config.Database) (*ent.Client, *sql.DB) {
	// Open database
	switch conf.Driver {
	case "postgres":
		db, err = sql.Open("pgx", conf.Source)
	case "mysql":
		db, err = sql.Open("mysql", conf.Source)
	case "sqlite3":
		db, err = sql.Open("sqlite3", conf.Source)
	default:
		log.Fatalf(context.Background(), "Dialect %v not supported.", conf.Driver)
	}

	if err != nil {
		log.Fatalf(context.Background(), "Failed to connect to database %v", err)
		return nil, nil
	}

	// Idle Connection
	db.SetMaxIdleConns(conf.MaxIdleConn)
	// Max Open Connection
	db.SetMaxOpenConns(conf.MaxOpenConn)
	// Max Connect Lifetime
	db.SetConnMaxLifetime(conf.ConnMaxLifeTime)

	// Create ent client
	client := ent.NewClient(ent.Driver(entsql.OpenDB(conf.Driver, db)))

	// Enable SQL logging
	if conf.Logging {
		client = client.Debug()
	}

	// Auto migrate
	if conf.Migrate {
		if err = client.Schema.Create(context.Background(), migrate.WithForeignKeys(false), migrate.WithDropIndex(true), migrate.WithDropColumn(true)); err != nil {
			log.Fatalf(context.Background(), "data.client.Schema.Create error: %v", err)
		}
	}

	return client, db
}

// GetEntClient returns the ent client
func (d *Data) GetEntClient() *ent.Client {
	return d.ec
}

// GetDB returns the database
func (d *Data) GetDB() *sql.DB {
	return d.db
}

// newMeilisearch creates a new Meilisearch client.
func newMeilisearch(conf *config.Meilisearch) *meili.Client {
	if conf == nil || conf.Host == "" {
		log.Printf(context.Background(), "Meilisearch configuration is nil or empty")
		return nil
	}

	// Create client
	ms = meili.NewMeilisearch(conf.Host, conf.APIKey)

	// Check connection
	_, err := ms.GetClient().Health()
	if err != nil {
		log.Errorf(context.Background(), "Meilisearch connect error: %v", err)
		return nil
	}

	return ms
}

// GetMeilisearch returns the Meilisearch client
func (d *Data) GetMeilisearch() *meili.Client {
	return d.ms
}

func newElasticsearch(conf *config.Elasticsearch) *elastic.Client {
	if conf == nil || len(conf.Addresses) == 0 {
		log.Printf(context.Background(), "Elasticsearch configuration is nil or empty")
		return nil
	}

	// Create client
	es, err := elastic.NewClient(conf.Addresses, conf.Username, conf.Password)
	if err != nil {
		log.Errorf(context.Background(), "Elasticsearch client creation error: %v", err)
		return nil
	}

	// Check connection
	res, err := es.GetClient().Info()
	if err != nil {
		log.Errorf(context.Background(), "Elasticsearch connect error: %v", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf(context.Background(), "Elasticsearch response body close error: %v", err)
		}
	}(res.Body)

	if res.IsError() {
		log.Errorf(context.Background(), "Elasticsearch info error: %s", res.Status())
		return nil
	}

	return es
}

func (d *Data) GetElasticsearchClient() *elastic.Client {
	return d.es
}

// Close closes all the resources in Data and returns any errors encountered.
func (d *Data) Close() (errs []error) {
	if d.rc != nil {
		if err := d.rc.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if d.db != nil {
		if err := d.db.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if d.ec != nil {
		if err := d.ec.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Ping .
func (d *Data) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
