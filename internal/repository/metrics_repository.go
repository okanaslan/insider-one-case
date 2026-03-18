package repository

import (
	"context"
	"log/slog"

	"github.com/ClickHouse/clickhouse-go/v2"

	"insider-one-case/internal/model"
)

type MetricsRepository struct {
	conn clickhouse.Conn
	log  *slog.Logger
}

func NewMetricsRepository(conn clickhouse.Conn, log *slog.Logger) *MetricsRepository {
	return &MetricsRepository{conn: conn, log: log}
}

func (r *MetricsRepository) QueryMetrics(ctx context.Context, query model.MetricsQuery) ([]model.MetricsPoint, error) {
	if r.conn == nil {
		r.log.Debug("clickhouse not configured, returning placeholder metrics", "metric", query.MetricName)
		return nil, nil
	}

	_ = ctx
	_ = query
	// TODO: implement metrics SQL query against ClickHouse.
	return nil, nil
}
