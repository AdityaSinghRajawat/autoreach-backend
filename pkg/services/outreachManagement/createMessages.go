package outreachManagement

import (
	"context"
	"fmt"
	"slices"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) CreateMessages(
	ctx context.Context,
	candidateID string,
	request outreachTypes.CreateMessagesRequest,
) (*outreachTypes.CreateMessagesResponse, *config.CustomError) {
	if !request.HasOnlyEmail() {
		return nil, config.NewValidationError(
			fmt.Errorf("create outreach messages: %w", config.ErrChannelNotSupported),
			[]config.ErrorDetail{{Field: "channels", Issue: config.ErrChannelNotSupported.Error()}},
		)
	}

	message := request.ToMessageModel(candidateID)

	if details := templateContentIssues(&outreachModel.Template{
		Channel: modelHelpers.OutreachChannelEmail,
		Subject: message.Subject,
		Body:    message.BodyContent,
	}); len(details) > 0 {
		return nil, config.NewValidationError(
			fmt.Errorf("create outreach messages: %w", config.ErrUnknownPlaceholder),
			details,
		)
	}

	contact, customErr := s.reachableContact(ctx, candidateID, request.JobID, request.ContactID)
	if customErr != nil {
		return nil, customErr
	}

	account, err := s.senderAccount(ctx, candidateID, request.ChannelAccountID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create outreach messages: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if account == nil {
		return nil, config.NewValidationError(
			fmt.Errorf("create outreach messages: %w", config.ErrNoActiveMailbox),
			[]config.ErrorDetail{
				{Field: "channelAccountId", Issue: config.ErrNoActiveMailbox.Error()},
			},
		)
	}

	messageContext, customErr := s.messageContext(ctx, candidateID, request.JobID, contact)
	if customErr != nil {
		return nil, customErr
	}

	values := messageContext.PlaceholderValues()
	subject := utils.RenderPlaceholders(*message.Subject, values)
	message.Subject = &subject
	message.BodyContent = utils.RenderPlaceholders(message.BodyContent, values)
	message.ChannelAccountID = &account.ID

	saved, customErr := s.saveMessage(ctx, message, "create outreach messages")
	if customErr != nil {
		return nil, customErr
	}

	return &outreachTypes.CreateMessagesResponse{Messages: []*outreachModel.Message{saved}}, nil
}

func (s *OutreachService) reachableContact(
	ctx context.Context,
	candidateID, jobID, contactID string,
) (*enrichmentModel.CompanyContact, *config.CustomError) {
	contacts, err := s.companyContactRepo.ListByJob(ctx, jobID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("find outreach contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	index := slices.IndexFunc(contacts, func(contact *enrichmentModel.CompanyContact) bool {
		return contact.ID == contactID
	})
	if index < 0 || !isEmailable(contacts[index]) {
		return nil, config.NewValidationError(
			fmt.Errorf("find outreach contact: %w", config.ErrContactNotReachable),
			[]config.ErrorDetail{
				{Field: "contactId", Issue: config.ErrContactNotReachable.Error()},
			},
		)
	}

	suppressed, err := s.suppressionRepo.IsSuppressed(ctx, candidateID, *contacts[index].WorkEmail)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("find outreach contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if suppressed {
		return nil, config.NewValidationError(
			fmt.Errorf("find outreach contact: %w", config.ErrContactSuppressed),
			[]config.ErrorDetail{{Field: "contactId", Issue: config.ErrContactSuppressed.Error()}},
		)
	}

	return contacts[index], nil
}
