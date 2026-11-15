package outreachManagement

import (
	"time"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	enrichmentModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/enrichmentManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	outreachModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/outreachManagement"
)

type Delivery struct {
	Account      *outreachModel.ChannelAccount
	EmailAccount sharedTypes.EmailAccount
	Contact      *enrichmentModel.CompanyContact
	Candidate    *candidateModel.Candidate
	Location     *time.Location
}

type CloseOutcome struct {
	Status modelHelpers.OutreachStatus
	Reason error
}

type SendReservation struct {
	Reserved   bool
	CapReached bool
}
