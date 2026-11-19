package outreachManagement

import (
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

var (
	sqlCampaignStatusActive       = utils.BuildSQLEnumList(modelHelpers.CampaignStatusActive)
	sqlChannelAccountStatusActive = utils.BuildSQLEnumList(
		modelHelpers.ChannelAccountStatusActive,
	)
	sqlOutreachChannelLinkedinConnect = utils.BuildSQLEnumList(
		modelHelpers.OutreachChannelLinkedinConnect,
	)
	sqlOutreachStatusActionRequired = utils.BuildSQLEnumList(
		modelHelpers.OutreachStatusActionRequired,
	)
	sqlOutreachStatusBounced   = utils.BuildSQLEnumList(modelHelpers.OutreachStatusBounced)
	sqlOutreachStatusCancelled = utils.BuildSQLEnumList(modelHelpers.OutreachStatusCancelled)
	sqlOutreachStatusFailed    = utils.BuildSQLEnumList(modelHelpers.OutreachStatusFailed)
	sqlOutreachStatusPaused    = utils.BuildSQLEnumList(modelHelpers.OutreachStatusPaused)
	sqlOutreachStatusPending   = utils.BuildSQLEnumList(modelHelpers.OutreachStatusPending)
	sqlOutreachStatusReplied   = utils.BuildSQLEnumList(modelHelpers.OutreachStatusReplied)
	sqlOutreachStatusSending   = utils.BuildSQLEnumList(modelHelpers.OutreachStatusSending)
	sqlSuppressionTypeDomain   = utils.BuildSQLEnumList(modelHelpers.SuppressionTypeDomain)
	sqlSuppressionTypeEmail    = utils.BuildSQLEnumList(modelHelpers.SuppressionTypeEmail)
)
