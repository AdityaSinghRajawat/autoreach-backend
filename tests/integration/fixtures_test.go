//go:build integration

package integration

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func adminPrincipal() sharedTypes.Principal {
	return sharedTypes.NewUserPrincipal(
		"11111111-1111-4111-8111-111111111111",
		"",
		[]string{config.RoleAdmin},
	)
}
