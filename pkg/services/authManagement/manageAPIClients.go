package authManagement

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"

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

func (s *AuthService) ListAPIClients(
	ctx context.Context,
) ([]*authModel.APIClient, *config.CustomError) {
	clients, err := s.apiClientRepo.List(ctx)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list api clients: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return clients, nil
}

func (s *AuthService) RevokeAPIClient(ctx context.Context, clientID string) *config.CustomError {
	if _, err := uuid.Parse(clientID); err != nil {
		return config.NewCustomError(
			fmt.Errorf("revoke api client: %w", config.ErrAPIClientNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	revoked, err := s.apiClientRepo.MarkRevoked(ctx, clientID, utils.GetCurrTimeStamp())
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("revoke api client: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !revoked {
		return config.NewCustomError(
			fmt.Errorf("revoke api client: %w", config.ErrAPIClientNotFound),
			config.GetErrCodeNotFound(),
		)
	}

	return nil
}

func (s *AuthService) VerifyAPIKey(
	ctx context.Context,
	apiKey, requiredScope string,
) *config.CustomError {
	parts := strings.SplitN(apiKey, config.GetAPIKeySeparator(), 3)
	if len(parts) != 3 || parts[0] != config.GetAPIKeyPrefix() {
		return config.NewCustomError(
			fmt.Errorf("verify api key: %w", config.ErrAPIKeyInvalid),
			config.GetErrCodeUnauthorized(),
		)
	}

	client, err := s.apiClientRepo.FindActiveByPrefix(
		ctx,
		parts[0]+config.GetAPIKeySeparator()+parts[1],
	)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("verify api key: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if client == nil || !matchesKeyHash(client, apiKey) {
		return config.NewCustomError(
			fmt.Errorf("verify api key: %w", config.ErrAPIKeyInvalid),
			config.GetErrCodeUnauthorized(),
		)
	}

	if !client.HasScope(requiredScope) {
		return config.NewCustomError(
			fmt.Errorf("verify api key: %w", config.ErrAPIKeyScopeMissing),
			config.GetErrCodeForbidden(),
		)
	}

	usedAt := utils.GetCurrTimeStamp()

	utils.RunInBackground(ctx, "update api client last used", func(ctx context.Context) {
		if err := s.apiClientRepo.UpdateLastUsedAt(ctx, client.ID, usedAt); err != nil {
			utils.LogWarn(ctx, "api client last-used update failed", "error", err.Error())
		}
	})

	return nil
}

func matchesKeyHash(client *authModel.APIClient, apiKey string) bool {
	return subtle.ConstantTimeCompare(
		[]byte(client.KeyHash),
		[]byte(utils.Fingerprint(apiKey)),
	) == 1
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
