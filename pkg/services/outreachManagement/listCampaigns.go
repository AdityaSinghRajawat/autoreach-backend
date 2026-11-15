package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) ListCampaigns(
	ctx context.Context,
	candidateID string,
) ([]*outreachTypes.CampaignDetail, *config.CustomError) {
	campaigns, err := s.campaignRepo.ListByCandidate(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list campaigns: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return s.withSteps(ctx, campaigns)
}

func (s *OutreachService) withSteps(
	ctx context.Context,
	campaigns []*outreachModel.Campaign,
) ([]*outreachTypes.CampaignDetail, *config.CustomError) {
	campaignIDs := make([]string, 0, len(campaigns))
	for _, campaign := range campaigns {
		campaignIDs = append(campaignIDs, campaign.ID)
	}

	steps, err := s.campaignRepo.ListSteps(ctx, campaignIDs)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("list campaign steps: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	stepsByCampaign := map[string][]*outreachModel.CampaignStep{}
	for _, step := range steps {
		stepsByCampaign[step.CampaignID] = append(stepsByCampaign[step.CampaignID], step)
	}

	details := make([]*outreachTypes.CampaignDetail, 0, len(campaigns))
	for _, campaign := range campaigns {
		campaignSteps := stepsByCampaign[campaign.ID]
		if campaignSteps == nil {
			campaignSteps = []*outreachModel.CampaignStep{}
		}

		details = append(
			details,
			&outreachTypes.CampaignDetail{Campaign: campaign, Steps: campaignSteps},
		)
	}

	return details, nil
}
