package enrichmentManagement

import (
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/db"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

var (
	sqlEmailStatusUnverified   = db.Enum(modelHelpers.EmailStatusUnverified)
	sqlEnrichmentStatusPending = db.Enum(modelHelpers.EnrichmentStatusPending)
)
