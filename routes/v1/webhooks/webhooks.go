package webhooks

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadWebhookRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	handler := deps.Handlers.Enrichment

	group.POST(
		config.GetSignalHireWebhookPath(),
		middlewares.WebhookToken(config.GetSignalHireWebhookSecret()),
		handler.SignalHireWebhook,
	)
}
