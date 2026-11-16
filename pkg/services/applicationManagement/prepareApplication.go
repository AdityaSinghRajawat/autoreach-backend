package applicationManagement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	applicationTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/applicationManagement"
	applicationModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/applicationManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/services/applicationManagement/appliers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
	workerConfigs "github.com/AdityaSinghRajawat/autoreach-backend/worker/configs"
)

// Submits only through an employer's API key; otherwise the candidate submits the prepared answers.
// The application is claimed first, so a duplicate task or a reconciler requeue never runs it twice.
func (s *ApplicationService) PrepareApplication(
	ctx context.Context,
	applicationID string,
) *config.CustomError {
	now := utils.GetCurrTimeStamp()

	claimed, err := s.applicationRepo.Claim(
		ctx,
		applicationID,
		now.Add(-claimExpiry()),
		now,
	)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("prepare application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if !claimed {
		return nil
	}

	application, err := s.applicationRepo.FindByID(ctx, applicationID)
	if err != nil || application == nil {
		return config.NewCustomError(
			fmt.Errorf("prepare application: %w", errors.Join(err, config.ErrApplicationNotFound)),
			config.GetErrCodeInternalServerError(),
		)
	}

	if hasUnconfirmedSubmission(application) {
		return s.markOutcomeUnknown(ctx, application, config.ErrSubmissionOutcomeUnknown)
	}

	job, err := s.jobPostingRepo.FindByID(ctx, application.JobID)
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("prepare application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if job == nil {
		application.MarkFailed(config.ErrJobNotFound.Error())

		return s.saveApplication(ctx, application, "prepare application")
	}

	applier := appliers.Find(s.appliers, job)
	if applier == nil {
		appendLog(application, "UNSUPPORTED_FORM", config.GetUnsupportedATSNote())
		application.Status = modelHelpers.ApplicationStatusActionRequired

		return s.saveApplication(ctx, application, "prepare application")
	}

	return s.prepareWith(ctx, application, job, applier)
}

func (s *ApplicationService) prepareWith(
	ctx context.Context,
	application *applicationModel.Application,
	job *jobModel.JobPosting,
	applier appliers.Applier,
) *config.CustomError {
	if decision, err := s.rateLimiter.Allow(
		ctx,
		config.GetATSRateLimitPolicy(),
		applier.Code(),
		1,
	); err == nil &&
		!decision.Allowed {
		return s.releaseForRetry(ctx, application, config.ErrATSRateLimited)
	}

	form, err := applier.FetchForm(ctx, job)
	if err != nil {
		return s.handleUpstreamError(ctx, application, err, "FORM_UNAVAILABLE")
	}

	submission, customErr := s.buildSubmission(ctx, application, form)
	if customErr != nil {
		return customErr
	}

	appendLog(
		application,
		"FORM_PREPARED",
		fmt.Sprintf("%d answers prepared", len(submission.Answers)),
	)

	if form.HasHiddenQuestions {
		appendLog(application, "HIDDEN_QUESTIONS", config.GetFormHasHiddenQuestionsNote())
	}

	if hasMissingAnswers(application) || !applier.CanSubmit(job) || submission.Resume == nil {
		application.Status = modelHelpers.ApplicationStatusActionRequired

		return s.saveApplication(ctx, application, "prepare application")
	}

	appendLog(application, config.GetApplicationStepSubmitting(), "")

	if customErr := s.saveApplication(ctx, application, "prepare application"); customErr != nil {
		return customErr
	}

	confirmation, err := applier.Submit(ctx, job, submission)
	if err != nil {
		return s.handleSubmitError(ctx, application, err)
	}

	application.MarkSubmitted(nonEmpty(confirmation))
	appendLog(
		application,
		config.GetApplicationStepSubmitted(),
		"submitted through the employer's API",
	)

	return s.saveApplication(ctx, application, "prepare application")
}

func (s *ApplicationService) buildSubmission(
	ctx context.Context,
	application *applicationModel.Application,
	form applicationTypes.ApplicationForm,
) (applicationTypes.Submission, *config.CustomError) {
	candidate, customErr := s.findCandidate(ctx, application.CandidateID)
	if customErr != nil {
		return applicationTypes.Submission{}, customErr
	}

	profile, err := applicationProfileOf(candidate)
	if err != nil {
		return applicationTypes.Submission{}, config.NewCustomError(
			fmt.Errorf("prepare application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	submission, err := s.attachResume(ctx, application)
	if err != nil {
		return applicationTypes.Submission{}, config.NewCustomError(
			fmt.Errorf("prepare application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	submission.FirstName, submission.LastName = splitName(candidate.FullName)
	submission.Email = candidate.Email

	answers, missing := prepareAnswers(form, profile, candidate, submission.Resume != nil)
	submission.Answers = answers

	if application.PreparedAnswers, err = json.Marshal(answers); err == nil {
		application.MissingAnswers, err = json.Marshal(missing)
	}

	if err != nil {
		return applicationTypes.Submission{}, config.NewCustomError(
			fmt.Errorf("prepare application: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return submission, nil
}

func (s *ApplicationService) attachResume(
	ctx context.Context,
	application *applicationModel.Application,
) (applicationTypes.Submission, error) {
	resumes, err := s.resumeRepo.ListByCandidate(ctx, application.CandidateID)
	if err != nil {
		return applicationTypes.Submission{}, err
	}

	for _, resume := range resumes {
		if !resume.IsActive {
			continue
		}

		content, err := s.fileStore.Read(ctx, resume.StorageKey)
		if err != nil {
			return applicationTypes.Submission{}, err
		}

		application.ResumeID = &resume.ID

		return applicationTypes.Submission{
			Resume:     content,
			ResumeName: resume.FileName,
			ResumeType: resume.ContentType,
		}, nil
	}

	appendLog(application, "NO_RESUME", config.ErrNoActiveResume.Error())

	return applicationTypes.Submission{}, nil
}

// Transient upstream errors are retried by the queue; anything else is handed to the candidate.
func (s *ApplicationService) handleUpstreamError(
	ctx context.Context,
	application *applicationModel.Application,
	cause error,
	step string,
) *config.CustomError {
	if isTransientUpstreamError(cause) {
		return s.releaseForRetry(ctx, application, cause)
	}

	reason := cause.Error()
	application.Status = modelHelpers.ApplicationStatusActionRequired
	application.FailureReason = &reason
	appendLog(application, step, reason)

	return s.saveApplication(ctx, application, "prepare application")
}

// handleSubmitError retries nothing: after a timeout or a server error the employer may already have
// the application, so the candidate is asked to check instead of risking a duplicate.
func (s *ApplicationService) handleSubmitError(
	ctx context.Context,
	application *applicationModel.Application,
	cause error,
) *config.CustomError {
	if isTransientUpstreamError(cause) {
		return s.markOutcomeUnknown(
			ctx,
			application,
			fmt.Errorf("%w: %w", config.ErrSubmissionOutcomeUnknown, cause),
		)
	}

	return s.handleUpstreamError(ctx, application, cause, "SUBMISSION_REJECTED")
}

func (s *ApplicationService) markOutcomeUnknown(
	ctx context.Context,
	application *applicationModel.Application,
	cause error,
) *config.CustomError {
	reason := cause.Error()
	application.Status = modelHelpers.ApplicationStatusActionRequired
	application.FailureReason = &reason
	appendLog(application, "SUBMISSION_UNCONFIRMED", reason)

	return s.saveApplication(ctx, application, "prepare application")
}

// releaseForRetry hands the claim back before a retryable failure, so the retry can claim it.
func (s *ApplicationService) releaseForRetry(
	ctx context.Context,
	application *applicationModel.Application,
	cause error,
) *config.CustomError {
	if err := s.applicationRepo.Release(ctx, application.ID, utils.GetCurrTimeStamp()); err != nil {
		utils.LogWarn(
			ctx,
			"application claim release failed",
			"application_id",
			application.ID,
			"error",
			err.Error(),
		)
	}

	return config.NewCustomError(
		fmt.Errorf("prepare application: %w", cause),
		config.GetErrCodeServiceUnavailable(),
	)
}

func isTransientUpstreamError(err error) bool {
	var apiErr *utils.APIError

	return !errors.As(err, &apiErr) || apiErr.IsTransient()
}

// hasUnconfirmedSubmission reports a submission that was started but never recorded as done.
func hasUnconfirmedSubmission(application *applicationModel.Application) bool {
	entries := []applicationTypes.ExecutionLogEntry{}
	_ = json.Unmarshal(application.ExecutionLogs, &entries)

	submitting := false

	for _, entry := range entries {
		switch entry.Step {
		case config.GetApplicationStepSubmitting():
			submitting = true
		case config.GetApplicationStepSubmitted():
			submitting = false
		}
	}

	return submitting
}

func (s *ApplicationService) saveApplication(
	ctx context.Context,
	application *applicationModel.Application,
	operation string,
) *config.CustomError {
	application.UpdateTimestamp()

	if err := application.Validate(); err != nil {
		return config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if _, err := s.applicationRepo.Upsert(ctx, application); err != nil {
		return config.NewCustomError(
			fmt.Errorf("%s: %w", operation, err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}

func appendLog(application *applicationModel.Application, step, detail string) {
	entries := []applicationTypes.ExecutionLogEntry{}
	_ = json.Unmarshal(application.ExecutionLogs, &entries)

	entries = append(entries, applicationTypes.ExecutionLogEntry{
		At:     utils.GetCurrTimeStamp(),
		Step:   step,
		Detail: detail,
	})

	if encoded, err := json.Marshal(entries); err == nil {
		application.ExecutionLogs = encoded
	}
}

func hasMissingAnswers(application *applicationModel.Application) bool {
	missing := []applicationTypes.MissingAnswer{}
	_ = json.Unmarshal(application.MissingAnswers, &missing)

	return len(missing) > 0
}

func nonEmpty(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

// claimExpiry is how long a claim protects a run: twice the task timeout, after which that run is dead.
func claimExpiry() time.Duration {
	taskConfig, _ := workerConfigs.GetTaskConfig(workerConfigs.TaskPrepareApplication)

	return 2 * taskConfig.Timeout
}
