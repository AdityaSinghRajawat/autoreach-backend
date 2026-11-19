package candidateManagement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	candidateTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/candidateManagement"
	candidateModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/candidateManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *CandidateService) ProcessResume(
	ctx context.Context,
	resumeID string,
) (*candidateTypes.ProcessResumeResult, *config.CustomError) {
	resume, err := s.resumeRepo.FindByID(ctx, resumeID)

	resume, customErr := config.RequireFound(
		"process resume",
		resume,
		err,
		config.ErrResumeNotFound,
	)
	if customErr != nil {
		return nil, customErr
	}

	if resume.IsEmbedded() {
		return s.finishProcessedResume(ctx, resume)
	}

	if !resume.IsParsed() {
		text, err := s.extractResumeText(ctx, resume)
		if isTerminalDocumentError(err) {
			return s.failResume(ctx, resume, err)
		}

		if err != nil {
			return nil, config.NewCustomError(
				fmt.Errorf("process resume: %w", err),
				config.GetErrCodeInternalServerError(),
			)
		}

		if customErr := s.saveParsedResume(ctx, resume, text); customErr != nil {
			return nil, customErr
		}
	}

	embeddings, err := s.embedder.CreateEmbeddings(
		ctx,
		[]string{utils.TruncateRunes(resume.RawText, config.GetAIEmbeddingMaxInputChars())},
	)
	if err != nil {
		return s.handleEmbeddingFailure(ctx, resume, err)
	}

	if len(embeddings) != 1 {
		return nil, config.NewCustomError(
			fmt.Errorf("process resume: %w", config.ErrEmbeddingCountMismatch),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	if customErr := s.saveResumeEmbedding(ctx, resume, embeddings[0]); customErr != nil {
		return nil, customErr
	}

	return s.finishProcessedResume(ctx, resume)
}

func (s *CandidateService) extractResumeText(
	ctx context.Context,
	resume *candidateModel.Resume,
) (string, error) {
	content, err := s.fileStore.Read(ctx, resume.StorageKey)
	if err != nil {
		return "", err
	}

	maxChars := config.GetResumeMaxTextChars()

	var text string

	switch resume.ContentType {
	case modelHelpers.ContentTypePDF:
		text, err = utils.ExtractPDFText(content, maxChars)
	case modelHelpers.ContentTypeDOCX:
		text, err = utils.ExtractDocxText(content, maxChars)
	default:
		err = config.ErrUnreadableDocument
	}

	if err != nil {
		return "", err
	}

	if strings.TrimSpace(text) == "" {
		return "", config.ErrEmptyDocument
	}

	return text, nil
}

func (s *CandidateService) saveParsedResume(
	ctx context.Context,
	resume *candidateModel.Resume,
	text string,
) *config.CustomError {
	metadata, err := json.Marshal(s.parseResumeMetadata(ctx, text))
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("process resume: encode metadata: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	resume.MarkParsed(text, metadata)
	resume.UpdateTimestamp()

	if err := resume.Validate(); err != nil {
		return config.NewCustomError(
			fmt.Errorf("process resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	if _, err := s.resumeRepo.Upsert(ctx, resume); err != nil {
		return config.NewCustomError(
			fmt.Errorf("process resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	return nil
}

func (s *CandidateService) saveResumeEmbedding(
	ctx context.Context,
	resume *candidateModel.Resume,
	embedding []float32,
) *config.CustomError {
	resume.UpdateTimestamp()

	err := s.transactionManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.resumeRepo.UpdateEmbedding(
			ctx,
			resume.ID,
			embedding,
			s.embedder.Model(),
			resume.UpdatedAt,
		); err != nil {
			return err
		}

		_, err := s.resumeRepo.Activate(ctx, resume.CandidateID, resume.ID, resume.UpdatedAt)

		return err
	})
	if err != nil {
		return config.NewCustomError(
			fmt.Errorf("process resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	resume.Status = modelHelpers.ResumeStatusEmbedded
	resume.IsActive = true

	return nil
}

func (s *CandidateService) finishProcessedResume(
	ctx context.Context,
	resume *candidateModel.Resume,
) (*candidateTypes.ProcessResumeResult, *config.CustomError) {
	if err := s.publishMatchResume(ctx, resume.ID); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("process resume: publish match: %w", err),
			config.GetErrCodeServiceUnavailable(),
		)
	}

	return &candidateTypes.ProcessResumeResult{ResumeID: resume.ID, Status: resume.Status}, nil
}

func (s *CandidateService) handleEmbeddingFailure(
	ctx context.Context,
	resume *candidateModel.Resume,
	cause error,
) (*candidateTypes.ProcessResumeResult, *config.CustomError) {
	errCode := utils.ClassifyUpstreamError(cause)
	if !config.IsRetryableErrCode(errCode) {
		if _, customErr := s.failResume(ctx, resume, cause); customErr != nil {
			return nil, customErr
		}
	}

	return nil, config.NewCustomError(fmt.Errorf("process resume: %w", cause), errCode)
}

func (s *CandidateService) failResume(
	ctx context.Context,
	resume *candidateModel.Resume,
	cause error,
) (*candidateTypes.ProcessResumeResult, *config.CustomError) {
	resume.MarkFailed(failureReason(cause))
	resume.UpdateTimestamp()

	if _, err := s.resumeRepo.Upsert(ctx, resume); err != nil {
		return nil, config.NewCustomError(
			fmt.Errorf("process resume: %w", err),
			config.GetErrCodeInternalServerError(),
		)
	}

	utils.LogWarn(ctx, "resume processing failed", "resume_id", resume.ID, "error", cause.Error())

	return &candidateTypes.ProcessResumeResult{ResumeID: resume.ID, Status: resume.Status}, nil
}

func isTerminalDocumentError(err error) bool {
	return errors.Is(err, config.ErrUnreadableDocument) ||
		errors.Is(err, config.ErrEmptyDocument) ||
		errors.Is(err, config.ErrDocumentTooLarge) ||
		errors.Is(err, config.ErrStoredFileNotFound)
}

func failureReason(cause error) string {
	for _, known := range []error{
		config.ErrEmptyDocument,
		config.ErrDocumentTooLarge,
		config.ErrStoredFileNotFound,
		config.ErrUnreadableDocument,
	} {
		if errors.Is(cause, known) {
			return known.Error()
		}
	}

	return config.GetPublicMessageForErrCode(config.GetErrCodeUpstreamRejected())
}
