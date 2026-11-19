package applicationManagement

import (
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

var (
	sqlApplicationStatusInProgress = utils.BuildSQLEnumList(
		modelHelpers.ApplicationStatusInProgress,
	)
	sqlApplicationStatusQueued = utils.BuildSQLEnumList(modelHelpers.ApplicationStatusQueued)
	sqlJobStatusActive         = utils.BuildSQLEnumList(modelHelpers.JobStatusActive)
)
