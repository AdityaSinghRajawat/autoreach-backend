package authManagement

import (
	"strings"

	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
)

type CreateAPIClientRequest struct {
	Name   string   `json:"name"   validate:"required,max=100"`
	Scopes []string `json:"scopes" validate:"required,min=1,dive,oneof=jobs:ingest"`
}

type CreateAPIClientResponse struct {
	Client *authModel.APIClient `json:"client"`
	APIKey string               `json:"apiKey"`
}

func (r *CreateAPIClientRequest) ToAPIClientModel() *authModel.APIClient {
	client := authModel.NewAPIClient()
	client.Name = strings.TrimSpace(r.Name)
	client.Scopes = r.Scopes

	return client
}
