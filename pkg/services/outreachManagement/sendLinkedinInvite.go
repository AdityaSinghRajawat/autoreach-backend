package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
)

func (s *OutreachService) SendLinkedinInvite(
	ctx context.Context,
	messageID string,
) *config.CustomError {
	return s.deliverLinkedin(ctx, messageID, modelHelpers.OutreachChannelLinkedinConnect)
}
