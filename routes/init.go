package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	swaggerDocs "github.com/AdityaSinghRajawat/autoreach-backend/docs/swagger"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/handlers"
	adminRoutes "github.com/AdityaSinghRajawat/autoreach-backend/routes/v1/admin"
	authRoutes "github.com/AdityaSinghRajawat/autoreach-backend/routes/v1/auth"
	candidateRoutes "github.com/AdityaSinghRajawat/autoreach-backend/routes/v1/candidates"
	dashboardRoutes "github.com/AdityaSinghRajawat/autoreach-backend/routes/v1/dashboard"
	jobRoutes "github.com/AdityaSinghRajawat/autoreach-backend/routes/v1/jobs"
	webhookRoutes "github.com/AdityaSinghRajawat/autoreach-backend/routes/v1/webhooks"
)

func NewRoutes(deps *dependencies.Dependencies) http.Handler {
	if !config.IsLocal() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.TrustedPlatform = config.GetTrustedPlatform()

	if err := router.SetTrustedProxies(config.GetTrustedProxies()); err != nil {
		log.Fatalf("set trusted proxies: %v", err)
	}

	router.Use(middlewares.Recovery())
	router.Use(middlewares.Metrics())
	router.Use(middlewares.RequestLogger(deps.Logger))
	router.Use(middlewares.ErrorHandler())
	router.Use(middlewares.CORS(config.GetCORSAllowedOrigins()))
	router.Use(middlewares.BodyLimit(config.GetMaxBodyBytes()))

	router.NoRoute(middlewares.NotFound())
	router.NoMethod(middlewares.MethodNotAllowed())

	router.GET("/health", handlers.Health)
	router.GET(
		config.GetTrackingOpenPath()+":"+config.GetTrackingTokenParam(),
		deps.Handlers.Outreach.TrackOpen,
	)
	router.GET(
		config.GetTrackingClickPath()+":"+config.GetTrackingTokenParam(),
		deps.Handlers.Outreach.TrackClick,
	)
	router.GET(
		config.GetUnsubscribePath()+":"+config.GetMessageIDParam(),
		deps.Handlers.Outreach.ShowUnsubscribePage,
	)
	router.POST(
		config.GetUnsubscribePath()+":"+config.GetMessageIDParam(),
		deps.Handlers.Outreach.Unsubscribe,
	)
	loadOperatorRoutes(router)
	loadAPIRoutes(router, deps)

	return router
}

func loadOperatorRoutes(router *gin.Engine) {
	operator := router.Group("")
	if config.IsProduction() {
		operator.Use(middlewares.InternalSharedSecret(config.GetInternalSharedSecret()))
	}

	operator.GET("/metrics", gin.WrapH(promhttp.Handler()))
	swaggerDocs.LoadSwaggerRoutes(operator)
}

func loadAPIRoutes(router *gin.Engine, deps *dependencies.Dependencies) {
	api := router.Group(config.GetAPIBasePath(), middlewares.RateLimit(
		deps.RateLimiter,
		config.RateLimitPolicyIP,
	))
	authRoutes.LoadAuthRoutes(api, deps)
	candidateRoutes.LoadCandidateRoutes(api, deps)
	dashboardRoutes.LoadDashboardRoutes(api, deps)
	jobRoutes.LoadJobRoutes(api, deps)
	adminRoutes.LoadAdminRoutes(api, deps)
	webhookRoutes.LoadWebhookRoutes(api, deps)
}
