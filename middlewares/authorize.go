package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func Authorize(permission config.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, found := utils.GetPrincipal(c)
		if !found {
			utils.HandleCustomError(c, unauthorized(
				"authorize "+string(permission),
				config.ErrPrincipalMissing,
			))

			return
		}

		ownerCandidateID := c.Param(config.GetCandidateIDParam())
		if ownerCandidateID != "" && uuid.Validate(ownerCandidateID) != nil {
			utils.HandleCustomError(c, notAccessible(permission))

			return
		}

		scope, customErr := matchingScope(principal, permission, ownerCandidateID)
		if customErr != nil {
			utils.HandleCustomError(c, customErr)

			return
		}

		if scope == config.ScopeAny && ownerCandidateID != "" &&
			ownerCandidateID != principal.CandidateID {
			utils.LogInfo(
				c.Request.Context(),
				"privileged access to candidate data",
				"principal_kind", string(principal.Kind),
				"principal_subject", principal.Subject,
				"permission", string(permission),
				"owner_candidate_id", ownerCandidateID,
			)
		}

		c.Next()
	}
}

func matchingScope(
	principal sharedTypes.Principal,
	permission config.Permission,
	ownerCandidateID string,
) (config.AccessScope, *config.CustomError) {
	scopes := principal.Grants[permission]

	for _, scope := range scopes {
		if scopeAllows(scope, principal, ownerCandidateID) {
			return scope, nil
		}
	}

	if len(scopes) > 0 && ownerCandidateID != "" {
		return "", notAccessible(permission)
	}

	return "", config.NewCustomError(
		fmt.Errorf("authorize %s: %w", permission, config.ErrPermissionDenied),
		config.GetErrCodeForbidden(),
	)
}

func scopeAllows(
	scope config.AccessScope,
	principal sharedTypes.Principal,
	ownerCandidateID string,
) bool {
	switch scope {
	case config.ScopeAny:
		return true
	case config.ScopeOwn:
		return ownerCandidateID != "" && ownerCandidateID == principal.CandidateID
	default:
		return false
	}
}

func notAccessible(permission config.Permission) *config.CustomError {
	return config.NewCustomError(
		fmt.Errorf("authorize %s: %w", permission, config.ErrResourceNotAccessible),
		config.GetErrCodeNotFound(),
	)
}
