package utils

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func SetPrincipal(c *gin.Context, principal sharedTypes.Principal) {
	c.Set(config.GetPrincipalContextKey(), principal)
}

func GetPrincipal(c *gin.Context) (sharedTypes.Principal, bool) {
	value, found := c.Get(config.GetPrincipalContextKey())
	if !found {
		return sharedTypes.Principal{}, false
	}

	principal, isPrincipal := value.(sharedTypes.Principal)

	return principal, isPrincipal
}
