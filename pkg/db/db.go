package db

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/skr1ms/PaymentAlphaBank.git/config"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type Db struct {
	*bun.DB
}

func NewDb(config *config.Config) (*Db, error) {
	db, err := pgx.ParseConfig(config.DbConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Postgres config: %w", err)
	}
	db.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	psqlDB := stdlib.OpenDB(*db)
	return &Db{bun.NewDB(psqlDB, pgdialect.New())}, nil
}
