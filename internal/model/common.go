package model

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error" example:"invalid_request"`
	Message string `json:"message" example:"malformed JSON body"`
}

type HealthData struct {
	Status string `json:"status" example:"ok"`
	App    string `json:"app" example:"insider-one-case"`
	Env    string `json:"env" example:"development"`
	Time   string `json:"time" example:"2026-03-19T22:50:00Z"`
}

type HealthResponse struct {
	Success bool       `json:"success" example:"true"`
	Message string     `json:"message" example:"ok"`
	Data    HealthData `json:"data"`
}
