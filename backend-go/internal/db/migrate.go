package db

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type MigrationDirection string

const (
	MigrationUp   MigrationDirection = "up"
	MigrationDown MigrationDirection = "down"
)

var ErrNoMigrationChange = migrate.ErrNoChange

func RunMigrations(databaseURL string, direction MigrationDirection) error {
	return RunMigrationsFromPath(databaseURL, filepath.Join("internal", "db", "migrations"), direction)
}

func RunMigrationsFromPath(databaseURL, migrationsPath string, direction MigrationDirection) error {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	sourceURL, err := migrationsSourceURL(migrationsPath)
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	switch direction {
	case MigrationUp:
		err = migrator.Up()
	case MigrationDown:
		err = migrator.Down()
	default:
		return fmt.Errorf("unknown migration direction %q", direction)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		return ErrNoMigrationChange
	}
	if err != nil {
		return fmt.Errorf("run %s migrations: %w", direction, err)
	}

	return nil
}

func migrationsSourceURL(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve migrations path: %w", err)
	}

	return "file://" + filepath.ToSlash(absPath), nil
}
