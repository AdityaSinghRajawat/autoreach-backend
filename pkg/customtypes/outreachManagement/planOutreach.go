package outreachManagement

import (
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
