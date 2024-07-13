package data

import (
	"context"
	"database/sql"
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
