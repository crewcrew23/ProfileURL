package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationPath, migrationTable string

	flag.StringVar(&storagePath, "storage", "", "path to storage")
	flag.StringVar(&migrationPath, "migration-path", "", "path to migration")
	flag.StringVar(&migrationTable, "migration-table", "migrations", "name of migration")

	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}

	if migrationPath == "" {
		panic("migration-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s&_busy_timeout=5000", storagePath, migrationTable))

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	log.Println("migrations applied successfully")
}
