package candidateManagement

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

var (
	sqlJobStatusActive      = db.Enum(modelHelpers.JobStatusActive)
	sqlResumeStatusEmbedded = db.Enum(modelHelpers.ResumeStatusEmbedded)
	sqlResumeStatusParsed   = db.Enum(modelHelpers.ResumeStatusParsed)
	sqlResumeStatusUploaded = db.Enum(modelHelpers.ResumeStatusUploaded)
)
