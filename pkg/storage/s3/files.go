package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

func (s *S3Storage) Save(ctx context.Context, key string, content []byte) error {
	if err := utils.ValidateStorageKey(key); err != nil {
		return err
	}

	input := &awsS3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(content),
		ContentLength: aws.Int64(int64(len(content))),
		ContentType:   aws.String(http.DetectContentType(content)),
	}
	if s.serverSideEncryption != "" {
		input.ServerSideEncryption = s.serverSideEncryption
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	return nil
}

func (s *S3Storage) Read(ctx context.Context, key string) ([]byte, error) {
	if err := utils.ValidateStorageKey(key); err != nil {
		return nil, err
	}

	output, err := s.client.GetObject(ctx, &awsS3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if isNotFound(err) {
		return nil, fmt.Errorf("read file: %w", config.ErrStoredFileNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	defer func() { _ = output.Body.Close() }()

	content, err := io.ReadAll(io.LimitReader(output.Body, config.GetMaxStoredObjectBytes()+1))
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if int64(len(content)) > config.GetMaxStoredObjectBytes() {
		return nil, fmt.Errorf("read file: %w", config.ErrStoredFileTooLarge)
	}

	return content, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if err := utils.ValidateStorageKey(key); err != nil {
		return err
	}

	_, err := s.client.DeleteObject(ctx, &awsS3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

func isNotFound(err error) bool {
	var apiErr smithy.APIError

	return errors.As(err, &apiErr) && config.IsStorageNotFoundCode(apiErr.ErrorCode())
}
