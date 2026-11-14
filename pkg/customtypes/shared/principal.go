package shared

import modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"

type Principal struct {
	UserID      string
	CandidateID string
	Role        modelHelpers.UserRole
}

func (p Principal) CanAccessCandidate(candidateID string) bool {
	if p.Role.IsAdmin() {
		return true
	}

	return p.CandidateID != "" && p.CandidateID == candidateID
}
