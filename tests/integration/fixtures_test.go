//go:build integration

package integration

import (
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

func adminPrincipal() sharedTypes.Principal {
	return sharedTypes.Principal{
		UserID: "11111111-1111-4111-8111-111111111111",
		Role:   modelHelpers.UserRoleAdmin,
	}
}
