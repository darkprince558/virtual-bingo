package main

import (
	"errors"
	"log"
	"os"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/db"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: go run ./cmd/migrate <up|down>")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	var err error
	switch os.Args[1] {
	case "up":
		err = db.RunMigrations(databaseURL, db.MigrationUp)
	case "down":
		err = db.RunMigrations(databaseURL, db.MigrationDown)
	default:
		log.Fatalf("unknown migration command %q; expected up or down", os.Args[1])
	}

	if err == nil {
		log.Printf("migration %s complete", os.Args[1])
		return
	}

	if errors.Is(err, db.ErrNoMigrationChange) {
		log.Println("no migration changes to apply")
		return
	}

	log.Fatalf("run migrations: %v", err)
}
