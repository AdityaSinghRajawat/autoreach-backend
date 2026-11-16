package adminManagement

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (h *AdminHandler) ListDeadTasks(c *gin.Context) {
	var params sharedTypes.PageParams
	if customErr := utils.BindQueryAndValidate(c, &params); customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	tasks, meta, customErr := h.adminService.ListDeadTasks(
		c.Request.Context(),
		c.Param(config.GetQueueParam()),
		params,
	)
	if customErr != nil {
		utils.HandleCustomError(c, customErr)

		return
	}

	utils.BuildAndSendPaginatedResponse(c, tasks, *meta)
}
