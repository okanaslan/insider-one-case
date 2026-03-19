package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pressly/goose/v3"

	"insider-one-case/internal/config"
	"insider-one-case/internal/db"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sqlDB, err := db.NewClickHouseSQLDB(ctx, cfg)
	if err != nil {
		slog.Error("failed to connect clickhouse for migrations", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(migrationFiles)
	if err := goose.SetDialect("clickhouse"); err != nil {
		slog.Error("failed to set goose dialect", "error", err)
		os.Exit(1)
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
			slog.Error("failed to apply migrations", "error", err)
			os.Exit(1)
		}
		fmt.Println("migrations applied")
	case "status":
		if err := goose.StatusContext(ctx, sqlDB, "migrations"); err != nil {
			slog.Error("failed to get migration status", "error", err)
			os.Exit(1)
		}
	default:
		slog.Error("unsupported migration command", "command", command)
		os.Exit(1)
	}
}
