package outreachManagement

import (
	"context"
	"errors"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) saveCampaign(
	ctx context.Context,
	campaign *outreachModel.Campaign,
	stepRequests []outreachTypes.CampaignStepRequest,
	operation string,
) (*outreachTypes.CampaignDetail, *config.CustomError) {
	if customErr := s.checkCampaignReferences(
		ctx,
		campaign,
		stepRequests,
		operation,
	); customErr != nil {
		return nil, customErr
	}

	campaign.UpdateTimestamp()

	if err := campaign.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeBadRequest(),
		)
	}

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if _, err := s.campaignRepo.Upsert(ctx, campaign); err != nil {
			return err
		}

		if err := s.syncCampaignMessages(ctx, campaign); err != nil {
			return err
		}

		if len(stepRequests) == 0 {
			return nil
		}

		return s.campaignRepo.ReplaceSteps(
			ctx,
			campaign.ID,
			outreachTypes.ToCampaignStepModels(campaign.ID, stepRequests),
		)
	})
	if errors.Is(err, config.ErrActiveCampaignExists) {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeConflict(),
		)
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return s.GetCampaign(ctx, campaign.CandidateID, campaign.ID)
}

func (s *OutreachService) syncCampaignMessages(
	ctx context.Context,
	campaign *outreachModel.Campaign,
) error {
	switch campaign.Status {
	case modelHelpers.CampaignStatusPaused:
		return s.messageRepo.SetCampaignPaused(ctx, campaign.ID, campaign.UpdatedAt)
	case modelHelpers.CampaignStatusArchived:
		return s.messageRepo.SetCampaignCancelled(ctx, campaign.ID, campaign.UpdatedAt)
	default:
		return s.messageRepo.SetCampaignResumed(ctx, campaign.ID, campaign.UpdatedAt)
	}
}

func (s *OutreachService) checkCampaignReferences(
	ctx context.Context,
	campaign *outreachModel.Campaign,
	stepRequests []outreachTypes.CampaignStepRequest,
	operation string,
) *config.CustomError {
	if campaign.ChannelAccountID != nil {
		if _, customErr := s.getChannelAccount(
			ctx,
			campaign.CandidateID,
			*campaign.ChannelAccountID,
		); customErr != nil {
			return customErr
		}
	}

	if len(stepRequests) == 0 {
		return nil
	}

	templates, err := s.templateRepo.ListByIDs(
		ctx,
		campaign.CandidateID,
		outreachTypes.StepTemplateIDs(stepRequests),
	)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	emailTemplates := map[string]bool{}
	for _, template := range templates {
		emailTemplates[template.ID] = template.Channel == modelHelpers.OutreachChannelEmail
	}

	for _, step := range stepRequests {
		if !emailTemplates[step.TemplateID] {
			return config.NewValidationError(
				fmt.Errorf("%s: %w", operation, config.ErrCampaignStepTemplate),
				[]config.ErrorDetail{
					{Field: "steps", Issue: config.ErrCampaignStepTemplate.Error()},
				},
			)
		}
	}

	return nil
}
