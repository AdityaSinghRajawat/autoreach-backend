package enrichmentManagement

import (
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

var (
	sqlEmailStatusUnverified   = utils.BuildSQLEnumList(modelHelpers.EmailStatusUnverified)
	sqlEnrichmentStatusPending = utils.BuildSQLEnumList(modelHelpers.EnrichmentStatusPending)
)
