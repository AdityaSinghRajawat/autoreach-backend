package config

import "time"

type appConstants struct {
	serviceName  string
	modeServer   string
	modeWorker   string
	envFile      string
	envProd      string
	envLocal     string
	envTest      string
	metricsSpace string
}

type httpConstants struct {
	correlationIDHeader   string
	internalSecretHeader  string
	retryAfterHeader      string
	rateLimitHeader       string
	rateLimitPolicyHeader string
	originHeader          string
	apiBasePath           string
	corsAllowedMethods    string
	corsAllowedHeaders    string
	corsMaxAgeSeconds     string
	corsExposedHeaders    string
	defaultPage           int
	defaultPerPage        int
	maxPerPage            int
}

type dbConstants struct {
	connectTimeout           time.Duration
	maxConnLifetime          time.Duration
	healthCheckPeriod        time.Duration
	idleInTransactionTimeout time.Duration
	extensionSchema          string
}

type redisConstants struct {
	connectTimeout     time.Duration
	keyDelimiter       string
	keyPrefix          string
	rateLimitNamespace string
}

type loggerConstants struct {
	maxSizeMB  int
	maxBackups int
	maxAgeDays int
}

type authConstants struct {
	authorizationHeader string
	bearerPrefix        string
	principalContextKey string
	candidateIDParam    string
	tokenIssuer         string
	refreshTokenBytes   int
	argonMemoryKiB      uint32
	argonIterations     uint32
	argonParallelism    uint8
	argonSaltLength     uint32
	argonKeyLength      uint32
	minPasswordLength   int
}

type jobConstants struct {
	jobIDParam         string
	clientIDParam      string
	apiKeyHeader       string
	apiKeyPrefix       string
	apiKeyPrefixBytes  int
	apiKeySecretBytes  int
	maxIngestBatchSize int
	hashVersion        string
	lockNamespace      string
	seatNamespace      string
	modeAdminCreate    string
}

type constants struct {
	job    jobConstants
	auth   authConstants
	app    appConstants
	http   httpConstants
	db     dbConstants
	redis  redisConstants
	logger loggerConstants
}

var constI = constants{
	job: jobConstants{
		jobIDParam:         "jobId",
		clientIDParam:      "clientId",
		apiKeyHeader:       "X-API-Key",
		apiKeyPrefix:       "ark",
		apiKeyPrefixBytes:  6,
		apiKeySecretBytes:  32,
		maxIngestBatchSize: 500,
		hashVersion:        "v1",
		lockNamespace:      "lock",
		seatNamespace:      "seat",
		modeAdminCreate:    "create-admin",
	},
	auth: authConstants{
		authorizationHeader: "Authorization",
		bearerPrefix:        "Bearer ",
		principalContextKey: "principal",
		candidateIDParam:    "candidateId",
		tokenIssuer:         "autoreach",
		refreshTokenBytes:   32,
		argonMemoryKiB:      19456,
		argonIterations:     2,
		argonParallelism:    1,
		argonSaltLength:     16,
		argonKeyLength:      32,
		minPasswordLength:   10,
	},
	app: appConstants{
		serviceName:  "autoreach",
		modeServer:   "server",
		modeWorker:   "worker",
		envFile:      ".env",
		envProd:      "production",
		envLocal:     "local",
		envTest:      "test",
		metricsSpace: "autoreach",
	},
	http: httpConstants{
		correlationIDHeader:   "X-Correlation-ID",
		internalSecretHeader:  "X-Internal-Secret",
		retryAfterHeader:      "Retry-After",
		rateLimitHeader:       "RateLimit",
		rateLimitPolicyHeader: "RateLimit-Policy",
		originHeader:          "Origin",
		apiBasePath:           "/api/v1",
		corsAllowedMethods:    "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		corsAllowedHeaders:    "Authorization, Content-Type, X-Correlation-ID, X-API-Key",
		corsMaxAgeSeconds:     "600",
		corsExposedHeaders:    "RateLimit, RateLimit-Policy, Retry-After, X-Correlation-ID",
		defaultPage:           1,
		defaultPerPage:        20,
		maxPerPage:            100,
	},
	db: dbConstants{
		connectTimeout:           10 * time.Second,
		maxConnLifetime:          30 * time.Minute,
		healthCheckPeriod:        30 * time.Second,
		idleInTransactionTimeout: 60 * time.Second,
		extensionSchema:          "public",
	},
	redis: redisConstants{
		connectTimeout:     5 * time.Second,
		keyDelimiter:       ":",
		keyPrefix:          "autoreach",
		rateLimitNamespace: "rate_limit",
	},
	logger: loggerConstants{
		maxSizeMB:  100,
		maxBackups: 14,
		maxAgeDays: 30,
	},
}

func GetServiceName() string      { return constI.app.serviceName }
func GetModeServer() string       { return constI.app.modeServer }
func GetModeWorker() string       { return constI.app.modeWorker }
func GetMetricsNamespace() string { return constI.app.metricsSpace }

func GetCorrelationIDHeader() string  { return constI.http.correlationIDHeader }
func GetInternalSecretHeader() string { return constI.http.internalSecretHeader }
func GetRetryAfterHeader() string     { return constI.http.retryAfterHeader }
func GetRateLimitHeader() string      { return constI.http.rateLimitHeader }
func GetRateLimitPolicyHeader() string {
	return constI.http.rateLimitPolicyHeader
}
func GetOriginHeader() string       { return constI.http.originHeader }
func GetAPIBasePath() string        { return constI.http.apiBasePath }
func GetCORSAllowedMethods() string { return constI.http.corsAllowedMethods }
func GetCORSAllowedHeaders() string { return constI.http.corsAllowedHeaders }
func GetCORSMaxAgeSeconds() string  { return constI.http.corsMaxAgeSeconds }
func GetCORSExposedHeaders() string { return constI.http.corsExposedHeaders }
func GetDefaultPage() int           { return constI.http.defaultPage }
func GetDefaultPerPage() int        { return constI.http.defaultPerPage }
func GetMaxPerPage() int            { return constI.http.maxPerPage }

func GetDBConnectTimeout() time.Duration           { return constI.db.connectTimeout }
func GetDBMaxConnLifetime() time.Duration          { return constI.db.maxConnLifetime }
func GetDBHealthCheckPeriod() time.Duration        { return constI.db.healthCheckPeriod }
func GetDBIdleInTransactionTimeout() time.Duration { return constI.db.idleInTransactionTimeout }

func GetRedisConnectTimeout() time.Duration { return constI.redis.connectTimeout }
func GetKeyDelimiter() string               { return constI.redis.keyDelimiter }
func GetKeyPrefix() string                  { return constI.redis.keyPrefix }
func GetRateLimitNamespace() string         { return constI.redis.rateLimitNamespace }

func GetLogMaxSizeMB() int  { return constI.logger.maxSizeMB }
func GetLogMaxBackups() int { return constI.logger.maxBackups }
func GetLogMaxAgeDays() int { return constI.logger.maxAgeDays }

func GetAuthorizationHeader() string { return constI.auth.authorizationHeader }
func GetBearerPrefix() string        { return constI.auth.bearerPrefix }
func GetPrincipalContextKey() string { return constI.auth.principalContextKey }
func GetCandidateIDParam() string    { return constI.auth.candidateIDParam }
func GetTokenIssuer() string         { return constI.auth.tokenIssuer }
func GetRefreshTokenBytes() int      { return constI.auth.refreshTokenBytes }
func GetArgonMemoryKiB() uint32      { return constI.auth.argonMemoryKiB }
func GetArgonIterations() uint32     { return constI.auth.argonIterations }
func GetArgonParallelism() uint8     { return constI.auth.argonParallelism }
func GetArgonSaltLength() uint32     { return constI.auth.argonSaltLength }
func GetArgonKeyLength() uint32      { return constI.auth.argonKeyLength }
func GetMinPasswordLength() int      { return constI.auth.minPasswordLength }

func GetJobIDParam() string      { return constI.job.jobIDParam }
func GetClientIDParam() string   { return constI.job.clientIDParam }
func GetAPIKeyHeader() string    { return constI.job.apiKeyHeader }
func GetAPIKeyPrefix() string    { return constI.job.apiKeyPrefix }
func GetAPIKeyPrefixBytes() int  { return constI.job.apiKeyPrefixBytes }
func GetAPIKeySecretBytes() int  { return constI.job.apiKeySecretBytes }
func GetMaxIngestBatchSize() int { return constI.job.maxIngestBatchSize }
func GetJobHashVersion() string  { return constI.job.hashVersion }
func GetLockNamespace() string   { return constI.job.lockNamespace }
func GetSeatNamespace() string   { return constI.job.seatNamespace }
func GetModeAdminCreate() string { return constI.job.modeAdminCreate }
