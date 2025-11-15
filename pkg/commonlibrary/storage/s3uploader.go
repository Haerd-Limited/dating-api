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
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

// DeleteAllFilesByPrefix deletes all files in S3 that match the given prefix.
// This is useful for deleting all files belonging to a user.
func (u *S3Uploader) DeleteAllFilesByPrefix(ctx context.Context, prefix string) error {
	// List all objects with the given prefix
	listInput := &s3.ListObjectsV2Input{
		Bucket: aws.String(u.BucketName),
		Prefix: aws.String(prefix),
	}

	var objectKeys []string

	// Paginate through all objects
	paginator := s3.NewListObjectsV2Paginator(u.Client, listInput)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
		}

		for _, obj := range page.Contents {
			objectKeys = append(objectKeys, *obj.Key)
		}
	}

	if len(objectKeys) == 0 {
		// No files to delete
		return nil
	}

	// Delete objects in batches (S3 allows up to 1000 objects per delete request)
	const batchSize = 1000
	for i := 0; i < len(objectKeys); i += batchSize {
		end := i + batchSize
		if end > len(objectKeys) {
			end = len(objectKeys)
		}

		batch := objectKeys[i:end]
		deleteObjects := make([]types.ObjectIdentifier, len(batch))

		for j, key := range batch {
			deleteObjects[j] = types.ObjectIdentifier{
				Key: aws.String(key),
			}
		}

		_, err := u.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(u.BucketName),
			Delete: &types.Delete{
				Objects: deleteObjects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete batch of objects with prefix %s: %w", prefix, err)
		}
	}

	return nil
}
