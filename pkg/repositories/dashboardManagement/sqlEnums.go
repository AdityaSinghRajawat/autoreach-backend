package dashboardManagement

import (
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

var (
	sqlApplicationStatusInterviewing = utils.BuildSQLEnumList(
		modelHelpers.ApplicationStatusInterviewing,
	)
	sqlJobStatusActive                = utils.BuildSQLEnumList(modelHelpers.JobStatusActive)
	sqlOutreachChannelEmail           = utils.BuildSQLEnumList(modelHelpers.OutreachChannelEmail)
	sqlOutreachChannelLinkedinConnect = utils.BuildSQLEnumList(
		modelHelpers.OutreachChannelLinkedinConnect,
	)
	sqlOutreachStatusAccepted = utils.BuildSQLEnumList(modelHelpers.OutreachStatusAccepted)
	sqlOutreachStatusBounced  = utils.BuildSQLEnumList(modelHelpers.OutreachStatusBounced)
	sqlOutreachStatusReplied  = utils.BuildSQLEnumList(modelHelpers.OutreachStatusReplied)
)
