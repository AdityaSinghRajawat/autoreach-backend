package swagger

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:embed openapi.json
var openAPISpec []byte

const specPath = "/openapi.json"

func GetOpenAPISpec() []byte {
	return append([]byte(nil), openAPISpec...)
}

func LoadSwaggerRoutes(group *gin.RouterGroup) {
	group.GET(specPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", openAPISpec)
	})
	group.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(specPath)),
	)
}
