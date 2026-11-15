package greenhouse

import (
	"context"
	"fmt"

	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

func (i *GreenhouseIntegration) SubmitApplication(
	ctx context.Context,
	boardToken, jobID, apiKey string,
	submission applicationTypes.Submission,
) error {
	fields := map[string][]string{
		"first_name": {submission.FirstName},
		"last_name":  {submission.LastName},
		"email":      {submission.Email},
	}

	for _, answer := range submission.Answers {
		switch {
		case answer.FieldName == "resume" || fields[answer.FieldName] != nil:
			continue
		case len(answer.Values) > 0:
			fields[answer.FieldName+"[]"] = answer.Values
		default:
			fields[answer.FieldName] = []string{answer.Value}
		}
	}

	files := []sharedTypes.MultipartFile{{
		FieldName:   "resume",
		FileName:    submission.ResumeName,
		ContentType: submission.ResumeType,
		Content:     submission.Resume,
	}}

	if err := i.httpClient.PostMultipart(
		ctx,
		jobPath(boardToken, jobID),
		basicAuth(apiKey),
		fields,
		files,
		nil,
	); err != nil {
		return fmt.Errorf("submit greenhouse application: %w", err)
	}

	return nil
}
