package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"insider-one-case/internal/model"
)

type fakeMetricsRepo struct {
	totalCount   uint64
	uniqueUsers  uint64
	groups       []model.MetricsGroup
	totalsErr    error
	groupedErr   error
	groupedCalls int
}

func (f *fakeMetricsRepo) QueryTotals(ctx context.Context, query model.MetricsQuery) (uint64, uint64, error) {
	_ = ctx
	_ = query
	if f.totalsErr != nil {
		return 0, 0, f.totalsErr
	}
	return f.totalCount, f.uniqueUsers, nil
}

func (f *fakeMetricsRepo) QueryGroupedByChannel(ctx context.Context, query model.MetricsQuery) ([]model.MetricsGroup, error) {
	_ = ctx
	_ = query
	f.groupedCalls++
	if f.groupedErr != nil {
		return nil, f.groupedErr
	}
	return f.groups, nil
}

func TestMetricsServiceQueryTotalsOnly(t *testing.T) {
	repo := &fakeMetricsRepo{totalCount: 12, uniqueUsers: 5}
	svc := &MetricsService{repo: repo}

	resp, err := svc.Query(context.Background(), model.MetricsQuery{EventName: "purchase", From: 1, To: 2})
	require.NoError(t, err)
	require.Equal(t, uint64(12), resp.TotalCount)
	require.Equal(t, uint64(5), resp.UniqueUsers)
	require.Empty(t, resp.GroupBy)
	require.Empty(t, resp.Groups)
	require.Equal(t, 0, repo.groupedCalls)
}

func TestMetricsServiceQueryGroupedByChannel(t *testing.T) {
	repo := &fakeMetricsRepo{
		totalCount:  20,
		uniqueUsers: 11,
		groups: []model.MetricsGroup{
			{Key: "mobile", Count: 13, UniqueUsers: 7},
			{Key: "web", Count: 7, UniqueUsers: 4},
		},
	}
	svc := &MetricsService{repo: repo}

	resp, err := svc.Query(context.Background(), model.MetricsQuery{EventName: "purchase", From: 1, To: 2, GroupBy: "channel"})
	require.NoError(t, err)
	require.Equal(t, "channel", resp.GroupBy)
	require.Len(t, resp.Groups, 2)
	require.Equal(t, 1, repo.groupedCalls)
}

func TestMetricsServiceReturnsErrorWhenGroupedQueryFails(t *testing.T) {
	repo := &fakeMetricsRepo{groupedErr: errors.New("query failed")}
	svc := &MetricsService{repo: repo}

	_, err := svc.Query(context.Background(), model.MetricsQuery{EventName: "purchase", From: 1, To: 2, GroupBy: "channel"})
	require.Error(t, err)
}
