package outreachManagement

import (
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type CreateMessagesRequest struct {
	JobID            string   `json:"jobId"            validate:"required,uuid"`
	ContactID        string   `json:"contactId"        validate:"required,uuid"`
	Channels         []string `json:"channels"         validate:"required,min=1,unique,dive,oneof=EMAIL LINKEDIN_CONNECT LINKEDIN_DM"`
	Subject          string   `json:"subject"          validate:"required,max=500"`
	EmailBody        string   `json:"emailBody"        validate:"required,max=20000"`
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

func (r *CreateMessagesRequest) HasOnlyEmail() bool {
	return len(r.Channels) == 1 && r.Channels[0] == string(modelHelpers.OutreachChannelEmail)
}

func (r *CreateMessagesRequest) ToMessageModel(candidateID string) *outreachModel.Message {
	message := outreachModel.NewMessage()
	message.CandidateID = candidateID
	message.JobID = r.JobID
	message.ContactID = r.ContactID
	message.ChannelAccountID = r.ChannelAccountID
	message.Subject = &r.Subject
	message.BodyContent = r.EmailBody

	return message
}

type MessageFilter struct {
	Status string `form:"status" validate:"omitempty,oneof=PENDING SENDING SENT REPLIED FAILED BOUNCED PAUSED CANCELLED"`
}
