package jobManagement

import (
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

var (
	sqlJobStatusActive = utils.BuildSQLEnumList(modelHelpers.JobStatusActive)
	sqlJobStatusClosed = utils.BuildSQLEnumList(modelHelpers.JobStatusClosed)
)
