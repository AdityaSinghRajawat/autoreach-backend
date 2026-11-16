package outreachManagement

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// RecordClick redirects only to an http(s) link that this service signed and that appears in the
// tracked message, so the endpoint is not an open redirect even if the signing key leaked.
func (s *OutreachService) RecordClick(
	ctx context.Context,
	token, link, signature string,
) (string, *config.CustomError) {
	message := s.trackedMessage(ctx, token)

	if !utils.VerifySignature(trackingKey(), signature, token, link) || message == nil ||
		!isWebLink(link) || !strings.Contains(message.BodyContent, link) {
		return "", config.NewCustomError(
			fmt.Errorf("record click: %w", config.ErrInvalidTrackingLink),
			config.GetErrCodeNotFound(),
		)
	}

	if customErr := s.recordEvent(
		ctx,
		message.ID,
		modelHelpers.OutreachEventClicked,
	); customErr != nil {
		utils.LogError(ctx, customErr)
	}

	return link, nil
}

func isWebLink(link string) bool {
	parsed, err := url.Parse(link)

	return err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host != ""
}
