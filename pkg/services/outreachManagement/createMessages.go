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
	if missing := request.MissingFields(); len(missing) > 0 {
		details := make([]config.ErrorDetail, 0, len(missing))
		for _, field := range missing {
			details = append(
				details,
				config.ErrorDetail{Field: field, Issue: config.ErrMessageFieldRequired.Error()},
			)
		}

		return nil, config.NewValidationError(
			fmt.Errorf("create outreach messages: %w", config.ErrMessageFieldRequired),
			details,
		)
	}

	messages := request.ToMessageModels(candidateID)

	for _, message := range messages {
		if details := templateContentIssues(&outreachModel.Template{
			Channel: message.Channel,
			Subject: message.Subject,
			Body:    message.BodyContent,
		}); len(details) > 0 {
			return nil, config.NewValidationError(
				fmt.Errorf("create outreach messages: %w", config.ErrUnknownPlaceholder),
				details,
			)
		}
	}

	contact, customErr := s.reachableContact(ctx, candidateID, request, messages)
	if customErr != nil {
		return nil, customErr
	}

	messageContext, customErr := s.messageContext(ctx, candidateID, request.JobID, contact)
	if customErr != nil {
		return nil, customErr
	}

	if customErr := s.prepareManualMessages(
		ctx,
		candidateID,
		request,
		messages,
		messageContext,
	); customErr != nil {
		return nil, customErr
	}

	saved := make([]*outreachModel.Message, 0, len(messages))

	var saveErr *config.CustomError

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		for _, message := range messages {
			created, customErr := s.saveMessage(ctx, message, "create outreach messages")
			if customErr != nil {
				saveErr = customErr

				return customErr
			}

			saved = append(saved, created)
		}

		return nil
	})

	if saveErr != nil {
		return nil, saveErr
	}

	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("create outreach messages: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return &outreachTypes.CreateMessagesResponse{Messages: saved}, nil
}

func (s *OutreachService) prepareManualMessages(
	ctx context.Context,
	candidateID string,
	request outreachTypes.CreateMessagesRequest,
	messages []*outreachModel.Message,
	messageContext outreachTypes.MessageContext,
) *config.CustomError {
	values := messageContext.PlaceholderValues()

	for _, message := range messages {
		message.BodyContent = utils.RenderPlaceholders(message.BodyContent, values)

		switch message.Channel {
		case modelHelpers.OutreachChannelEmail:
			account, err := s.senderAccount(ctx, candidateID, request.ChannelAccountID)
			if err != nil {
				return config.NewCustomError(
					fmt.Errorf("create outreach messages: %w", err),
					config.GetErrCodeInternalServerError(),
				)
			}

			if account == nil {
				return config.NewValidationError(
					fmt.Errorf("create outreach messages: %w", config.ErrNoActiveMailbox),
					[]config.ErrorDetail{
						{Field: "channelAccountId", Issue: config.ErrNoActiveMailbox.Error()},
					},
				)
			}

			subject := utils.RenderPlaceholders(*message.Subject, values)
			message.Subject = &subject
			message.ChannelAccountID = &account.ID
		case modelHelpers.OutreachChannelLinkedinConnect:
			if len([]rune(message.BodyContent)) > config.GetLinkedinNoteMaxLength() {
				return config.NewValidationError(
					fmt.Errorf("create outreach messages: %w", config.ErrLinkedinNoteTooLong),
					[]config.ErrorDetail{
						{Field: "linkedinNote", Issue: config.ErrLinkedinNoteTooLong.Error()},
					},
				)
			}

			note := message.BodyContent
			message.LinkedinNote = &note
		}
	}

	return nil
}

func (s *OutreachService) reachableContact(
	ctx context.Context,
	candidateID string,
	request outreachTypes.CreateMessagesRequest,
	messages []*outreachModel.Message,
) (*enrichmentModel.CompanyContact, *config.CustomError) {
	matched, err := s.jobMatchRepo.Exists(ctx, candidateID, request.JobID)
	if customErr := config.RequireAffected(
		"find outreach contact",
		matched,
		err,
		config.ErrJobNotMatched,
	); customErr != nil {
		return nil, customErr
	}

	contacts, err := s.companyContactRepo.ListByJob(ctx, request.JobID)
	if err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("find outreach contact: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	index := slices.IndexFunc(contacts, func(contact *enrichmentModel.CompanyContact) bool {
		return contact.ID == request.ContactID
	})

	for _, message := range messages {
		if index < 0 || !outreachTypes.CanReachOn(contacts[index], message.Channel) {
			return nil, config.NewValidationError(
				fmt.Errorf("find outreach contact: %w", config.ErrContactNotReachable),
				[]config.ErrorDetail{
					{Field: "contactId", Issue: config.ErrContactNotReachable.Error()},
				},
			)
		}
	}

	suppressed, err := s.isContactSuppressed(ctx, candidateID, contacts[index])
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
