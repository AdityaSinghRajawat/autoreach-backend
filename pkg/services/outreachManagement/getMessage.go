package outreachManagement

import (
	"context"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

func (s *OutreachService) GetMessage(
	ctx context.Context,
	candidateID, messageID string,
) (*outreachModel.Message, *config.CustomError) {
	message, err := s.messageRepo.FindByCandidate(ctx, candidateID, messageID)

	return config.RequireFound("get outreach message", message, err, config.ErrMessageNotFound)
}
