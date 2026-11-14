package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type appConfig struct {
	env      string
	port     string
	logDir   string
	logLevel string
}

type httpConfig struct {
	readHeaderTimeout  time.Duration
	readTimeout        time.Duration
	writeTimeout       time.Duration
	idleTimeout        time.Duration
	shutdownTimeout    time.Duration
	maxBodyBytes       int64
	corsAllowedOrigins []string
	rateLimitRequests  int
	rateLimitWindow    time.Duration
}

type dbConfig struct {
	host             string
	port             string
	user             string
	password         string
	name             string
	schema           string
	sslMode          string
	maxConns         int
	statementTimeout time.Duration
	lockTimeout      time.Duration
}

type redisConfig struct {
	addr     string
	password string
	database int
}

type secretConfig struct {
	internalSharedSecret string
	jwtSecret            string
	adminPassword        string
}

type authConfig struct {
	accessTokenTTL    time.Duration
	refreshTokenTTL   time.Duration
	rateLimitRequests int
	rateLimitWindow   time.Duration
}

type concurrencyConfig struct {
	backgroundMaxGoroutines int
	backgroundDrainTimeout  time.Duration
	concurrentTaskLimit     int
}

type workerConfig struct {
	provider             string
	concurrency          int
	dbMaxConns           int
	shutdownTimeout      time.Duration
	embedGroupMaxSize    int
	embedGroupMaxDelay   time.Duration
	embedGroupGrace      time.Duration
	cronStopGracePeriod  time.Duration
	reconcilerInterval   time.Duration
	reconcilerStaleAfter time.Duration
	reconcilerBatchSize  int
}

type openAIConfig struct {
	apiKey             string
	baseURL            string
	embeddingModel     string
	embeddingDimension int
	timeout            time.Duration
	maxInputChars      int
}

type matchConfig struct {
	storageLocalDir    string
	resumeMaxBytes     int64
	resumeMaxTextChars int
	topK               int
	efSearch           int
	minStoredScore     float64
	calibrationLow     float64
	calibrationHigh    float64
}

type envConfig struct {
	match       matchConfig
	concurrency concurrencyConfig
	worker      workerConfig
	openAI      openAIConfig
	app         appConfig
	http        httpConfig
	db          dbConfig
	redis       redisConfig
	secret      secretConfig
	auth        authConfig
}

const (
	minJWTSecretLength  = 32
	minGroupGracePeriod = time.Second
)

var (
	envConfigI    envConfig
	envConfigOnce sync.Once
)

func Init() {
	envConfigOnce.Do(func() {
		_ = godotenv.Load(constI.app.envFile)
		envConfigI = loadEnvConfig()

		loadTableNames()
		loadEnrichmentConfig()
		enforceProductionRequirements()
	})
}

func loadEnvConfig() envConfig {
	workerConcurrency := getEnvIntWithDefault("WORKER_CONCURRENCY", 10)

	return envConfig{
		match: matchConfig{
			storageLocalDir:    getEnvWithDefault("STORAGE_LOCAL_DIR", "./data/uploads"),
			resumeMaxBytes:     getEnvInt64WithDefault("RESUME_MAX_BYTES", 4<<20),
			resumeMaxTextChars: getEnvIntWithDefault("RESUME_MAX_TEXT_CHARS", 200000),
			topK:               max(getEnvIntWithDefault("MATCH_TOP_K", 50), 1),
			efSearch:           getEnvIntWithDefault("MATCH_EF_SEARCH", 100),
			minStoredScore:     getEnvFloatWithDefault("MATCH_MIN_STORED_SCORE", 0.50),
			calibrationLow:     getEnvFloatWithDefault("MATCH_CALIBRATION_LOW", 0),
			calibrationHigh:    getEnvFloatWithDefault("MATCH_CALIBRATION_HIGH", 1),
		},
		concurrency: concurrencyConfig{
			backgroundMaxGoroutines: max(getEnvIntWithDefault("BACKGROUND_MAX_GOROUTINES", 64), 1),
			backgroundDrainTimeout: getEnvDurationWithDefault(
				"BACKGROUND_DRAIN_TIMEOUT",
				15*time.Second,
			),
			concurrentTaskLimit: max(getEnvIntWithDefault("CONCURRENT_TASK_LIMIT", 16), 1),
		},
		worker: workerConfig{
			provider:    getEnvWithDefault("WORKER_MESSAGE_PROVIDER", "asynq"),
			concurrency: workerConcurrency,
			dbMaxConns:  getEnvIntWithDefault("WORKER_DB_MAX_CONNS", workerConcurrency+5),
			shutdownTimeout: getEnvDurationWithDefault(
				"WORKER_SHUTDOWN_TIMEOUT",
				25*time.Second,
			),
			embedGroupMaxSize: getEnvIntWithDefault("WORKER_EMBED_GROUP_MAX_SIZE", 64),
			embedGroupMaxDelay: getEnvDurationWithDefault(
				"WORKER_EMBED_GROUP_MAX_DELAY",
				10*time.Second,
			),
			embedGroupGrace: max(
				getEnvDurationWithDefault("WORKER_EMBED_GROUP_GRACE_PERIOD", 2*time.Second),
				minGroupGracePeriod,
			),
			cronStopGracePeriod: getEnvDurationWithDefault(
				"CRON_STOP_GRACE_PERIOD",
				20*time.Second,
			),
			reconcilerInterval: getEnvDurationWithDefault("RECONCILER_INTERVAL", 5*time.Minute),
			reconcilerStaleAfter: getEnvDurationWithDefault(
				"RECONCILER_STALE_AFTER",
				5*time.Minute,
			),
			reconcilerBatchSize: getEnvIntWithDefault("RECONCILER_BATCH_SIZE", 500),
		},
		openAI: openAIConfig{
			apiKey:  getEnvWithDefault("OPENAI_API_KEY", ""),
			baseURL: getEnvWithDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			embeddingModel: getEnvWithDefault(
				"OPENAI_EMBEDDING_MODEL",
				"text-embedding-3-small",
			),
			embeddingDimension: getEnvIntWithDefault("OPENAI_EMBEDDING_DIMENSIONS", 1536),
			timeout:            getEnvDurationWithDefault("OPENAI_TIMEOUT", 30*time.Second),
			maxInputChars:      getEnvIntWithDefault("OPENAI_EMBEDDING_MAX_INPUT_CHARS", 24000),
		},
		app: appConfig{
			env:      getEnvWithDefault("APP_ENV", constI.app.envLocal),
			port:     getEnvWithDefault("PORT", "8080"),
			logDir:   getEnvWithDefault("LOG_DIR", ""),
			logLevel: getEnvWithDefault("LOG_LEVEL", "info"),
		},
		http: httpConfig{
			readHeaderTimeout: getEnvDurationWithDefault(
				"HTTP_READ_HEADER_TIMEOUT",
				10*time.Second,
			),
			readTimeout:        getEnvDurationWithDefault("HTTP_READ_TIMEOUT", 30*time.Second),
			writeTimeout:       getEnvDurationWithDefault("HTTP_WRITE_TIMEOUT", 30*time.Second),
			idleTimeout:        getEnvDurationWithDefault("HTTP_IDLE_TIMEOUT", 120*time.Second),
			shutdownTimeout:    getEnvDurationWithDefault("HTTP_SHUTDOWN_TIMEOUT", 20*time.Second),
			maxBodyBytes:       getEnvInt64WithDefault("HTTP_MAX_BODY_BYTES", 5<<20),
			corsAllowedOrigins: getEnvListWithDefault("CORS_ALLOWED_ORIGINS", nil),
			rateLimitRequests:  getEnvIntWithDefault("HTTP_RATE_LIMIT_REQUESTS", 300),
			rateLimitWindow:    getEnvDurationWithDefault("HTTP_RATE_LIMIT_WINDOW", time.Minute),
		},
		db: dbConfig{
			host:             getEnvWithDefault("DB_HOST", "localhost"),
			port:             getEnvWithDefault("DB_PORT", "5432"),
			user:             getEnvWithDefault("DB_USER", "autoreach"),
			password:         getEnvWithDefault("DB_PASSWORD", ""),
			name:             getEnvWithDefault("DB_NAME", "autoreach"),
			schema:           getEnvWithDefault("DB_SCHEMA", "autoreach"),
			sslMode:          getEnvWithDefault("DB_SSLMODE", "disable"),
			maxConns:         getEnvIntWithDefault("DB_MAX_CONNS", 20),
			statementTimeout: getEnvDurationWithDefault("DB_STATEMENT_TIMEOUT", 15*time.Second),
			lockTimeout:      getEnvDurationWithDefault("DB_LOCK_TIMEOUT", 5*time.Second),
		},
		redis: redisConfig{
			addr:     getEnvWithDefault("REDIS_ADDR", "localhost:6379"),
			password: getEnvWithDefault("REDIS_PASSWORD", ""),
			database: getEnvIntWithDefault("REDIS_DB", 0),
		},
		secret: secretConfig{
			internalSharedSecret: getEnvWithDefault("INTERNAL_SHARED_SECRET", ""),
			jwtSecret:            getEnvWithDefault("JWT_SECRET", ""),
			adminPassword:        getEnvWithDefault("ADMIN_PASSWORD", ""),
		},
		auth: authConfig{
			accessTokenTTL:    getEnvDurationWithDefault("JWT_ACCESS_TTL", 15*time.Minute),
			refreshTokenTTL:   getEnvDurationWithDefault("JWT_REFRESH_TTL", 30*24*time.Hour),
			rateLimitRequests: getEnvIntWithDefault("AUTH_RATE_LIMIT_REQUESTS", 10),
			rateLimitWindow:   getEnvDurationWithDefault("AUTH_RATE_LIMIT_WINDOW", time.Minute),
		},
	}
}

func enforceProductionRequirements() {
	if len(envConfigI.secret.jwtSecret) < minJWTSecretLength {
		log.Fatalf("JWT_SECRET must be at least %d characters", minJWTSecretLength)
	}

	if !IsProduction() {
		return
	}

	required := map[string]string{
		"INTERNAL_SHARED_SECRET": envConfigI.secret.internalSharedSecret,
		"DB_PASSWORD":            envConfigI.db.password,
		"REDIS_PASSWORD":         envConfigI.redis.password,
	}
	for name, value := range required {
		if value == "" {
			log.Fatalf("%s is required in production", name)
		}
	}
}

func getEnvWithDefault(key, fallback string) string {
	value, found := os.LookupEnv(key)
	if !found || strings.TrimSpace(value) == "" {
		return fallback
	}

	return strings.TrimSpace(value)
}

func getEnvIntWithDefault(key string, fallback int) int {
	value, err := strconv.Atoi(getEnvWithDefault(key, ""))
	if err != nil {
		return fallback
	}

	return value
}

func getEnvInt64WithDefault(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(getEnvWithDefault(key, ""), 10, 64)
	if err != nil {
		return fallback
	}

	return value
}

func getEnvFloatWithDefault(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(getEnvWithDefault(key, ""), 64)
	if err != nil {
		return fallback
	}

	return value
}

func getEnvDurationWithDefault(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(getEnvWithDefault(key, ""))
	if err != nil {
		return fallback
	}

	return value
}

func getEnvListWithDefault(key string, fallback []string) []string {
	raw := getEnvWithDefault(key, "")
	if raw == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")

	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}

	return values
}

func GetAppEnv() string   { return envConfigI.app.env }
func IsProduction() bool  { return envConfigI.app.env == constI.app.envProd }
func IsLocal() bool       { return envConfigI.app.env == constI.app.envLocal }
func GetPort() string     { return envConfigI.app.port }
func GetLogDir() string   { return envConfigI.app.logDir }
func GetLogLevel() string { return envConfigI.app.logLevel }

func GetReadHeaderTimeout() time.Duration     { return envConfigI.http.readHeaderTimeout }
func GetReadTimeout() time.Duration           { return envConfigI.http.readTimeout }
func GetWriteTimeout() time.Duration          { return envConfigI.http.writeTimeout }
func GetIdleTimeout() time.Duration           { return envConfigI.http.idleTimeout }
func GetServerShutdownTimeout() time.Duration { return envConfigI.http.shutdownTimeout }
func GetMaxBodyBytes() int64                  { return envConfigI.http.maxBodyBytes }
func GetHTTPRateLimitRequests() int           { return envConfigI.http.rateLimitRequests }
func GetHTTPRateLimitWindow() time.Duration   { return envConfigI.http.rateLimitWindow }

func GetCORSAllowedOrigins() []string {
	return append([]string(nil), envConfigI.http.corsAllowedOrigins...)
}

func GetDBSchema() string                  { return envConfigI.db.schema }
func GetDBName() string                    { return envConfigI.db.name }
func GetDBMaxConns() int                   { return envConfigI.db.maxConns }
func GetDBStatementTimeout() time.Duration { return envConfigI.db.statementTimeout }
func GetDBLockTimeout() time.Duration      { return envConfigI.db.lockTimeout }

func GetDBDSN() string {
	query := url.Values{}
	query.Set("sslmode", envConfigI.db.sslMode)
	query.Set("search_path", envConfigI.db.schema+","+constI.db.extensionSchema)
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(envConfigI.db.user, envConfigI.db.password),
		Host:     envConfigI.db.host + ":" + envConfigI.db.port,
		Path:     "/" + envConfigI.db.name,
		RawQuery: query.Encode(),
	}

	return dsn.String()
}

func GetRedisAddr() string     { return envConfigI.redis.addr }
func GetRedisPassword() string { return envConfigI.redis.password }
func GetRedisDB() int          { return envConfigI.redis.database }

func GetInternalSharedSecret() string { return envConfigI.secret.internalSharedSecret }
func GetJWTSecret() []byte            { return []byte(envConfigI.secret.jwtSecret) }

func GetAccessTokenTTL() time.Duration      { return envConfigI.auth.accessTokenTTL }
func GetRefreshTokenTTL() time.Duration     { return envConfigI.auth.refreshTokenTTL }
func GetAuthRateLimitRequests() int         { return envConfigI.auth.rateLimitRequests }
func GetAuthRateLimitWindow() time.Duration { return envConfigI.auth.rateLimitWindow }

func GetWorkerMessageProvider() string        { return envConfigI.worker.provider }
func GetWorkerConcurrency() int               { return envConfigI.worker.concurrency }
func GetWorkerDBMaxConns() int                { return envConfigI.worker.dbMaxConns }
func GetWorkerShutdownTimeout() time.Duration { return envConfigI.worker.shutdownTimeout }
func GetEmbedGroupMaxSize() int               { return envConfigI.worker.embedGroupMaxSize }
func GetEmbedGroupMaxDelay() time.Duration    { return envConfigI.worker.embedGroupMaxDelay }
func GetEmbedGroupGracePeriod() time.Duration { return envConfigI.worker.embedGroupGrace }
func GetCronStopGracePeriod() time.Duration   { return envConfigI.worker.cronStopGracePeriod }
func GetReconcilerInterval() time.Duration    { return envConfigI.worker.reconcilerInterval }
func GetReconcilerStaleAfter() time.Duration  { return envConfigI.worker.reconcilerStaleAfter }
func GetReconcilerBatchSize() int             { return envConfigI.worker.reconcilerBatchSize }

func GetOpenAIAPIKey() string              { return envConfigI.openAI.apiKey }
func GetOpenAIBaseURL() string             { return envConfigI.openAI.baseURL }
func GetOpenAIEmbeddingModel() string      { return envConfigI.openAI.embeddingModel }
func GetOpenAIEmbeddingDimensions() int    { return envConfigI.openAI.embeddingDimension }
func GetOpenAITimeout() time.Duration      { return envConfigI.openAI.timeout }
func GetOpenAIEmbeddingMaxInputChars() int { return envConfigI.openAI.maxInputChars }

func GetBackgroundMaxGoroutines() int { return envConfigI.concurrency.backgroundMaxGoroutines }

func GetBackgroundDrainTimeout() time.Duration { return envConfigI.concurrency.backgroundDrainTimeout }

func GetConcurrentTaskLimit() int { return envConfigI.concurrency.concurrentTaskLimit }

func GetStorageLocalDir() string { return envConfigI.match.storageLocalDir }
func GetResumeMaxBytes() int64   { return envConfigI.match.resumeMaxBytes }
func GetResumeMaxTextChars() int { return envConfigI.match.resumeMaxTextChars }
func GetMatchTopK() int          { return envConfigI.match.topK }

func GetMatchEFSearch() int {
	return max(envConfigI.match.efSearch, envConfigI.match.topK)
}

func GetMatchMinStoredScore() float64 { return envConfigI.match.minStoredScore }

func GetAdminPassword() string { return envConfigI.secret.adminPassword }
