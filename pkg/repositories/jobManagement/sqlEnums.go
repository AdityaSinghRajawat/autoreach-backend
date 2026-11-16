package jobManagement

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

var (
	sqlJobStatusActive = db.Enum(modelHelpers.JobStatusActive)
	sqlJobStatusClosed = db.Enum(modelHelpers.JobStatusClosed)
)
