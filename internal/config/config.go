package config

import (
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	DefaultWorkerBatchSize        = 250
	MinWorkerBatchSize            = 1
	MaxWorkerBatchSize            = 1000
	DefaultWorkerFlushIntervalMS  = 250
	MinWorkerFlushIntervalMS      = 10
	MaxWorkerFlushIntervalMS      = 5000
	DefaultIngestQueueBufferSize  = 5000
	MinIngestQueueBufferSize      = 1
	MaxIngestQueueBufferSize      = 20000
	DefaultIngestEnqueueTimeoutMS = 25
	MinIngestEnqueueTimeoutMS     = 1
	MaxIngestEnqueueTimeoutMS     = 250
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
	WorkerBatchSize        int    `env:"WORKER_BATCH_SIZE" envDefault:"250"`
	WorkerFlushIntervalMS  int    `env:"WORKER_FLUSH_INTERVAL_MS" envDefault:"250"`
	IngestQueueBufferSize  int    `env:"INGEST_QUEUE_BUFFER_SIZE" envDefault:"5000"`
	IngestEnqueueTimeoutMS int    `env:"INGEST_ENQUEUE_TIMEOUT_MS" envDefault:"25"`
}

func (c Config) Normalized() Config {
	c.WorkerBatchSize = clampIntOrDefault(c.WorkerBatchSize, DefaultWorkerBatchSize, MinWorkerBatchSize, MaxWorkerBatchSize)
	c.WorkerFlushIntervalMS = clampIntOrDefault(c.WorkerFlushIntervalMS, DefaultWorkerFlushIntervalMS, MinWorkerFlushIntervalMS, MaxWorkerFlushIntervalMS)
	c.IngestQueueBufferSize = clampIntOrDefault(c.IngestQueueBufferSize, DefaultIngestQueueBufferSize, MinIngestQueueBufferSize, MaxIngestQueueBufferSize)
	c.IngestEnqueueTimeoutMS = clampIntOrDefault(c.IngestEnqueueTimeoutMS, DefaultIngestEnqueueTimeoutMS, MinIngestEnqueueTimeoutMS, MaxIngestEnqueueTimeoutMS)
	return c
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, err
	}

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg.Normalized(), nil
}

func clampIntOrDefault(value, defaultValue, minValue, maxValue int) int {
	if value <= 0 {
		return defaultValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
