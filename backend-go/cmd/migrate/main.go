package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: go run ./cmd/migrate <up|down>")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("create postgres migration driver: %v", err)
	}

	sourceURL, err := migrationsSourceURL()
	if err != nil {
		log.Fatalf("resolve migrations path: %v", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		log.Fatalf("create migrator: %v", err)
	}

	switch os.Args[1] {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Down()
	default:
		log.Fatalf("unknown migration command %q; expected up or down", os.Args[1])
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("no migration changes to apply")
		return
	}
	if err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Printf("migration %s complete", os.Args[1])
}

func migrationsSourceURL() (string, error) {
	path := filepath.Join("internal", "db", "migrations")
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	return "file://" + filepath.ToSlash(absPath), nil
}
