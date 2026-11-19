package config

import (
	"io/fs"
	"time"
)

type sharedConstants struct {
	tokenTypeBearer         string
	apiKeySeparator         string
	stuckWorkReconcilerLock string
	openAPISpecPath         string
	swaggerUIPath           string
	defaultTimezone         string
	emptyJSONObject         string
	uniqueViolationCode     string
	foreignKeyViolationCode string
	argonEncodingFormat     string
	maxErrorBodyBytes       int64
	maxResponseBodyBytes    int64
	cronClaimHoldPercent    time.Duration
	taskIDFormat            string
	redactedEmail           string
	dailyCounterTTL         time.Duration
	storageDirectoryMode    fs.FileMode
	storageFileMode         fs.FileMode
}

type headerConstants struct {
	contentType          string
	accept               string
	cacheControl         string
	jsonContentType      string
	jsonContentTypeUTF8  string
	noCache              string
	corsAllowOrigin      string
	corsAllowCredentials string
	corsAllowMethods     string
	corsAllowHeaders     string
	corsMaxAge           string
	corsExposeHeaders    string
	vary                 string
	trueValue            string
}

type documentConstants struct {
	pdfExtension       string
	docxExtension      string
	pdfMagic           string
	zipMagic           string
	docxBodyPath       string
	docxTextElement    string
	docxParagraph      string
	docxTab            string
	docxLineBreak      string
	fallbackFileName   string
	maxFileNameLength  int
	documentReadFactor int
	docxMaxBodyBytes   int64
	pdfMaxPages        int
	pdfMaxDecodedBytes int64
	pdfWordGapRatio    float64
	pdfLineShiftRatio  float64
}

type calibrationConstants struct {
	badPercentile  float64
	goodPercentile float64
	targetScore    float64
}

type validationConstants struct {
	linkedinHost        string
	linkedinProfilePath string
	atsTypeFieldName    string
	salaryMinFieldName  string
	defaultIssue        string
	issues              map[string]string
	parameterisedIssues map[string]string
}

type vendorConstants struct {
	apolloAPIKeyHeader       string
	apolloVerifiedStatus     string
	hunterAPIKeyHeader       string
	hunterVerifiedStatus     string
	contactOutTokenHeader    string
	signalHireAPIKeyHeader   string
	signalHireStatusSuccess  string
	signalHireChannelEmail   string
	signalHireChannelPhone   string
	signalHireSubTypeWork    string
	signalHireVerifiedRating map[string]struct{}
	signalHireMaxSearchSize  int
	reasonNoContactReturned  string
}

var sharedConstI = sharedConstants{
	tokenTypeBearer:         "Bearer",
	apiKeySeparator:         "_",
	stuckWorkReconcilerLock: "cron:stuck_work_reconciler",
	openAPISpecPath:         "/openapi.json",
	swaggerUIPath:           "/swagger/*any",
	defaultTimezone:         "UTC",
	emptyJSONObject:         `{}`,
	uniqueViolationCode:     "23505",
	foreignKeyViolationCode: "23503",
	argonEncodingFormat:     "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
	maxErrorBodyBytes:       2048,
	maxResponseBodyBytes:    64 << 20,
	cronClaimHoldPercent:    90,
	taskIDFormat:            "%s:%x",
	redactedEmail:           "[redacted]@$1",
	dailyCounterTTL:         48 * time.Hour,
	storageDirectoryMode:    0o750,
	storageFileMode:         0o600,
}

var headerConstI = headerConstants{
	contentType:          "Content-Type",
	accept:               "Accept",
	cacheControl:         "Cache-Control",
	jsonContentType:      "application/json",
	jsonContentTypeUTF8:  "application/json; charset=utf-8",
	noCache:              "no-cache",
	corsAllowOrigin:      "Access-Control-Allow-Origin",
	corsAllowCredentials: "Access-Control-Allow-Credentials",
	corsAllowMethods:     "Access-Control-Allow-Methods",
	corsAllowHeaders:     "Access-Control-Allow-Headers",
	corsMaxAge:           "Access-Control-Max-Age",
	corsExposeHeaders:    "Access-Control-Expose-Headers",
	vary:                 "Vary",
	trueValue:            "true",
}

var documentConstI = documentConstants{
	pdfExtension:       ".pdf",
	docxExtension:      ".docx",
	pdfMagic:           "%PDF-",
	zipMagic:           "PK\x03\x04",
	docxBodyPath:       "word/document.xml",
	docxTextElement:    "t",
	docxParagraph:      "p",
	docxTab:            "tab",
	docxLineBreak:      "br",
	fallbackFileName:   "resume",
	maxFileNameLength:  255,
	documentReadFactor: 4,
	docxMaxBodyBytes:   64 << 20,
	pdfMaxPages:        50,
	pdfMaxDecodedBytes: 32 << 20,
	pdfWordGapRatio:    0.15,
	pdfLineShiftRatio:  0.5,
}

var calibrationConstI = calibrationConstants{
	badPercentile:  0.50,
	goodPercentile: 0.25,
	targetScore:    0.80,
}

var validationConstI = validationConstants{
	linkedinHost:        "linkedin.com",
	linkedinProfilePath: "https://www.linkedin.com/in/",
	atsTypeFieldName:    "ATSType",
	salaryMinFieldName:  "SalaryMin",
	defaultIssue:        "is invalid",
	issues: map[string]string{
		"required":                   "is required",
		"email":                      "must be a valid email address",
		"uuid":                       "must be a valid UUID",
		"url":                        "must be a valid URL",
		"fqdn":                       "must be a valid domain name",
		"validate_timezone":          "must be a valid IANA time zone",
		"validate_linkedin_url":      "must be an https linkedin.com URL",
		"validate_external_url":      "must be a public https URL on a host that matches atsType",
		"validate_salary_range":      "must be greater than or equal to salaryMin",
		"validate_job_title":         "must contain letters or digits",
		"validate_json_object":       "must be a JSON object",
		"validate_resume_file":       "must be a PDF or DOCX document",
		"validate_role":              "must be a known role",
		"validate_api_client_grant":  "must be a grant API clients may hold",
		"validate_board_token":       "must be a valid identifier for boardType",
		"validate_board_types":       "is not allowed for this boardType",
		"validate_suppression_value": "must be an email address for EMAIL or a domain for DOMAIN",
		"required_if":                "is required in this case",
		"required_unless":            "is required in this case",
		"excluded_if":                "is not allowed in this case",
	},
	parameterisedIssues: map[string]string{
		"min":   "must be at least ",
		"max":   "must be at most ",
		"gte":   "must be greater than or equal to ",
		"lte":   "must be less than or equal to ",
		"gt":    "must be greater than ",
		"lt":    "must be less than ",
		"len":   "must have length ",
		"oneof": "must be one of: ",
	},
}

var vendorConstI = vendorConstants{
	apolloAPIKeyHeader:       "X-Api-Key",
	apolloVerifiedStatus:     "verified",
	hunterAPIKeyHeader:       "X-API-KEY",
	hunterVerifiedStatus:     "valid",
	contactOutTokenHeader:    "token",
	signalHireAPIKeyHeader:   "apikey",
	signalHireStatusSuccess:  "success",
	signalHireChannelEmail:   "email",
	signalHireChannelPhone:   "phone",
	signalHireSubTypeWork:    "work",
	signalHireVerifiedRating: map[string]struct{}{"100": {}, "verified": {}},
	signalHireMaxSearchSize:  100,
	reasonNoContactReturned:  "provider returned no contact",
}

func GetTokenTypeBearer() string          { return sharedConstI.tokenTypeBearer }
func GetTaskIDFormat() string             { return sharedConstI.taskIDFormat }
func GetRedactedEmailReplacement() string { return sharedConstI.redactedEmail }
func GetAPIKeySeparator() string          { return sharedConstI.apiKeySeparator }
func GetStuckWorkReconcilerLock() string  { return sharedConstI.stuckWorkReconcilerLock }
func GetOpenAPISpecPath() string          { return sharedConstI.openAPISpecPath }
func GetSwaggerUIPath() string            { return sharedConstI.swaggerUIPath }
func GetDefaultTimezone() string          { return sharedConstI.defaultTimezone }
func GetEmptyJSONObject() string          { return sharedConstI.emptyJSONObject }
func GetUniqueViolationCode() string      { return sharedConstI.uniqueViolationCode }
func GetForeignKeyViolationCode() string {
	return sharedConstI.foreignKeyViolationCode
}
func GetArgonEncodingFormat() string       { return sharedConstI.argonEncodingFormat }
func GetMaxErrorBodyBytes() int64          { return sharedConstI.maxErrorBodyBytes }
func GetMaxResponseBodyBytes() int64       { return sharedConstI.maxResponseBodyBytes }
func GetDailyCounterTTL() time.Duration    { return sharedConstI.dailyCounterTTL }
func GetStorageDirectoryMode() fs.FileMode { return sharedConstI.storageDirectoryMode }
func GetStorageFileMode() fs.FileMode      { return sharedConstI.storageFileMode }

func GetContentTypeHeader() string          { return headerConstI.contentType }
func GetAcceptHeader() string               { return headerConstI.accept }
func GetCacheControlHeader() string         { return headerConstI.cacheControl }
func GetJSONContentType() string            { return headerConstI.jsonContentType }
func GetJSONContentTypeUTF8() string        { return headerConstI.jsonContentTypeUTF8 }
func GetNoCacheValue() string               { return headerConstI.noCache }
func GetCORSAllowOriginHeader() string      { return headerConstI.corsAllowOrigin }
func GetCORSAllowCredentialsHeader() string { return headerConstI.corsAllowCredentials }
func GetCORSAllowMethodsHeader() string     { return headerConstI.corsAllowMethods }
func GetCORSAllowHeadersHeader() string     { return headerConstI.corsAllowHeaders }
func GetCORSMaxAgeHeader() string           { return headerConstI.corsMaxAge }
func GetCORSExposeHeadersHeader() string    { return headerConstI.corsExposeHeaders }
func GetVaryHeader() string                 { return headerConstI.vary }
func GetTrueValue() string                  { return headerConstI.trueValue }

func GetPDFExtension() string      { return documentConstI.pdfExtension }
func GetDOCXExtension() string     { return documentConstI.docxExtension }
func GetPDFMagic() []byte          { return []byte(documentConstI.pdfMagic) }
func GetZipMagic() []byte          { return []byte(documentConstI.zipMagic) }
func GetDocxBodyPath() string      { return documentConstI.docxBodyPath }
func GetDocxTextElement() string   { return documentConstI.docxTextElement }
func GetDocxParagraph() string     { return documentConstI.docxParagraph }
func GetDocxTab() string           { return documentConstI.docxTab }
func GetDocxLineBreak() string     { return documentConstI.docxLineBreak }
func GetFallbackFileName() string  { return documentConstI.fallbackFileName }
func GetMaxFileNameLength() int    { return documentConstI.maxFileNameLength }
func GetDocumentReadFactor() int   { return documentConstI.documentReadFactor }
func GetDocxMaxBodyBytes() int64   { return documentConstI.docxMaxBodyBytes }
func GetPDFMaxPages() int          { return documentConstI.pdfMaxPages }
func GetPDFMaxDecodedBytes() int64 { return documentConstI.pdfMaxDecodedBytes }

// A gap wider than this share of the font size between two glyphs is a space between words.
func GetPDFWordGapRatio() float64 { return documentConstI.pdfWordGapRatio }

// A vertical move larger than this share of the font size starts a new line.
func GetPDFLineShiftRatio() float64 { return documentConstI.pdfLineShiftRatio }

func GetCalibrationBadPercentile() float64  { return calibrationConstI.badPercentile }
func GetCalibrationGoodPercentile() float64 { return calibrationConstI.goodPercentile }
func GetCalibrationTargetScore() float64    { return calibrationConstI.targetScore }

func GetLinkedinHost() string           { return validationConstI.linkedinHost }
func GetLinkedinProfilePath() string    { return validationConstI.linkedinProfilePath }
func GetATSTypeFieldName() string       { return validationConstI.atsTypeFieldName }
func GetSalaryMinFieldName() string     { return validationConstI.salaryMinFieldName }
func GetDefaultValidationIssue() string { return validationConstI.defaultIssue }

func GetValidationIssue(tag string) (string, bool) {
	issue, found := validationConstI.issues[tag]

	return issue, found
}

func GetParameterisedValidationIssue(tag string) (string, bool) {
	prefix, found := validationConstI.parameterisedIssues[tag]

	return prefix, found
}

func GetApolloAPIKeyHeader() string        { return vendorConstI.apolloAPIKeyHeader }
func GetApolloVerifiedEmailStatus() string { return vendorConstI.apolloVerifiedStatus }
func GetHunterAPIKeyHeader() string        { return vendorConstI.hunterAPIKeyHeader }
func GetHunterVerifiedEmailStatus() string { return vendorConstI.hunterVerifiedStatus }
func GetContactOutTokenHeader() string     { return vendorConstI.contactOutTokenHeader }
func GetSignalHireAPIKeyHeader() string    { return vendorConstI.signalHireAPIKeyHeader }
func GetSignalHireStatusSuccess() string   { return vendorConstI.signalHireStatusSuccess }
func GetSignalHireChannelEmail() string    { return vendorConstI.signalHireChannelEmail }
func GetSignalHireChannelPhone() string    { return vendorConstI.signalHireChannelPhone }
func GetSignalHireSubTypeWork() string     { return vendorConstI.signalHireSubTypeWork }
func GetSignalHireMaxSearchSize() int      { return vendorConstI.signalHireMaxSearchSize }
func GetReasonNoContactReturned() string   { return vendorConstI.reasonNoContactReturned }

func IsSignalHireVerifiedRating(rating string) bool {
	_, verified := vendorConstI.signalHireVerifiedRating[rating]

	return verified
}

// Ends a claim just before the next tick so a replica with a fast clock can take it.
func GetCronClaimHold(interval time.Duration) time.Duration {
	return interval * sharedConstI.cronClaimHoldPercent / 100
}

var reservedNetworksI = []string{
	"0.0.0.0/8",
	"100.64.0.0/10",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"240.0.0.0/4",
	"64:ff9b::/96",
	"64:ff9b:1::/48",
	"100::/64",
	"2001:db8::/32",
	"2002::/16",
}

func GetReservedNetworks() []string { return append([]string(nil), reservedNetworksI...) }
