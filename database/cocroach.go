package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseMethods interface {
	NewPgxPool(ctx context.Context, maxRetries int) (*pgxpool.Pool, error)
	Ping(ctx context.Context, pgxPool *pgxpool.Pool, maxRetries int) error
}

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
	SSLMode  string
	MaxConn  int
}

// NewPgxPool establishes a connection pool with the database.
func (c *DbConfig) NewPgxPool(ctx context.Context, maxRetries int) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.DbName, c.SSLMode)
	slog.Info("Database connection string initialized")

	var (
		pool *pgxpool.Pool
		err  error
	)

	for i := 0; i < maxRetries; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		config, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse database config: %w", err)
		}

		config.MaxConns = int32(c.MaxConn)
		config.MinConns = 1

		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			slog.Info("Successfully connected to the database")
			return pool, nil
		}

		slog.Warn(
			"Database connection failed, retrying...",
			"attempt", i+1,
			"remainingRetries", maxRetries-i-1,
			"error", err,
		)
		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to connect to the database after %d retries: %w", maxRetries, err)
}

// Ping verifies the database connection by pinging it.
func (c *DbConfig) Ping(ctx context.Context, pgxPool *pgxpool.Pool, maxRetries int) error {
	var err error

	for i := 0; i < maxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = pgxPool.Ping(ctx)
		if err == nil {
			slog.Info("Successfully pinged the database")
			return nil
		}

		slog.Warn(
			"Database ping failed, retrying...",
			"attempt", i+1,
			"remainingRetries", maxRetries-i-1,
			"error", err,
		)
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("failed to ping the database after %d retries: %w", maxRetries, err)
}
