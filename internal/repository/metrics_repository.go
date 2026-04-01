package repository

import (
	"context"
	"fmt"
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

func (r *MetricsRepository) QueryTotals(ctx context.Context, query model.MetricsQuery) (uint64, uint64, error) {
	if r.conn == nil {
		r.log.Debug("clickhouse not configured, returning empty metrics totals", "event_name", query.EventName)
		return 0, 0, nil
	}

	rows, err := r.conn.Query(
		ctx,
		`SELECT count() AS total_count, uniqExact(user_id) AS unique_users
		 FROM events
		 WHERE event_name = ? AND timestamp >= ? AND timestamp < ?`,
		query.EventName,
		query.From,
		query.To,
	)
	if err != nil {
		return 0, 0, fmt.Errorf("query totals: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, 0, nil
	}

	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("iterate totals rows: %w", err)
	}

	var totalCount uint64
	var uniqueUsers uint64
	if err := rows.Scan(&totalCount, &uniqueUsers); err != nil {
		return 0, 0, fmt.Errorf("scan totals: %w", err)
	}

	return totalCount, uniqueUsers, nil
}

func (r *MetricsRepository) QueryGroupedByChannel(ctx context.Context, query model.MetricsQuery) ([]model.MetricsGroup, error) {
	if r.conn == nil {
		r.log.Debug("clickhouse not configured, returning empty grouped metrics", "event_name", query.EventName)
		return []model.MetricsGroup{}, nil
	}

	rows, err := r.conn.Query(
		ctx,
		`SELECT channel AS key, count() AS count, uniqExact(user_id) AS unique_users
		 FROM events
		 WHERE event_name = ? AND timestamp >= ? AND timestamp < ?
		 GROUP BY channel
		 ORDER BY count DESC`,
		query.EventName,
		query.From,
		query.To,
	)
	if err != nil {
		return nil, fmt.Errorf("query grouped by channel: %w", err)
	}
	defer rows.Close()

	groups := make([]model.MetricsGroup, 0)
	for rows.Next() {
		var item model.MetricsGroup
		if err := rows.Scan(&item.Key, &item.Count, &item.UniqueUsers); err != nil {
			return nil, fmt.Errorf("scan grouped row: %w", err)
		}
		groups = append(groups, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate grouped rows: %w", err)
	}

	return groups, nil
}
