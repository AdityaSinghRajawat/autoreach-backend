package dashboardManagement

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

var (
	sqlApplicationStatusInterviewing  = db.Enum(modelHelpers.ApplicationStatusInterviewing)
	sqlJobStatusActive                = db.Enum(modelHelpers.JobStatusActive)
	sqlOutreachChannelEmail           = db.Enum(modelHelpers.OutreachChannelEmail)
	sqlOutreachChannelLinkedinConnect = db.Enum(modelHelpers.OutreachChannelLinkedinConnect)
	sqlOutreachStatusAccepted         = db.Enum(modelHelpers.OutreachStatusAccepted)
	sqlOutreachStatusBounced          = db.Enum(modelHelpers.OutreachStatusBounced)
	sqlOutreachStatusReplied          = db.Enum(modelHelpers.OutreachStatusReplied)
)
