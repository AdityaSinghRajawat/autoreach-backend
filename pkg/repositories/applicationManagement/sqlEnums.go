package applicationManagement

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

var (
	sqlApplicationStatusInProgress = db.Enum(modelHelpers.ApplicationStatusInProgress)
	sqlApplicationStatusQueued     = db.Enum(modelHelpers.ApplicationStatusQueued)
	sqlJobStatusActive             = db.Enum(modelHelpers.JobStatusActive)
)
