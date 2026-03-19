package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppName                string `env:"APP_NAME" envDefault:"insider-one-case"`
	AppEnv                 string `env:"APP_ENV" envDefault:"development"`
	HTTPPort               string `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel               string `env:"LOG_LEVEL" envDefault:"info"`
	ClickHouseAddr         string `env:"CLICKHOUSE_ADDR" envDefault:"localhost:9000"`
	ClickHouseDatabase     string `env:"CLICKHOUSE_DATABASE" envDefault:"analytics"`
	ClickHouseUsername     string `env:"CLICKHOUSE_USERNAME" envDefault:"default"`
	ClickHousePassword     string `env:"CLICKHOUSE_PASSWORD" envDefault:""`
	RedisAddr              string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword          string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB                int    `env:"REDIS_DB" envDefault:"0"`
	AllowStartWithoutInfra bool   `env:"ALLOW_START_WITHOUT_INFRA" envDefault:"true"`
	WorkerBatchSize        int    `env:"WORKER_BATCH_SIZE" envDefault:"100"`
	WorkerFlushIntervalMS  int    `env:"WORKER_FLUSH_INTERVAL_MS" envDefault:"1000"`
	IngestQueueBufferSize  int    `env:"INGEST_QUEUE_BUFFER_SIZE" envDefault:"10000"`
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}

	var cfg Config
	err := env.Parse(&cfg)
	return cfg, err
}
