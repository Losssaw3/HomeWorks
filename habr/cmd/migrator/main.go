package main

import (
	// Не забудь добавить, в твоем коде его не было, но ты его используешь
	"errors"
	"flag"
	"fmt"

	// Меняем драйвер на postgres
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dbURL, migrationsPath, migrationsTable string
	flag.StringVar(&dbURL, "db-url", "", "Postgres connection URL")

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations folder")

	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")

	flag.Parse()

	if dbURL == "" {
		panic("path to db is empty")
	}

	if migrationsPath == "" {
		panic("path to migration is empty")
	}

	connectionString := fmt.Sprintf("%s&x-migrations-table=%s", dbURL, migrationsTable)

	m, err := migrate.New(
		"file://"+migrationsPath,
		connectionString,
	)

	if err != nil {
		panic("failed to create migrator")
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply (database is up to date)")
			return
		}
		panic(err)
	}
}
