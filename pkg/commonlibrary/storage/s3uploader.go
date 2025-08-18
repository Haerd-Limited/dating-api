package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	url2 "net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Uploader struct {
	Client     *s3.Client
	Uploader   *manager.Uploader
	BucketName string
}

func NewS3Uploader(bucketName string, region string) (*S3Uploader, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3Client)

	return &S3Uploader{
		Client:     s3Client,
		Uploader:   uploader,
		BucketName: bucketName,
	}, nil
}

func (u *S3Uploader) UploadFile(ctx context.Context, file multipart.File, filename string, contentType string) (string, error) {
	result, err := u.Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.BucketName),
		Key:         aws.String(filename),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return result.Location, nil
}

func (u *S3Uploader) DeleteFile(ctx context.Context, fileURL string) error {
	parsed, err := url2.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("invalid file URL: %w", err)
	}

	key, err := url2.PathUnescape(strings.TrimPrefix(parsed.Path, "/"))
	if err != nil {
		return fmt.Errorf("failed to unescape file path: %w", err)
	}

	_, err = u.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(u.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
