package signalhire

import (
	"encoding/json"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	enrichmentTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/enrichmentManagement"
	signalhireIntegration "github.com/AdityaSinghRajawat/autoreach-backend/pkg/integrations/enrichment/signalhire"
)

func (p *SignalHireProvider) ParseCallback(
	payload []byte,
) ([]enrichmentTypes.CallbackResult, error) {
	items := []signalhireIntegration.CallbackItem{}
	if err := json.Unmarshal(payload, &items); err != nil {
		return nil, fmt.Errorf(
			"signalhire parse callback: %w: %w",
			config.ErrMalformedProviderPayload,
			err,
		)
	}

	results := make([]enrichmentTypes.CallbackResult, 0, len(items))

	for _, item := range items {
		result := enrichmentTypes.CallbackResult{Item: item.Item}

		if item.Status == config.GetSignalHireStatusSuccess() && item.Candidate != nil {
			raw, err := json.Marshal(item.Candidate)
			if err != nil {
				raw = nil
			}

			result.Found = true
			result.Details = toContactDetails(*item.Candidate, raw)
		}

		results = append(results, result)
	}

	return results, nil
}

func toContactDetails(
	candidate signalhireIntegration.Candidate,
	raw []byte,
) enrichmentTypes.ContactDetails {
	details := enrichmentTypes.ContactDetails{
		FullName: candidate.FullName,
		Title:    candidate.Headline,
		Raw:      raw,
	}

	for _, channel := range candidate.Contacts {
		switch channel.Type {
		case config.GetSignalHireChannelEmail():
			if details.Email == "" || channel.SubType == config.GetSignalHireSubTypeWork() {
				details.Email = channel.Value
				details.EmailVerified = config.IsSignalHireVerifiedRating(channel.Rating)
			}
		case config.GetSignalHireChannelPhone():
			if details.Phone == "" {
				details.Phone = channel.Value
			}
		}
	}

	return details
}
