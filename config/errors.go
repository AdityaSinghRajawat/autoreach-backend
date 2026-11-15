package config

import "errors"

var (
	ErrMissingDependency     = errors.New("missing dependency")
	ErrMissingCredential     = errors.New("missing credential")
	ErrTaskPanicked          = errors.New("task panicked")
	ErrUnknownMode           = errors.New("unknown mode")
	ErrUnknownWorkerProvider = errors.New("unknown worker provider")
	ErrCreateAdminUsage      = errors.New(
		"usage: autoreach create-admin <email> (password is read from ADMIN_PASSWORD)",
	)
	ErrDuplicateRecord         = errors.New("duplicate record")
	ErrInvalidToken            = errors.New("invalid token")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrPasswordTooShort        = errors.New("password is too short")
	ErrUnsupportedPasswordHash = errors.New("unsupported password hash")
	ErrRefreshTokenUnknown     = errors.New("refresh token unknown")
	ErrRefreshTokenExpired     = errors.New("refresh token expired")
	ErrRefreshTokenReused      = errors.New("refresh token reused")
	ErrUserInactive            = errors.New("user missing or inactive")
	ErrPermissionDenied        = errors.New("permission denied")
	ErrResourceNotAccessible   = errors.New("resource not accessible")
	ErrInvalidGrant            = errors.New("invalid grant")
	ErrUnknownRole             = errors.New("unknown role")
	ErrPrincipalMissing        = errors.New("principal missing")
	ErrBearerTokenMissing      = errors.New("bearer token missing")
	ErrInternalSecretMismatch  = errors.New("internal secret mismatch")
	ErrAPIKeyMissing           = errors.New("api key missing")
	ErrAPIKeyInvalid           = errors.New("api key invalid")
	ErrAPIClientNotFound       = errors.New("api client not found")
	ErrCandidateNotFound       = errors.New("candidate not found")
	ErrResumeNotFound          = errors.New("resume not found")
	ErrResumeNotReady          = errors.New("resume is not embedded yet")
	ErrResumeFileMissing       = errors.New("resume file is missing")
	ErrResumeFileTooLarge      = errors.New("resume file is too large")
	ErrUnreadableDocument      = errors.New("document could not be read")
	ErrEmptyDocument           = errors.New("document has no extractable text")
	ErrDocumentTooLarge        = errors.New("document expands beyond the allowed size")
	ErrInvalidStorageKey       = errors.New("invalid storage key")
	ErrStoredFileTooLarge      = errors.New("stored file exceeds the size limit")
	ErrUnknownStorageDriver    = errors.New("unknown storage driver")
	ErrStorageUnavailable      = errors.New("file storage is unavailable")
	ErrStoredFileNotFound      = errors.New("stored file not found")
	ErrInvalidCalibrationPair  = errors.New(
		"calibration pair needs resume, job and a good or bad label",
	)
	ErrContactNotFound         = errors.New("contact not found")
	ErrJobNotMatched           = errors.New("job is not matched to the candidate")
	ErrProviderUnavailable     = errors.New("enrichment provider circuit is open")
	ErrProviderRateLimited     = errors.New("enrichment provider rate limit reached")
	ErrProviderBudgetExhausted = errors.New(
		"enrichment provider daily credit budget exhausted",
	)
	ErrNoDiscoveryProvider             = errors.New("no enrichment provider can discover people")
	ErrMalformedProviderPayload        = errors.New("malformed provider payload")
	ErrWebhookTokenMismatch            = errors.New("webhook token mismatch")
	ErrJobNotFound                     = errors.New("job not found")
	ErrNoFieldsSupplied                = errors.New("no fields supplied")
	ErrRouteNotFound                   = errors.New("route not found")
	ErrMethodNotAllowed                = errors.New("method not allowed")
	ErrRateLimitExceeded               = errors.New("rate limit exceeded")
	ErrModelReplyMissing               = errors.New("model returned no reply")
	ErrModelReplyIncomplete            = errors.New("model reply was refused or cut off")
	ErrCredentialEncryptionUnavailable = errors.New(
		"channel credential encryption is not configured",
	)
	ErrUnknownCredentialKey   = errors.New("credential key version is unknown")
	ErrInsecureMailNotAllowed = errors.New("unencrypted mail connections are not allowed")
	ErrChannelAccountNotFound = errors.New("channel account not found")
	ErrMailAuthUnsupported    = errors.New("mail server does not offer authentication")
	ErrTemplateNotFound       = errors.New("outreach template not found")
	ErrCampaignNotFound       = errors.New("outreach campaign not found")
	ErrUnknownPlaceholder     = errors.New("unknown placeholder")
	ErrSubjectRequired        = errors.New("an email template needs a subject")
	ErrTemplateInUse          = errors.New("template is used by a campaign")
	ErrCampaignStepTemplate   = errors.New(
		"campaign step template does not belong to this candidate",
	)
	ErrActiveCampaignExists   = errors.New("the candidate already has an active campaign")
	ErrMessageNotFound        = errors.New("outreach message not found")
	ErrMessageNotCancellable  = errors.New("only pending or paused messages can be cancelled")
	ErrContactNotReachable    = errors.New("contact cannot be reached on this channel for this job")
	ErrContactSuppressed      = errors.New("contact is on the suppression list")
	ErrNoActiveMailbox        = errors.New("no verified mailbox to send from")
	ErrChannelNotSupported    = errors.New("channel is not supported yet")
	ErrSendOutcomeUnknown     = errors.New("send outcome unknown after a worker interruption")
	ErrConversationEnded      = errors.New("the contact already replied or bounced")
	ErrInvalidTrackingLink    = errors.New("tracking link is invalid")
	ErrInvalidUnsubscribeLink = errors.New("unsubscribe link is invalid")
	ErrLinkedinNoteTooLong    = errors.New(
		"a LinkedIn connection note must be at most 300 characters",
	)
	ErrContactHasNoLinkedin    = errors.New("contact has no LinkedIn profile")
	ErrConnectionNotAccepted   = errors.New("the connection request was not accepted in time")
	ErrConfirmationNotAllowed  = errors.New("this message cannot move to that status")
	ErrUnknownLinkedinProvider = errors.New("unknown LinkedIn provider")
	ErrMessageFieldRequired    = errors.New("is required for the selected channel")
	ErrRateLimiterUnavailable  = errors.New("rate limiter is unavailable")
	ErrTrailingRequestData     = errors.New("unexpected trailing data")
	ErrUnexpectedRedisReply    = errors.New("unexpected redis reply")
	ErrEmbeddingCountMismatch  = errors.New(
		"embedder returned a different number of vectors",
	)
	ErrMalformedEmbedding = errors.New("malformed embedding")
	ErrNoValidMessages    = errors.New("no valid messages in group")
	ErrTaskFailed         = errors.New("task failed")
	ErrTaskConfigMissing  = errors.New("task config missing")
	ErrInvalidTopology    = errors.New("invalid worker topology")
)

type CustomErrorCode string

type ErrorDetail struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

type CustomError struct {
	Err     error
	ErrCode CustomErrorCode
	Details []ErrorDetail
}

func NewCustomError(err error, errCode CustomErrorCode) *CustomError {
	return &CustomError{Err: err, ErrCode: errCode}
}

func NewValidationError(err error, details []ErrorDetail) *CustomError {
	return &CustomError{Err: err, ErrCode: errCodesI.generic.validation, Details: details}
}

func (e *CustomError) Error() string {
	if e.Err == nil {
		return string(e.ErrCode)
	}

	return e.Err.Error()
}

func (e *CustomError) Unwrap() error {
	return e.Err
}

type genericErrCodes struct {
	internalServerError CustomErrorCode
	badRequest          CustomErrorCode
	validation          CustomErrorCode
	unauthorized        CustomErrorCode
	forbidden           CustomErrorCode
	notFound            CustomErrorCode
	methodNotAllowed    CustomErrorCode
	conflict            CustomErrorCode
	payloadTooLarge     CustomErrorCode
	rateLimited         CustomErrorCode
	serviceUnavailable  CustomErrorCode
	upstreamRejected    CustomErrorCode
}

type errCodes struct {
	generic genericErrCodes
}

var errCodesI = errCodes{
	generic: genericErrCodes{
		internalServerError: "INTERNAL_SERVER_ERROR",
		badRequest:          "BAD_REQUEST",
		validation:          "VALIDATION_ERROR",
		unauthorized:        "UNAUTHORIZED",
		forbidden:           "FORBIDDEN",
		notFound:            "NOT_FOUND",
		methodNotAllowed:    "METHOD_NOT_ALLOWED",
		conflict:            "CONFLICT",
		payloadTooLarge:     "PAYLOAD_TOO_LARGE",
		rateLimited:         "RATE_LIMITED",
		serviceUnavailable:  "SERVICE_UNAVAILABLE",
		upstreamRejected:    "UPSTREAM_REJECTED",
	},
}

var publicMessagesI = map[CustomErrorCode]string{
	errCodesI.generic.internalServerError: "An unexpected error occurred.",
	errCodesI.generic.badRequest:          "The request could not be understood.",
	errCodesI.generic.validation:          "The request payload failed validation checks.",
	errCodesI.generic.unauthorized:        "Authentication is required or has failed.",
	errCodesI.generic.forbidden:           "You do not have permission to perform this action.",
	errCodesI.generic.notFound:            "The requested resource was not found.",
	errCodesI.generic.methodNotAllowed:    "The HTTP method is not allowed for this resource.",
	errCodesI.generic.conflict:            "The request conflicts with the current state of the resource.",
	errCodesI.generic.payloadTooLarge:     "The request payload is too large.",
	errCodesI.generic.rateLimited:         "Too many requests. Please retry later.",
	errCodesI.generic.serviceUnavailable:  "The service is temporarily unavailable.",
	errCodesI.generic.upstreamRejected:    "An upstream provider rejected the request.",
}

func GetErrCodeInternalServerError() CustomErrorCode { return errCodesI.generic.internalServerError }
func GetErrCodeBadRequest() CustomErrorCode          { return errCodesI.generic.badRequest }
func GetErrCodeValidation() CustomErrorCode          { return errCodesI.generic.validation }
func GetErrCodeUnauthorized() CustomErrorCode        { return errCodesI.generic.unauthorized }
func GetErrCodeForbidden() CustomErrorCode           { return errCodesI.generic.forbidden }
func GetErrCodeNotFound() CustomErrorCode            { return errCodesI.generic.notFound }
func GetErrCodeMethodNotAllowed() CustomErrorCode    { return errCodesI.generic.methodNotAllowed }
func GetErrCodeConflict() CustomErrorCode            { return errCodesI.generic.conflict }
func GetErrCodePayloadTooLarge() CustomErrorCode     { return errCodesI.generic.payloadTooLarge }
func GetErrCodeRateLimited() CustomErrorCode         { return errCodesI.generic.rateLimited }
func GetErrCodeServiceUnavailable() CustomErrorCode  { return errCodesI.generic.serviceUnavailable }

func GetErrCodeUpstreamRejected() CustomErrorCode { return errCodesI.generic.upstreamRejected }

func GetAllErrCodes() []CustomErrorCode {
	codes := make([]CustomErrorCode, 0, len(publicMessagesI))
	for code := range publicMessagesI {
		codes = append(codes, code)
	}

	return codes
}

func GetPublicMessageForErrCode(errCode CustomErrorCode) string {
	message, found := publicMessagesI[errCode]
	if !found {
		return publicMessagesI[errCodesI.generic.internalServerError]
	}

	return message
}
