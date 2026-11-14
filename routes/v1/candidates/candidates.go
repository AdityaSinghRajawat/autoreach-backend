package candidates

import (
	"github.com/gin-gonic/gin"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/middlewares"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/dependencies"
)

func LoadCandidateRoutes(group *gin.RouterGroup, deps *dependencies.Dependencies) {
	candidateHandler := deps.Handlers.Candidate
	matchingHandler := deps.Handlers.Matching
	resumePath := "/resumes/:" + config.GetResumeIDParam()

	candidateGroup := group.Group(
		"/candidates/:"+config.GetCandidateIDParam(),
		middlewares.AuthenticateUser(),
		middlewares.TenantScope(),
	)
	candidateGroup.GET("", candidateHandler.GetCandidate)
	candidateGroup.PATCH("", candidateHandler.UpdateCandidate)

	candidateGroup.POST("/resumes", candidateHandler.UploadResume)
	candidateGroup.GET("/resumes", candidateHandler.ListResumes)
	candidateGroup.GET(resumePath, candidateHandler.GetResume)
	candidateGroup.PATCH(resumePath, candidateHandler.UpdateResume)
	candidateGroup.DELETE(resumePath, candidateHandler.DeleteResume)

	candidateGroup.GET("/jobs", matchingHandler.ListMatchedJobs)
}
