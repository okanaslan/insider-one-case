package model

// MetricsQueryParams holds raw query string values bound directly from the HTTP request.
type MetricsQueryParams struct {
	EventName string `form:"event_name"`
	From      string `form:"from"`
	To        string `form:"to"`
	GroupBy   string `form:"group_by"`
}

type MetricsQuery struct {
	EventName string
	From      int64
	To        int64
	GroupBy   string
}

type MetricsGroup struct {
	Key         string `json:"key"`
	Count       uint64 `json:"count"`
	UniqueUsers uint64 `json:"unique_users"`
}

type MetricsResponse struct {
	EventName   string         `json:"event_name"`
	From        int64          `json:"from"`
	To          int64          `json:"to"`
	TotalCount  uint64         `json:"total_count"`
	UniqueUsers uint64         `json:"unique_users"`
	GroupBy     string         `json:"group_by,omitempty"`
	Groups      []MetricsGroup `json:"groups,omitempty"`
}
