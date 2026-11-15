package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
)

func (s *OutreachService) UpdateCampaign(
	ctx context.Context,
	candidateID, campaignID string,
	request outreachTypes.UpdateCampaignRequest,
) (*outreachTypes.CampaignDetail, *config.CustomError) {
	if request.IsEmpty() {
		return nil, config.NewCustomError(
			fmt.Errorf("update campaign: %w", config.ErrNoFieldsSupplied),
			config.GetErrCodeBadRequest(),
		)
	}

	existing, customErr := s.findCampaign(ctx, candidateID, campaignID)
	if customErr != nil {
		return nil, customErr
	}

	return s.saveCampaign(ctx, request.ToCampaignModel(existing), request.Steps, "update campaign")
}
