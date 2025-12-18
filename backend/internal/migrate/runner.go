package migrate

import (
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

// Close closes the underlying database connection.
func (r *Runner) Close() error {
	return r.db.Close()
}
