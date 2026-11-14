package shared

import (
	"slices"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type Principal struct {
	Kind        config.PrincipalKind
	Subject     string
	CandidateID string
	Roles       []string
	Grants      config.ScopesByGrant
}

func NewUserPrincipal(userID, candidateID string, roles []string) Principal {
	return Principal{
		Kind:        config.PrincipalKindUser,
		Subject:     userID,
		CandidateID: candidateID,
		Roles:       slices.Clone(roles),
		Grants:      config.ResolveRoleGrants(roles),
	}
}

func NewAPIClientPrincipal(clientID string, grants []string) Principal {
	return Principal{
		Kind:    config.PrincipalKindAPIClient,
		Subject: clientID,
		Grants:  config.ResolveAPIClientGrants(grants),
	}
}
