package shared

import "github.com/AdityaSinghRajawat/autoreach-backend/config"

type ErrorBody struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Details []config.ErrorDetail `json:"details,omitempty"`
}

type ServerResponse struct {
	Success    bool            `json:"success"`
	StatusCode int             `json:"statusCode"`
	Data       any             `json:"data,omitempty"`
	Meta       *PaginationMeta `json:"meta,omitempty"`
	Error      *ErrorBody      `json:"error,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
