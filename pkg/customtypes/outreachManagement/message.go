package outreachManagement

import (
	"slices"
	"strings"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type CreateMessagesRequest struct {
	JobID            string   `json:"jobId"            validate:"required,uuid"`
	ContactID        string   `json:"contactId"        validate:"required,uuid"`
	Channels         []string `json:"channels"         validate:"required,min=1,unique,dive,oneof=EMAIL LINKEDIN_CONNECT LINKEDIN_DM"`
	Subject          string   `json:"subject"          validate:"max=500"`
	EmailBody        string   `json:"emailBody"        validate:"max=20000"`
	LinkedinNote     *string  `json:"linkedinNote"     validate:"omitnil,max=300"`
	ChannelAccountID *string  `json:"channelAccountId" validate:"omitnil,uuid"`
}

type UpdateMessageRequest struct {
	Status string `json:"status" validate:"required,oneof=CANCELLED"`
}

type ListMessagesParams struct {
	sharedTypes.PageParams
	MessageFilter
}

type CreateMessagesResponse struct {
	Messages []*outreachModel.Message `json:"messages"`
}

func (r *CreateMessagesRequest) MissingFields() []string {
	missing := []string{}

	for _, channel := range r.Channels {
		switch modelHelpers.OutreachChannel(channel) {
		case modelHelpers.OutreachChannelEmail:
			if strings.TrimSpace(r.Subject) == "" {
				missing = append(missing, "subject")
			}

			if strings.TrimSpace(r.EmailBody) == "" {
				missing = append(missing, "emailBody")
			}
		case modelHelpers.OutreachChannelLinkedinConnect:
			if r.LinkedinNote == nil || strings.TrimSpace(*r.LinkedinNote) == "" {
				missing = append(missing, "linkedinNote")
			}
		case modelHelpers.OutreachChannelLinkedinDM:
			if strings.TrimSpace(r.EmailBody) == "" {
				missing = append(missing, "emailBody")
			}
		}
	}

	slices.Sort(missing)

	return slices.Compact(missing)
}

func (r *CreateMessagesRequest) ToMessageModels(candidateID string) []*outreachModel.Message {
	messages := make([]*outreachModel.Message, 0, len(r.Channels))

	for _, channel := range r.Channels {
		message := outreachModel.NewMessage()
		message.CandidateID = candidateID
		message.JobID = r.JobID
		message.ContactID = r.ContactID
		message.Channel = modelHelpers.OutreachChannel(channel)

		switch message.Channel {
		case modelHelpers.OutreachChannelEmail:
			subject := r.Subject
			message.Subject = &subject
			message.BodyContent = r.EmailBody
		case modelHelpers.OutreachChannelLinkedinConnect:
			message.BodyContent = *r.LinkedinNote
		default:
			message.BodyContent = r.EmailBody
		}

		messages = append(messages, message)
	}

	return messages
}

type MessageFilter struct {
	Status string `form:"status" validate:"omitempty,oneof=PENDING SENDING ACTION_REQUIRED SENT ACCEPTED REPLIED FAILED BOUNCED PAUSED CANCELLED"`
}
