package db

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB, migrationsPath *string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	if migrationsPath == nil {
		return goose.Up(db, "./migrations")
	}

	return goose.Up(db, *migrationsPath)
}
