package outreachManagement

import (
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type PlanOutreachResult struct {
	Planned       int    `json:"planned"`
	SkippedReason string `json:"skippedReason,omitempty"`
}

type CampaignPlan struct {
	Campaign  *outreachModel.Campaign
	Account   *outreachModel.ChannelAccount
	Steps     []*outreachModel.CampaignStep
	Templates map[string]*outreachModel.Template
}

func (p *CampaignPlan) HasEmailStep() bool {
	for _, template := range p.Templates {
		if template.Channel == modelHelpers.OutreachChannelEmail {
			return true
		}
	}

	return false
}

func (p *CampaignPlan) CanReach(contact *enrichmentModel.CompanyContact) bool {
	for _, template := range p.Templates {
		if CanReachOn(contact, template.Channel) {
			return true
		}
	}

	return false
}

func CanReachOn(
	contact *enrichmentModel.CompanyContact,
	channel modelHelpers.OutreachChannel,
) bool {
	if channel == modelHelpers.OutreachChannelEmail {
		return contact.WorkEmail != nil &&
			contact.EmailVerificationStatus != modelHelpers.EmailStatusInvalid
	}

	return contact.LinkedinURL != nil
}
