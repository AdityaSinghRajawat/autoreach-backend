package swagger

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

//go:embed openapi.json
var openAPISpec []byte

func GetOpenAPISpec() []byte {
	return append([]byte(nil), openAPISpec...)
}

func LoadSwaggerRoutes(group *gin.RouterGroup) {
	group.GET(config.GetOpenAPISpecPath(), func(c *gin.Context) {
		c.Data(http.StatusOK, config.GetJSONContentTypeUTF8(), openAPISpec)
	})
	group.GET(
		config.GetSwaggerUIPath(),
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(config.GetOpenAPISpecPath())),
	)
}
