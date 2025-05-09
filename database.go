package main

import (
	"context"
	"errors"
	"maps"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/do/v2"
	pgxuuid "github.com/vgarvardt/pgx-google-uuid/v5"

	"github.com/dabbertorres/notes/internal/config"
	"github.com/dabbertorres/notes/internal/util"
)

type databasePool struct {
	*pgxpool.Pool
}

func setupDatabase(injector do.Injector) (databasePool, error) {
	ctx := do.MustInvoke[context.Context](injector)
	cfg := do.MustInvoke[*config.Config](injector)

	// get a default config
	dbCfg, err := pgxpool.ParseConfig("")
	if err != nil {
		return databasePool{}, err
	}

	dbCfg.ConnConfig.Host = cfg.Database.Host
	dbCfg.ConnConfig.Port = cfg.Database.Port
	dbCfg.ConnConfig.User = cfg.Database.User
	dbCfg.ConnConfig.Password = cfg.Database.Pass
	maps.Copy(dbCfg.ConnConfig.RuntimeParams, cfg.Database.Args)
	dbCfg.MaxConnLifetime = cfg.Database.MaxConnLifetime.Value
	dbCfg.MaxConnLifetimeJitter = cfg.Database.MaxConnLifetimeJitter.Value
	dbCfg.MaxConnIdleTime = cfg.Database.MaxConnIdleTime.Value
	dbCfg.MaxConns = int32(cfg.Database.MaxConns)
	dbCfg.MinConns = int32(cfg.Database.MinConns)
	dbCfg.HealthCheckPeriod = cfg.Database.HealthCheckPeriod.Value

	dbCfg.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		pgxuuid.Register(c.TypeMap())
		return nil
	}

	if cfg.Database.LogConnections {
		dbCfg.BeforeConnect = func(ctx context.Context, cc *pgx.ConnConfig) error {
			// TODO
			return nil
		}

		dbCfg.AfterConnect = util.AllUntilError2(dbCfg.AfterConnect, func(ctx context.Context, c *pgx.Conn) error {
			// TODO
			return nil
		})

		dbCfg.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
			// TODO
			return true
		}

		dbCfg.AfterRelease = func(c *pgx.Conn) bool {
			// TODO
			return true
		}

		dbCfg.BeforeClose = func(c *pgx.Conn) {
			// TODO
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		return databasePool{}, err
	}

	return databasePool{Pool: pool}, nil
}

func (d databasePool) HealthCheck() error {
	var errs []error

	ctx := context.Background()

	idle := d.AcquireAllIdle(ctx)
	for _, c := range idle {
		if err := c.Ping(ctx); err != nil {
			errs = append(errs, err)
		}
		c.Release()
	}

	return errors.Join(errs...)
}

func (d databasePool) Shutdown() error {
	d.Close()
	return nil
}
