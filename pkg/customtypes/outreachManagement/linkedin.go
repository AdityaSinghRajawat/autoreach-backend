package outreachManagement

type LinkedinDelivery struct {
	ProfileURL string
	Note       string
	Body       string
}

type LinkedinOutcome struct {
	AwaitingCandidate bool
}

type ConfirmMessageRequest struct {
	Status string `json:"status" validate:"required,oneof=SENT ACCEPTED REPLIED"`
}

type ClaimedMessage struct {
	ID      string
	Channel string
}
