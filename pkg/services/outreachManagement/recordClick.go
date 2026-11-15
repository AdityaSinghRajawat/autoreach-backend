package outreachManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

// RecordClick only redirects to links this service signed, so the endpoint is not an open redirect.
func (s *OutreachService) RecordClick(
	ctx context.Context,
	token, link, signature string,
) (string, *config.CustomError) {
	if !utils.VerifySignature(trackingKey(), signature, token, link) {
		return "", config.NewCustomError(
			fmt.Errorf("record click: %w", config.ErrInvalidTrackingLink),
			config.GetErrCodeNotFound(),
		)
	}

	if message := s.trackedMessage(ctx, token); message != nil {
		if customErr := s.recordEvent(
			ctx,
			message.ID,
			modelHelpers.OutreachEventClicked,
		); customErr != nil {
			utils.LogError(ctx, customErr)
		}
	}

	return link, nil
}
