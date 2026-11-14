package s3

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
)

type S3Storage struct {
	client               *awsS3.Client
	bucket               string
	serverSideEncryption types.ServerSideEncryption
}

func NewS3Storage(ctx context.Context, settings config.S3Settings) (*S3Storage, error) {
	if settings.Bucket == "" {
		return nil, fmt.Errorf("new s3 storage: %w: S3_BUCKET", config.ErrMissingCredential)
	}

	loadOptions := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion(settings.Region),
		awsConfig.WithHTTPClient(&http.Client{Timeout: settings.Timeout}),
	}

	if settings.AccessKeyID != "" || settings.SecretAccessKey != "" {
		loadOptions = append(loadOptions, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				settings.AccessKeyID,
				settings.SecretAccessKey,
				"",
			),
		))
	}

	sdkConfig, err := awsConfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("new s3 storage: load aws config: %w", err)
	}

	client := awsS3.NewFromConfig(sdkConfig, func(options *awsS3.Options) {
		options.UsePathStyle = settings.UsePathStyle

		if settings.Endpoint != "" {
			options.BaseEndpoint = aws.String(settings.Endpoint)
		}
	})

	storage := &S3Storage{
		client:               client,
		bucket:               settings.Bucket,
		serverSideEncryption: types.ServerSideEncryption(settings.ServerSideEncryption),
	}

	if err := storage.checkBucket(ctx, settings); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *S3Storage) checkBucket(ctx context.Context, settings config.S3Settings) error {
	checkCtx, cancel := context.WithTimeout(ctx, settings.Timeout)
	defer cancel()

	_, err := s.client.HeadBucket(checkCtx, &awsS3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return fmt.Errorf(
			"new s3 storage: bucket %q: %w: %w",
			s.bucket,
			config.ErrStorageUnavailable,
			err,
		)
	}

	return nil
}
