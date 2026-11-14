package matchingManagement

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/validations"
)

type JobMatch struct {
	ID            string     `db:"id"             json:"matchId"       validate:"required,uuid"`
	CandidateID   string     `db:"candidate_id"   json:"candidateId"   validate:"required,uuid"`
	JobID         string     `db:"job_id"         json:"jobId"         validate:"required,uuid"`
	ResumeID      string     `db:"resume_id"      json:"resumeId"      validate:"required,uuid"`
	RawSimilarity float64    `db:"raw_similarity" json:"rawSimilarity" validate:"gte=-1,lte=1"`
	MatchScore    float64    `db:"match_score"    json:"matchScore"    validate:"gte=0,lte=1"`
	FannedOutAt   *time.Time `db:"fanned_out_at"  json:"fannedOutAt"`
	CreatedAt     time.Time  `db:"created_at"     json:"createdAt"     validate:"required"`
	UpdatedAt     time.Time  `db:"updated_at"     json:"updatedAt"     validate:"required"`
}

func NewJobMatch() *JobMatch {
	now := utils.GetCurrTimeStamp()

	return &JobMatch{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (m *JobMatch) UpdateTimestamp() {
	m.UpdatedAt = utils.GetCurrTimeStamp()
}

func (m *JobMatch) Validate() error {
	if err := validations.Validate(m); err != nil {
		return fmt.Errorf("job match validation failed: %w", err)
	}

	return nil
}
