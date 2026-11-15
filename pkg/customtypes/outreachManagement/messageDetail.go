package outreachManagement

import (
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type MessageContact struct {
	ContactID   string  `json:"contactId"`
	FullName    string  `json:"fullName"`
	Title       string  `json:"title"`
	CompanyName string  `json:"companyName"`
	LinkedinURL *string `json:"linkedinUrl"`
}

type MessageDetail struct {
	*outreachModel.Message
	Contact *MessageContact `json:"contact"`
}

func NewMessageDetails(
	messages []*outreachModel.Message,
	contacts []*enrichmentModel.CompanyContact,
) []*MessageDetail {
	byID := make(map[string]*MessageContact, len(contacts))
	for _, contact := range contacts {
		byID[contact.ID] = &MessageContact{
			ContactID:   contact.ID,
			FullName:    contact.FullName,
			Title:       contact.Title,
			CompanyName: contact.CompanyName,
			LinkedinURL: contact.LinkedinURL,
		}
	}

	details := make([]*MessageDetail, 0, len(messages))
	for _, message := range messages {
		details = append(
			details,
			&MessageDetail{Message: message, Contact: byID[message.ContactID]},
		)
	}

	return details
}

func ContactIDs(messages []*outreachModel.Message) []string {
	contactIDs := make([]string, 0, len(messages))
	for _, message := range messages {
		contactIDs = append(contactIDs, message.ContactID)
	}

	return contactIDs
}
