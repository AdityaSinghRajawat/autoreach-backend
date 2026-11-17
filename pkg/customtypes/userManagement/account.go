package userManagement

import (
	"time"

	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	matchingModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/matchingManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	userModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/userManagement"
)

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,max=128"`
	NewPassword     string `json:"newPassword"     validate:"required,min=10,max=128"`
}

type DeleteAccountRequest struct {
	Password string `json:"password" validate:"required,max=128"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"newEmail" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,max=128"`
}

type ConfirmEmailChangeRequest struct {
	Token string `json:"token" validate:"required,max=256"`
	Email string `json:"email" validate:"required,email,max=255"`
}

type EmailChangePendingResponse struct {
	ConfirmationSent bool `json:"confirmationSent"`
}

type UserDataExport struct {
	ExportedAt        time.Time                         `json:"exportedAt"`
	User              *userModel.User                   `json:"user"`
	Candidate         *candidateModel.Candidate         `json:"candidate"`
	Resumes           []*candidateModel.Resume          `json:"resumes"`
	JobMatches        []*matchingModel.JobMatch         `json:"jobMatches"`
	Applications      []*applicationModel.Application   `json:"applications"`
	ChannelAccounts   []*outreachModel.ChannelAccount   `json:"channelAccounts"`
	OutreachTemplates []*outreachModel.Template         `json:"outreachTemplates"`
	OutreachCampaigns []*outreachModel.Campaign         `json:"outreachCampaigns"`
	OutreachMessages  []*outreachModel.Message          `json:"outreachMessages"`
	Suppressions      []*outreachModel.SuppressionEntry `json:"suppressions"`
	ContactSignals    []*enrichmentModel.ContactSignal  `json:"contactSignals"`
}

func (r *ChangeEmailRequest) NormalizedNewEmail() string {
	return modelHelpers.NormalizeEmail(r.NewEmail)
}

func (r *ConfirmEmailChangeRequest) NormalizedEmail() string {
	return modelHelpers.NormalizeEmail(r.Email)
}
