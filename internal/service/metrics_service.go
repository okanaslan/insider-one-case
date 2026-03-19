package service

import (
	"context"
	"log/slog"

	"insider-one-case/internal/model"
)

type MetricsRepository interface {
	QueryTotals(ctx context.Context, query model.MetricsQuery) (uint64, uint64, error)
	QueryGroupedByChannel(ctx context.Context, query model.MetricsQuery) ([]model.MetricsGroup, error)
}

type MetricsService struct {
	repo MetricsRepository
	log  *slog.Logger
}

func NewMetricsService(repo MetricsRepository, log *slog.Logger) *MetricsService {
	return &MetricsService{repo: repo, log: log}
}

func (s *MetricsService) Query(ctx context.Context, query model.MetricsQuery) (model.MetricsResponse, error) {
	totalCount, uniqueUsers, err := s.repo.QueryTotals(ctx, query)
	if err != nil {
		return model.MetricsResponse{}, err
	}

	response := model.MetricsResponse{
		EventName:   query.EventName,
		From:        query.From,
		To:          query.To,
		TotalCount:  totalCount,
		UniqueUsers: uniqueUsers,
	}

	if query.GroupBy == "channel" {
		groups, err := s.repo.QueryGroupedByChannel(ctx, query)
		if err != nil {
			return model.MetricsResponse{}, err
		}
		response.GroupBy = "channel"
		response.Groups = groups
	}

	return response, nil
}
