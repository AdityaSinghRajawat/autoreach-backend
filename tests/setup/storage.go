package setup

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

const (
	testS3EndpointEnv = "TEST_S3_ENDPOINT"
	testS3BucketEnv   = "TEST_S3_BUCKET"
	testS3Region      = "us-east-1"
	testS3Credential  = "test"
)

func S3Settings(t *testing.T) config.S3Settings {
	t.Helper()
	Config(t)

	settings := config.S3Settings{
		Bucket:               os.Getenv(testS3BucketEnv),
		Region:               testS3Region,
		Endpoint:             os.Getenv(testS3EndpointEnv),
		AccessKeyID:          testS3Credential,
		SecretAccessKey:      testS3Credential,
		UsePathStyle:         true,
		ServerSideEncryption: "",
		Timeout:              5 * time.Second,
	}
	if settings.Bucket == "" || settings.Endpoint == "" {
		t.Skip("TEST_S3_ENDPOINT and TEST_S3_BUCKET are not set")
	}

	if err := ensureBucket(settings); err != nil {
		if os.Getenv(requireServicesFlag) == "1" {
			require.NoError(t, err, "S3 test service is required but unavailable")
		}

		t.Skipf("S3 test service unavailable: %v", err)
	}

	return settings
}

func ensureBucket(settings config.S3Settings) error {
	ctx, cancel := context.WithTimeout(context.Background(), settings.Timeout)
	defer cancel()

	sdkConfig, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithRegion(settings.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			settings.AccessKeyID,
			settings.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return err
	}

	client := awsS3.NewFromConfig(sdkConfig, func(options *awsS3.Options) {
		options.UsePathStyle = true
		options.BaseEndpoint = aws.String(settings.Endpoint)
	})

	_, err = client.CreateBucket(ctx, &awsS3.CreateBucketInput{Bucket: aws.String(settings.Bucket)})

	var (
		owned  *types.BucketAlreadyOwnedByYou
		exists *types.BucketAlreadyExists
	)

	if err != nil && !errors.As(err, &owned) && !errors.As(err, &exists) {
		return err
	}

	return nil
}
