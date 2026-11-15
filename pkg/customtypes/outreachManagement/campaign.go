package outreachManagement

import (
	"strings"

	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type CampaignStepRequest struct {
	TemplateID string `json:"templateId" validate:"required,uuid"`
	DelayDays  int    `json:"delayDays"  validate:"gte=0,lte=60"`
}

type CreateCampaignRequest struct {
	Name             string                `json:"name"             validate:"required,max=255"`
	ChannelAccountID *string               `json:"channelAccountId" validate:"omitnil,uuid"`
	Status           *string               `json:"status"           validate:"omitnil,oneof=ACTIVE PAUSED"`
	Steps            []CampaignStepRequest `json:"steps"            validate:"required,min=1,max=5,dive"`
}

type UpdateCampaignRequest struct {
	Name             *string               `json:"name"             validate:"omitnil,min=1,max=255"`
	ChannelAccountID *string               `json:"channelAccountId" validate:"omitnil,uuid"`
	Status           *string               `json:"status"           validate:"omitnil,oneof=ACTIVE PAUSED ARCHIVED"`
	Steps            []CampaignStepRequest `json:"steps"            validate:"omitempty,min=1,max=5,dive"`
}

type CampaignDetail struct {
	*outreachModel.Campaign
	Steps []*outreachModel.CampaignStep `json:"steps"`
}

func (r *CreateCampaignRequest) ToCampaignModel(candidateID string) *outreachModel.Campaign {
	campaign := outreachModel.NewCampaign()
	campaign.CandidateID = candidateID
	campaign.Name = strings.TrimSpace(r.Name)
	campaign.ChannelAccountID = r.ChannelAccountID

	if r.Status != nil {
		campaign.Status = modelHelpers.CampaignStatus(*r.Status)
	}

	return campaign
}

func (r *UpdateCampaignRequest) IsEmpty() bool {
	return r.Name == nil && r.ChannelAccountID == nil && r.Status == nil && len(r.Steps) == 0
}

func (r *UpdateCampaignRequest) ToCampaignModel(
	existing *outreachModel.Campaign,
) *outreachModel.Campaign {
	campaign := *existing

	if r.Name != nil {
		campaign.Name = strings.TrimSpace(*r.Name)
	}

	if r.ChannelAccountID != nil {
		campaign.ChannelAccountID = r.ChannelAccountID
	}

	if r.Status != nil {
		campaign.Status = modelHelpers.CampaignStatus(*r.Status)
	}

	return &campaign
}

func ToCampaignStepModels(
	campaignID string,
	requests []CampaignStepRequest,
) []*outreachModel.CampaignStep {
	steps := make([]*outreachModel.CampaignStep, 0, len(requests))
	for index, request := range requests {
		steps = append(steps, &outreachModel.CampaignStep{
			CampaignID: campaignID,
			StepNumber: index + 1,
			TemplateID: request.TemplateID,
			DelayDays:  request.DelayDays,
		})
	}

	return steps
}

func StepTemplateIDs(requests []CampaignStepRequest) []string {
	templateIDs := make([]string, 0, len(requests))
	for _, request := range requests {
		templateIDs = append(templateIDs, request.TemplateID)
	}

	return templateIDs
}
