package _postgres

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"user_api/pkg/modules"
)

type PGXDialect struct {
	DB *sqlx.DB
}

func NewPGXDialect(ctx context.Context, cfg *modules.PostgreConfig) *PGXDialect {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	fmt.Println("Connecting with DSN:", dsn)

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		panic(err)
	}

	return &PGXDialect{DB: db}
}

func AutoMigrate(cfg *modules.PostgreConfig) {
	sourceURL := "file://database/migrations"
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		panic(err)
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
}
