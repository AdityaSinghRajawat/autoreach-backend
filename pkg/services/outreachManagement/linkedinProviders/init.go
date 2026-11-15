package linkedinProviders

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	outreachTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/outreachManagement"
	assistedProvider "github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/outreachManagement/linkedinProviders/assisted"
)

type Provider interface {
	Code() string
	SendInvite(
		ctx context.Context,
		delivery outreachTypes.LinkedinDelivery,
	) (outreachTypes.LinkedinOutcome, error)
	SendMessage(
		ctx context.Context,
		delivery outreachTypes.LinkedinDelivery,
	) (outreachTypes.LinkedinOutcome, error)
}

func NewProvider(code string) (Provider, error) {
	switch code {
	case config.GetLinkedinProviderAssisted():
		return assistedProvider.NewAssistedProvider(), nil
	default:
		return nil, fmt.Errorf(
			"new linkedin provider %q: %w",
			code,
			config.ErrUnknownLinkedinProvider,
		)
	}
}
