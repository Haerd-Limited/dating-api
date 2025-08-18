package aws

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/aws/domain"
	"github.com/Haerd-Limited/dating-api/internal/aws/mapper"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/audio"
	commonStorage "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/storage"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=aws
type AWSService interface {
	UploadVoiceNote(ctx context.Context, input domain.VoiceNoteUpload) (*string, error)
	UploadImage(ctx context.Context, input domain.ImageUpload) (*string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}

type awsService struct {
	logger     *zap.Logger
	S3Uploader *commonStorage.S3Uploader
}

func NewAwsService(
	logger *zap.Logger,
	s3Uploader *commonStorage.S3Uploader,
) AWSService {
	return &awsService{
		logger:     logger,
		S3Uploader: s3Uploader,
	}
}

func (s *awsService) UploadVoiceNote(ctx context.Context, input domain.VoiceNoteUpload) (*string, error) {
	contentType, err := audio.DetectContentType(input.VoiceNoteFile)
	if err != nil {
		return nil, fmt.Errorf("failed to detect voicenote content type: %v", err)
	}

	folderPath, err := mapper.MapStringToS3AudioFolderPath(string(input.FolderPath))
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("%s/%s_%v_%s", folderPath, input.AuthorID, time.Now().UnixNano(), input.VoiceNoteHeader.Filename)

	fileURL, err := s.S3Uploader.UploadFile(ctx, input.VoiceNoteFile, fileName, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload audio: %v", err)
	}

	return &fileURL, nil
}

func (s *awsService) UploadImage(ctx context.Context, input domain.ImageUpload) (*string, error) {
	// upload post image to s3
	imageContentType, err := audio.DetectContentType(input.ImageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to detect image content type: %v", err)
	}

	folderPath, err := mapper.MapStringToS3ImageFolderPath(string(input.FolderPath))
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("%s/%s_%v_%s", folderPath, input.AuthorID, time.Now().Unix(), input.ImageHeader.Filename)

	fileURL, err := s.S3Uploader.UploadFile(ctx, input.ImageFile, fileName, imageContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %v", err)
	}

	return &fileURL, nil
}

func (s *awsService) DeleteFile(ctx context.Context, fileURL string) error {
	return s.S3Uploader.DeleteFile(ctx, fileURL)
}
