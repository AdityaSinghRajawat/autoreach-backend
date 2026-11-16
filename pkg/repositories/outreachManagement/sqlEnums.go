package outreachManagement

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

var (
	sqlCampaignStatusActive           = db.Enum(modelHelpers.CampaignStatusActive)
	sqlChannelAccountStatusActive     = db.Enum(modelHelpers.ChannelAccountStatusActive)
	sqlOutreachChannelLinkedinConnect = db.Enum(modelHelpers.OutreachChannelLinkedinConnect)
	sqlOutreachStatusActionRequired   = db.Enum(modelHelpers.OutreachStatusActionRequired)
	sqlOutreachStatusBounced          = db.Enum(modelHelpers.OutreachStatusBounced)
	sqlOutreachStatusCancelled        = db.Enum(modelHelpers.OutreachStatusCancelled)
	sqlOutreachStatusFailed           = db.Enum(modelHelpers.OutreachStatusFailed)
	sqlOutreachStatusPaused           = db.Enum(modelHelpers.OutreachStatusPaused)
	sqlOutreachStatusPending          = db.Enum(modelHelpers.OutreachStatusPending)
	sqlOutreachStatusReplied          = db.Enum(modelHelpers.OutreachStatusReplied)
	sqlOutreachStatusSending          = db.Enum(modelHelpers.OutreachStatusSending)
	sqlSuppressionTypeDomain          = db.Enum(modelHelpers.SuppressionTypeDomain)
	sqlSuppressionTypeEmail           = db.Enum(modelHelpers.SuppressionTypeEmail)
)
