package outreachManagement

import (
	"net/url"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/emailTemplates"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *OutreachService) trackedHTMLBody(
	message *outreachModel.Message,
	candidate *candidateModel.Candidate,
) (string, error) {
	message.TrackingToken = nil

	if !candidate.IsEmailTrackingEnabled() || config.GetPublicBaseURL() == "" {
		return "", nil
	}

	token, err := utils.NewRandomHex(config.GetTrackingTokenBytes())
	if err != nil {
		return "", err
	}

	message.TrackingToken = &token

	return emailTemplates.GetOutreachEmailHTML(emailTemplates.NewOutreachEmailData(
		message.BodyContent,
		func(link string) string { return trackingClickURL(token, link) },
		config.GetPublicBaseURL()+config.GetTrackingOpenPath()+token,
	))
}

func trackingClickURL(token, link string) string {
	query := url.Values{}
	query.Set(config.GetTrackingURLParam(), link)
	query.Set(config.GetTrackingSignatureParam(), utils.Sign(trackingKey(), token, link))

	return config.GetPublicBaseURL() + config.GetTrackingClickPath() + token + "?" + query.Encode()
}

func trackingKey() []byte {
	return utils.DeriveKey(config.GetJWTSecret(), config.GetTrackingKeyLabel())
}
