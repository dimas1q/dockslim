package migrate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Runner struct {
	instance *migrate.Migrate
	db       *sql.DB
}

// NewRunner constructs a migration runner using the provided DSN and migrations directory.
func NewRunner(dsn, migrationsPath string) (*Runner, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, err
	}

	sourceURL := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Runner{instance: m, db: db}, nil
}

// Up applies all available migrations.
func (r *Runner) Up() error {
	if err := r.instance.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

// UpWithLock acquires a PostgreSQL advisory lock to prevent concurrent migration runs.
func (r *Runner) UpWithLock(ctx context.Context, lockName string) error {
	if lockName == "" {
		lockName = "dockslim_migrations"
	}

	lockSQL := "SELECT pg_advisory_lock(hashtext($1))"
	unlockSQL := "SELECT pg_advisory_unlock(hashtext($1))"

	conn, err := r.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("opening migration lock connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, lockSQL, lockName); err != nil {
		return fmt.Errorf("acquiring migration lock: %w", err)
	}
	defer conn.ExecContext(context.Background(), unlockSQL, lockName)

	return r.Up()
}

// Close closes the underlying database connection.
func (r *Runner) Close() error {
	return r.db.Close()
}
