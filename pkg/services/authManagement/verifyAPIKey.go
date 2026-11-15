package authManagement

import (
	"context"
	"crypto/subtle"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	authModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/authManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *AuthService) VerifyAPIKey(
	ctx context.Context,
	apiKey string,
) (*sharedTypes.Principal, *config.CustomError) {
	parts := strings.SplitN(apiKey, config.GetAPIKeySeparator(), 3)
	if len(parts) != 3 || parts[0] != config.GetAPIKeyPrefix() {
		return nil, config.NewCustomError(
			fmt.Errorf("verify api key: %w", config.ErrAPIKeyInvalid),
			config.GetErrCodeUnauthorized(),
		)
	}

	client, err := s.apiClientRepo.FindActiveByPrefix(
		ctx,
		parts[0]+config.GetAPIKeySeparator()+parts[1],
	)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("verify api key: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if client == nil || !matchesKeyHash(client, apiKey) {
		return nil, config.NewCustomError(
			fmt.Errorf("verify api key: %w", config.ErrAPIKeyInvalid),
			config.GetErrCodeUnauthorized(),
		)
	}

	usedAt := utils.GetCurrTimeStamp()

	utils.RunInBackground(ctx, "update api client last used", func(ctx context.Context) {
		if err := s.apiClientRepo.UpdateLastUsedAt(ctx, client.ID, usedAt); err != nil {
			utils.LogWarn(ctx, "api client last-used update failed", "error", err.Error())
		}
	})

	principal := sharedTypes.NewAPIClientPrincipal(client.ID, client.Scopes)

	return &principal, nil
}

func matchesKeyHash(client *authModel.APIClient, apiKey string) bool {
	return subtle.ConstantTimeCompare(
		[]byte(client.KeyHash),
		[]byte(utils.Fingerprint(apiKey)),
	) == 1
}
