package outreachManagement

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *OutreachHandler) ShowUnsubscribePage(c *gin.Context) {
	customErr := h.outreachService.VerifyUnsubscribeLink(
		c.Request.Context(),
		c.Param(config.GetMessageIDParam()),
		c.Query(config.GetTrackingSignatureParam()),
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	page, err := emailTemplates.GetUnsubscribePageHTML(emailTemplates.UnsubscribePageData{
		ActionURL:    c.Request.URL.RequestURI(),
		Unsubscribed: false,
	})
	if err != nil {
		utils.HandleCustomError(
			c,
			config.NewCustomError(err, config.GetErrCodeInternalServerError()),
		)

		return
	}

	c.Data(http.StatusOK, config.GetHTMLContentType(), []byte(page))
}
