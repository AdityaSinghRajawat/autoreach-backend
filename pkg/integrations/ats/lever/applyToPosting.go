package lever

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func (i *LeverIntegration) ApplyToPosting(
	ctx context.Context,
	site, postingID, apiKey string,
	submission applicationTypes.Submission,
) (string, error) {
	fields := map[string][]string{
		"name":  {submission.FirstName + " " + submission.LastName},
		"email": {submission.Email},
	}

	for _, answer := range submission.Answers {
		if answer.FieldName != "resume" && fields[answer.FieldName] == nil {
			fields[answer.FieldName] = []string{answer.Value}
		}
	}

	files := []sharedTypes.MultipartFile{{
		FieldName:   "resume",
		FileName:    submission.ResumeName,
		ContentType: submission.ResumeType,
		Content:     submission.Resume,
	}}

	response := applyResponse{}

	err := i.httpClient.PostMultipart(
		ctx,
		postingPath(site, postingID)+config.GetLeverSiteKeyQuery()+apiKey,
		nil,
		fields,
		files,
		&response,
	)
	if err != nil {
		return "", fmt.Errorf("apply to lever posting: %w", err)
	}

	if !response.OK {
		return "", fmt.Errorf("apply to lever posting: %w", config.ErrATSRejected)
	}

	return response.ApplicationID, nil
}
