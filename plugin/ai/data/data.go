package data

import (
	"context"
	"database/sql"
	"fmt"
	"ncobase/plugin/ai/data/ent"
	"ncobase/plugin/ai/data/ent/migrate"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/logging/logger"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
)

type Data struct {
	*data.Data
	EC     *ent.Client
	ECRead *ent.Client
}

func New(conf *config.Data, env ...string) (*Data, func(name ...string), error) {
	d, cleanup, err := data.New(conf)
	if err != nil {
		return nil, nil, err
	}

	masterDB := d.GetMasterDB()
	if masterDB == nil {
		return nil, cleanup, fmt.Errorf("master database connection is nil")
	}

	entClient, err := newEntClient(masterDB, conf.Database.Master, conf.Database.Migrate, env...)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create AI ent client: %v", err)
	}

	entClientRead := entClient
	if readDB, err := d.GetSlaveDB(); err == nil && readDB != nil && readDB != masterDB {
		entClientRead, err = newEntClient(readDB, conf.Database.Master, false, env...)
		if err != nil {
			logger.Warnf(nil, "Failed to create AI read-only ent client, using master: %v", err)
			entClientRead = entClient
		}
	}

	return &Data{
		Data:   d,
		EC:     entClient,
		ECRead: entClientRead,
	}, cleanup, nil
}

func newEntClient(db *sql.DB, conf *config.DBNode, enableMigrate bool, env ...string) (*ent.Client, error) {
	client := ent.NewClient(ent.Driver(dialect.DebugWithContext(
		entsql.OpenDB(conf.Driver, db),
		func(ctx context.Context, i ...any) {
			if conf.Logging {
				logger.Infof(ctx, "%v", i)
			}
		},
	)))

	if conf.Logging {
		client = client.Debug()
	}

	if enableMigrate {
		migrateOpts := []schema.MigrateOption{
			migrate.WithForeignKeys(false),
		}
		if len(env) == 0 || env[0] != "production" {
			migrateOpts = append(migrateOpts, migrate.WithDropIndex(true), migrate.WithDropColumn(true))
		}
		if err := client.Schema.Create(context.Background(), migrateOpts...); err != nil {
			return nil, fmt.Errorf("failed to migrate AI database schema: %v", err)
		}
	}

	return client, nil
}

func (d *Data) GetMasterEntClient() *ent.Client {
	return d.EC
}

func (d *Data) GetSlaveEntClient() *ent.Client {
	if d.ECRead != nil {
		return d.ECRead
	}
	return d.EC
}

func (d *Data) Close() (errs []error) {
	if d.EC != nil {
		if err := d.EC.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close AI master ent client: %v", err))
		}
	}
	if d.ECRead != nil && d.ECRead != d.EC {
		if err := d.ECRead.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close AI read ent client: %v", err))
		}
	}
	if baseErrs := d.Data.Close(); len(baseErrs) > 0 {
		errs = append(errs, baseErrs...)
	}
	return errs
}
