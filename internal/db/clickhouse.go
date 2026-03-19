package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"insider-one-case/internal/config"
)

func NewClickHouseConn(ctx context.Context, cfg config.Config) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(clickHouseOptions(cfg))
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}

func NewClickHouseSQLDB(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	db := clickhouse.OpenDB(clickHouseOptions(cfg))
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func clickHouseOptions(cfg config.Config) *clickhouse.Options {
	return &clickhouse.Options{
		Addr: []string{cfg.ClickHouseAddr},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouseDatabase,
			Username: cfg.ClickHouseUsername,
			Password: cfg.ClickHousePassword,
		},
		DialTimeout: 5 * time.Second,
	}
}
