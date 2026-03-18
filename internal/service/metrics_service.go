package service

import (
	"context"
	"log/slog"
	"time"

	"insider-one-case/internal/model"
	"insider-one-case/internal/repository"
)

type MetricsService struct {
	repo *repository.MetricsRepository
	log  *slog.Logger
}

func NewMetricsService(repo *repository.MetricsRepository, log *slog.Logger) *MetricsService {
	return &MetricsService{repo: repo, log: log}
}

func (s *MetricsService) Query(ctx context.Context, query model.MetricsQuery) (model.MetricsResponse, error) {
	points, err := s.repo.QueryMetrics(ctx, query)
	if err != nil {
		return model.MetricsResponse{}, err
	}

	if len(points) == 0 {
		// TODO: replace placeholder series once metric query implementation is complete.
		points = []model.MetricsPoint{
			{Timestamp: time.Now().UTC().Add(-1 * time.Minute), Value: 10},
			{Timestamp: time.Now().UTC(), Value: 12},
		}
	}

	return model.MetricsResponse{
		MetricName: query.MetricName,
		Points:     points,
	}, nil
}
