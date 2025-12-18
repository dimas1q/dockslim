package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect opens a PostgreSQL connection and verifies connectivity.
func Connect(ctx context.Context, dsn string) (*sql.DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	contextWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := database.PingContext(contextWithTimeout); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}
