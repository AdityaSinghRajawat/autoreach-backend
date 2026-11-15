package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) CreateCampaign(
	ctx context.Context,
	campaign *outreachModel.Campaign,
	stepRequests []outreachTypes.CampaignStepRequest,
) (*outreachTypes.CampaignDetail, *config.CustomError) {
	return s.saveCampaign(ctx, campaign, stepRequests, "create campaign")
}
