package outreachManagement

import (
	"context"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// Only follows signed links found in the message, so it is not an open redirect.
func (s *OutreachService) RecordClick(
	ctx context.Context,
	token, link, signature string,
) (string, *config.CustomError) {
	message := s.trackedMessage(ctx, token)

	if !utils.VerifySignature(trackingKey(), signature, token, link) || message == nil ||
		!utils.IsHTTPURL(link) || !strings.Contains(message.BodyContent, link) {
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
