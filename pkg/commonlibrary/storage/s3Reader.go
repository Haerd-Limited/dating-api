package storage

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type S3Reader struct {
	logger *zap.Logger
	Client *s3.Client
	Bucket string
}

func NewS3Reader(ctx context.Context, logger *zap.Logger, region, bucketName string) (*S3Reader, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &S3Reader{
		Client: s3.NewFromConfig(cfg),
		Bucket: bucketName,
		logger: logger,
	}, nil
}

func (r *S3Reader) GetObjectBytes(ctx context.Context, key string) ([]byte, error) {
	out, err := r.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			r.logger.Sugar().Errorw("failed to close S3 reader", "error", err)
		}
	}(out.Body)

	return io.ReadAll(out.Body)
}

func (r *S3Reader) PutObject(ctx context.Context, key string, body []byte) error {
	_, err := r.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(key),
		Body: io.NopCloser(io.NewSectionReader(
			// cheap reader over bytes
			// but simpler: bytes.NewReader(body)
			// leaving this for clarity:
			bytes.NewReader(body), 0, int64(len(body)),
		)),
	})

	return err
}
