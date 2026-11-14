package config

import (
	"strings"
	"time"
)

const (
	StorageDriverLocal = "local"
	StorageDriverS3    = "s3"

	defaultS3Region       = "us-east-1"
	maxStoredObjectBytes  = 32 << 20
	storageKeySeparator   = "/"
	storageKeyParentToken = ".."
)

var storageNotFoundCodesI = map[string]struct{}{
	"NoSuchKey": {},
	"NotFound":  {},
}

type S3Settings struct {
	Bucket               string
	Region               string
	Endpoint             string
	AccessKeyID          string
	SecretAccessKey      string
	UsePathStyle         bool
	ServerSideEncryption string
	Timeout              time.Duration
}

type storageConfig struct {
	driver   string
	localDir string
	s3       S3Settings
}

var storageConfigI storageConfig

func loadStorageConfig() {
	storageConfigI = storageConfig{
		driver:   strings.ToLower(getEnvWithDefault("STORAGE_DRIVER", StorageDriverLocal)),
		localDir: getEnvWithDefault("STORAGE_LOCAL_DIR", "./data/uploads"),
		s3: S3Settings{
			Bucket:               getEnvWithDefault("S3_BUCKET", ""),
			Region:               getEnvWithDefault("S3_REGION", defaultS3Region),
			Endpoint:             getEnvWithDefault("S3_ENDPOINT", ""),
			AccessKeyID:          getEnvWithDefault("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey:      getEnvWithDefault("S3_SECRET_ACCESS_KEY", ""),
			UsePathStyle:         getEnvBoolWithDefault("S3_USE_PATH_STYLE", false),
			ServerSideEncryption: getEnvWithDefault("S3_SERVER_SIDE_ENCRYPTION", "AES256"),
			Timeout:              getEnvDurationWithDefault("S3_TIMEOUT", 15*time.Second),
		},
	}
}

func GetStorageDriver() string { return storageConfigI.driver }

func GetStorageLocalDir() string       { return storageConfigI.localDir }
func GetS3Settings() S3Settings        { return storageConfigI.s3 }
func GetMaxStoredObjectBytes() int64   { return maxStoredObjectBytes }
func GetStorageKeySeparator() string   { return storageKeySeparator }
func GetStorageKeyParentToken() string { return storageKeyParentToken }

func IsStorageNotFoundCode(code string) bool {
	_, found := storageNotFoundCodesI[code]

	return found
}
