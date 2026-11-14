package candidateManagement

import (
	"context"
	"fmt"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
	workerCustomTypes "github.com/AdityaSinghRajawat/autoreach-backend/worker/customtypes"
)

func (s *CandidateService) UploadResume(
	ctx context.Context,
	resume *candidateModel.Resume,
	content []byte,
) (*candidateTypes.UploadResumeResponse, *config.CustomError) {
	resume.UpdateTimestamp()

	if err := resume.Validate(); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("upload resume: %w", err),
			config.GetErrCodeBadRequest(),
		)
	}

	if err := s.fileStore.Save(ctx, resume.StorageKey, content); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("upload resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	uploaded, err := s.resumeRepo.Upsert(ctx, resume)
	if err != nil {
		s.deleteStoredFile(ctx, resume.StorageKey)

		return nil, config.NewCustomError(
			fmt.Errorf("upload resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	status := workerConfigs.PublishStatusEnqueued
	if err := s.publishProcessResume(ctx, uploaded.ID); err != nil {
		status = workerConfigs.PublishStatusDeferred

		utils.LogWarn(
			ctx,
			"process resume task publish failed, reconciler will retry",
			"resume_id", uploaded.ID,
			"error", err.Error(),
		)
	}

	return candidateTypes.NewUploadResumeResponse(
		uploaded,
		string(workerConfigs.TaskProcessResume),
		status,
	), nil
}

func (s *CandidateService) publishProcessResume(ctx context.Context, resumeID string) error {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskProcessResume,
		Payload: workerCustomTypes.ProcessResumePayload{ResumeID: resumeID},
	})

	return err
}

func (s *CandidateService) publishMatchResume(ctx context.Context, resumeID string) error {
	_, err := s.workerClient.Publish(ctx, workerConfigs.PublishParams{
		TaskID:  workerConfigs.TaskMatchResume,
		Payload: workerCustomTypes.MatchResumePayload{ResumeID: resumeID},
	})

	return err
}

func (s *CandidateService) deleteStoredFile(ctx context.Context, storageKey string) {
	utils.RunInBackground(ctx, "delete stored resume file", func(ctx context.Context) {
		if err := s.fileStore.Delete(ctx, storageKey); err != nil {
			utils.LogWarn(ctx, "stored resume file delete failed", "error", err.Error())
		}
	})
}
