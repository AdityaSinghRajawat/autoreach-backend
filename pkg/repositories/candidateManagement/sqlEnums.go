package candidateManagement

import (
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

var (
	sqlJobStatusActive      = utils.BuildSQLEnumList(modelHelpers.JobStatusActive)
	sqlResumeStatusEmbedded = utils.BuildSQLEnumList(modelHelpers.ResumeStatusEmbedded)
	sqlResumeStatusParsed   = utils.BuildSQLEnumList(modelHelpers.ResumeStatusParsed)
	sqlResumeStatusUploaded = utils.BuildSQLEnumList(modelHelpers.ResumeStatusUploaded)
)
