package authManagement

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	authTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/authManagement"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) CreateAPIClient(
	ctx context.Context,
	client *authModel.APIClient,
) (*authTypes.CreateAPIClientResponse, *config.CustomError) {
	keyPrefix, apiKey, err := generateAPIKey()
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create api client: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	client.KeyPrefix = keyPrefix
	client.KeyHash = utils.Fingerprint(apiKey)
	client.UpdateTimestamp()

	if err := client.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create api client: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	created, err := s.apiClientRepo.Upsert(ctx, client)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create api client: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return &authTypes.CreateAPIClientResponse{Client: created, APIKey: apiKey}, nil
}

func generateAPIKey() (string, string, error) {
	prefixBytes := make([]byte, config.GetAPIKeyPrefixBytes())
	secretBytes := make([]byte, config.GetAPIKeySecretBytes())

	if _, err := rand.Read(prefixBytes); err != nil {
		return "", "", fmt.Errorf("generate api key: %w", err)
	}

	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", fmt.Errorf("generate api key: %w", err)
	}

	keyPrefix := config.GetAPIKeyPrefix() + config.GetAPIKeySeparator() + hex.EncodeToString(
		prefixBytes,
	)
	apiKey := keyPrefix + config.GetAPIKeySeparator() + base64.RawURLEncoding.EncodeToString(
		secretBytes,
	)

	return keyPrefix, apiKey, nil
}
