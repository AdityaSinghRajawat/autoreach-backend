package outreachManagement

import (
	"context"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) PlanOutreach(
	ctx context.Context,
	candidateID, jobID string,
) (*outreachTypes.PlanOutreachResult, *config.CustomError) {
	plan, skippedReason, err := s.loadPlan(ctx, candidateID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("plan outreach: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if skippedReason != "" {
		return &outreachTypes.PlanOutreachResult{SkippedReason: skippedReason}, nil
	}

	contacts, err := s.outreachTargets(ctx, candidateID, jobID, plan)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("plan outreach: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	messages := []*outreachModel.Message{}

	for _, contact := range contacts {
		messageContext, customErr := s.messageContext(ctx, candidateID, jobID, contact)
		if customErr != nil {
			return nil, customErr
		}

		messages = append(messages, plannedMessages(plan, messageContext)...)
	}

	planned, err := s.messageRepo.InsertPlanned(ctx, messages)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("plan outreach: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return &outreachTypes.PlanOutreachResult{Planned: planned}, nil
}

func (s *OutreachService) loadPlan(
	ctx context.Context,
	candidateID string,
) (*outreachTypes.CampaignPlan, string, error) {
	candidate, err := s.candidateRepo.FindByID(ctx, candidateID)
	if err != nil || candidate == nil || !candidate.AutoOutreachEnabled {
		return nil, config.GetSkipAutoOutreachDisabled(), err
	}

	campaign, err := s.campaignRepo.FindActive(ctx, candidateID)
	if err != nil || campaign == nil {
		return nil, config.GetSkipNoActiveCampaign(), err
	}

	steps, err := s.campaignRepo.ListSteps(ctx, []string{campaign.ID})
	if err != nil {
		return nil, "", err
	}

	templateIDs := make([]string, 0, len(steps))
	for _, step := range steps {
		templateIDs = append(templateIDs, step.TemplateID)
	}

	templates, err := s.templateRepo.ListByIDs(ctx, candidateID, templateIDs)
	if err != nil {
		return nil, "", err
	}

	plan := &outreachTypes.CampaignPlan{
		Campaign:  campaign,
		Steps:     steps,
		Templates: map[string]*outreachModel.Template{},
	}

	for _, template := range templates {
		if !template.IsApproved() {
			return nil, config.GetSkipTemplateNotApproved(), nil
		}

		plan.Templates[template.ID] = template
	}

	if len(steps) == 0 || len(plan.Templates) != len(templateIDs) {
		return nil, config.GetSkipTemplateNotApproved(), nil
	}

	if !plan.HasEmailStep() {
		return plan, "", nil
	}

	plan.Account, err = s.senderAccount(ctx, candidateID, campaign.ChannelAccountID)
	if err != nil || plan.Account == nil {
		return nil, config.GetSkipNoActiveMailbox(), err
	}

	return plan, "", nil
}

func (s *OutreachService) senderAccount(
	ctx context.Context,
	candidateID string,
	preferredAccountID *string,
) (*outreachModel.ChannelAccount, error) {
	if preferredAccountID != nil {
		account, err := s.channelAccountRepo.FindByCandidate(ctx, candidateID, *preferredAccountID)
		if err != nil ||
			(account != nil && account.Status == modelHelpers.ChannelAccountStatusActive) {
			return account, err
		}
	}

	return s.channelAccountRepo.FindFirstActive(ctx, candidateID)
}

func (s *OutreachService) outreachTargets(
	ctx context.Context,
	candidateID, jobID string,
	plan *outreachTypes.CampaignPlan,
) ([]*enrichmentModel.CompanyContact, error) {
	contacts, err := s.companyContactRepo.ListByJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	targets := []*enrichmentModel.CompanyContact{}

	for _, contact := range contacts {
		if len(targets) == config.GetOutreachContactsPerJob() {
			break
		}

		if !plan.CanReach(contact) {
			continue
		}

		suppressed, err := s.isContactSuppressed(ctx, candidateID, contact)
		if err != nil {
			return nil, err
		}

		contacted, err := s.messageRepo.HasContacted(ctx, candidateID, contact.ID)
		if err != nil {
			return nil, err
		}

		if !suppressed && !contacted {
			targets = append(targets, contact)
		}
	}

	return targets, nil
}

func plannedMessages(
	plan *outreachTypes.CampaignPlan,
	messageContext outreachTypes.MessageContext,
) []*outreachModel.Message {
	values := messageContext.PlaceholderValues()
	scheduledAt := utils.GetCurrTimeStamp()
	messages := make([]*outreachModel.Message, 0, len(plan.Steps))

	for _, step := range plan.Steps {
		template := plan.Templates[step.TemplateID]
		scheduledAt = scheduledAt.Add(time.Duration(step.DelayDays) * 24 * time.Hour)

		if !outreachTypes.CanReachOn(messageContext.Contact, template.Channel) {
			continue
		}

		message := outreachModel.NewMessage()
		message.CandidateID = messageContext.Candidate.ID
		message.ContactID = messageContext.Contact.ID
		message.JobID = messageContext.Job.ID
		message.CampaignID = &plan.Campaign.ID
		message.TemplateID = &template.ID
		message.Channel = template.Channel
		message.StepNumber = step.StepNumber
		message.ScheduledAt = scheduledAt
		message.BodyContent = utils.RenderPlaceholders(template.Body, values)

		switch template.Channel {
		case modelHelpers.OutreachChannelEmail:
			message.ChannelAccountID = &plan.Account.ID
			subject := utils.RenderPlaceholders(valueOrEmpty(template.Subject), values)
			message.Subject = &subject
		case modelHelpers.OutreachChannelLinkedinConnect:
			attachLinkedinNote(message)
		}

		messages = append(messages, message)
	}

	return messages
}

// A rendered note over LinkedIn's limit is stored as FAILED, so the candidate sees why it was skipped.
func attachLinkedinNote(message *outreachModel.Message) {
	if len([]rune(message.BodyContent)) > config.GetLinkedinNoteMaxLength() {
		message.Close(modelHelpers.OutreachStatusFailed, config.ErrLinkedinNoteTooLong.Error())

		return
	}

	note := message.BodyContent
	message.LinkedinNote = &note
}

func (s *OutreachService) isContactSuppressed(
	ctx context.Context,
	candidateID string,
	contact *enrichmentModel.CompanyContact,
) (bool, error) {
	if contact.WorkEmail == nil {
		return false, nil
	}

	return s.suppressionRepo.IsSuppressed(ctx, candidateID, *contact.WorkEmail)
}
