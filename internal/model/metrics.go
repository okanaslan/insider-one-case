package model

import "time"

type MetricsQuery struct {
	MetricName  string
	From        time.Time
	To          time.Time
	Granularity string
	Limit       int
}

type MetricsPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type MetricsResponse struct {
	MetricName string         `json:"metric_name"`
	Points     []MetricsPoint `json:"points"`
}
