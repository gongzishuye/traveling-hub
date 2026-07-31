package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/faria/traveling-hub/internal/agent"
	"github.com/faria/traveling-hub/internal/event"
	"github.com/faria/traveling-hub/internal/frog"
	"github.com/faria/traveling-hub/internal/platform/config"
	postgresplatform "github.com/faria/traveling-hub/internal/platform/postgres"
	redisplatform "github.com/faria/traveling-hub/internal/platform/redis"
	"github.com/faria/traveling-hub/internal/simulation"
	"github.com/faria/traveling-hub/internal/world"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Config     config.Config
	DB         *sql.DB
	Redis      *redis.Client
	Agents     agent.Service
	Frogs      frog.Service
	World      world.Service
	Events     event.Service
	Simulation simulation.Service
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := postgresplatform.Open(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}
	if cfg.AutoMigrate {
		if err := postgresplatform.ApplyMigrations(ctx, db, "migrations"); err != nil {
			db.Close()
			return nil, err
		}
	}
	rdb, err := redisplatform.Open(ctx, cfg.RedisAddr)
	if err != nil {
		db.Close()
		return nil, err
	}
	agents := agent.NewService(db)
	frogs := frog.NewService(db)
	worldService := world.NewService(db, world.SystemClock())
	events := event.NewService(db)
	return &App{
		Config: cfg, DB: db, Redis: rdb, Agents: agents, Frogs: frogs,
		World: worldService, Events: events, Simulation: simulation.NewService(worldService, events),
	}, nil
}

func (a *App) Close() error {
	var first error
	if err := a.Redis.Close(); err != nil {
		first = err
	}
	if err := a.DB.Close(); err != nil && first == nil {
		first = err
	}
	return first
}

func (a *App) Healthy(ctx context.Context) error {
	if err := a.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	if err := a.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	return nil
}
