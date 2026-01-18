package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect opens a PostgreSQL connection and verifies connectivity.
func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	return ConnectWithRetry(ctx, dsn, 20, 500*time.Millisecond)
}

// ConnectWithRetry opens a PostgreSQL connection and retries until it succeeds or attempts are exhausted.
func ConnectWithRetry(ctx context.Context, dsn string, attempts int, delay time.Duration) (*sql.DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		lastErr = database.PingContext(pingCtx)
		cancel()
		if lastErr == nil {
			return database, nil
		}
		if i == attempts-1 {
			break
		}
		time.Sleep(delay)
	}

	database.Close()
	return nil, lastErr
}
