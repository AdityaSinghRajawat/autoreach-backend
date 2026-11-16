package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) GetCampaign(
	ctx context.Context,
	candidateID, campaignID string,
) (*outreachTypes.CampaignDetail, *config.CustomError) {
	campaign, customErr := s.findCampaign(ctx, candidateID, campaignID)
	if customErr != nil {
		return nil, customErr
	}

	details, customErr := s.withSteps(ctx, []*outreachModel.Campaign{campaign})
	if customErr != nil {
		return nil, customErr
	}

	return details[0], nil
}

func (s *OutreachService) findCampaign(
	ctx context.Context,
	candidateID, campaignID string,
) (*outreachModel.Campaign, *config.CustomError) {
	campaign, err := s.campaignRepo.FindByCandidate(ctx, candidateID, campaignID)

	return config.RequireFound("get campaign", campaign, err, config.ErrCampaignNotFound)
}
