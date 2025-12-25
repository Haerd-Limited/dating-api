package media

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/media/domain"
	"github.com/Haerd-Limited/dating-api/internal/openai"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	GeneratePhotoUploadUrl(ctx context.Context, userID string) (domain.UploadUrl, error)
	GenerateVoiceNoteUploadUrl(ctx context.Context, userID string, purpose string) (domain.UploadUrl, error)
	GenerateUploadURLsForProfilePhotos(ctx context.Context, userID string) ([]domain.UploadUrl, error)
	GenerateUploadURLsForProfilePrompts(ctx context.Context, userID string) ([]domain.UploadUrl, error)
	GenerateFeedbackAttachmentUploadUrl(ctx context.Context, userID string, mediaType string) (domain.UploadUrl, error)
	TranscribeInstagramReel(ctx context.Context, reelURL string) (string, error)
}

const (
	maxUploadCountPhotos     = 6
	minUploadCountPhotos     = 1
	maxUploadCountPrompts    = 6
	minUploadCountVoiceNotes = 1
	maxUploadBytes           = 5 << 20 // 5 MiB
	presignTTL               = 20 * time.Minute
	mimeJPEG                 = "image/jpeg"
	mimePNG                  = "image/png"
	mimeM4A                  = "audio/mp4" // m4a is an MP4 container; "audio/m4a" also seen but "audio/mp4" is safer
	mimeMP4                  = "video/mp4"
)

type service struct {
	logger        *zap.Logger
	awsService    aws.Service
	openaiService openai.Service
}

func NewMediaService(
	logger *zap.Logger,
	awsService aws.Service,
	openaiService openai.Service,
) Service {
	return &service{
		logger:        logger,
		awsService:    awsService,
		openaiService: openaiService,
	}
}

func (s *service) GeneratePhotoUploadUrl(ctx context.Context, userID string) (domain.UploadUrl, error) {
	url, err := s.awsService.GenerateUploadURLs(ctx, userID, minUploadCountPhotos, mimeJPEG, presignTTL, nil)
	if err != nil {
		return domain.UploadUrl{}, commonlogger.LogError(s.logger, "failed to generate photo upload url", err, zap.String("userID", userID))
	}

	if len(url) != minUploadCountPhotos {
		return domain.UploadUrl{}, fmt.Errorf("failed to generate photo upload url: expected %d urls, got %d", minUploadCountPhotos, len(url))
	}

	return domain.UploadUrl{
		Key:       url[0].Key,
		UploadUrl: url[0].URL,
		Headers:   url[0].Headers,
		MaxBytes:  maxUploadBytes,
	}, nil
}

func (s *service) GenerateVoiceNoteUploadUrl(ctx context.Context, userID string, purpose string) (domain.UploadUrl, error) {
	url, err := s.awsService.GenerateUploadURLs(ctx, userID, minUploadCountVoiceNotes, mimeM4A, presignTTL, &purpose)
	if err != nil {
		return domain.UploadUrl{}, commonlogger.LogError(s.logger, "failed to generate voicenote upload url", err, zap.String("userID", userID), zap.String("purpose", purpose))
	}

	if len(url) != minUploadCountVoiceNotes {
		return domain.UploadUrl{}, fmt.Errorf("failed to generate voicenote upload url: expected %d urls, got %d", minUploadCountVoiceNotes, len(url))
	}

	return domain.UploadUrl{
		Key:       url[0].Key,
		UploadUrl: url[0].URL,
		Headers:   url[0].Headers,
		MaxBytes:  maxUploadBytes,
	}, nil
}

func (s *service) GenerateUploadURLsForProfilePhotos(ctx context.Context, userID string) ([]domain.UploadUrl, error) {
	urls, err := s.awsService.GenerateUploadURLs(ctx, userID, maxUploadCountPhotos, mimeJPEG, presignTTL, nil)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "failed to generate upload urls", err, zap.String("userID", userID))
	}

	var photoUploadUrls []domain.UploadUrl
	for _, url := range urls {
		photoUploadUrls = append(photoUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  maxUploadBytes,
		})
	}

	return photoUploadUrls, nil
}

func (s *service) GenerateUploadURLsForProfilePrompts(ctx context.Context, userID string) ([]domain.UploadUrl, error) {
	urls, err := s.awsService.GenerateUploadURLs(ctx, userID, maxUploadCountPrompts, mimeM4A, presignTTL, nil)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "failed to generate upload urls", err, zap.String("userID", userID))
	}

	var voicePromptUploadUrls []domain.UploadUrl
	for _, url := range urls {
		voicePromptUploadUrls = append(voicePromptUploadUrls, domain.UploadUrl{
			Key:       url.Key,
			UploadUrl: url.URL,
			Headers:   url.Headers,
			MaxBytes:  maxUploadBytes,
		})
	}

	return voicePromptUploadUrls, nil
}

func (s *service) GenerateFeedbackAttachmentUploadUrl(ctx context.Context, userID string, mediaType string) (domain.UploadUrl, error) {
	var contentType string

	purpose := "feedback"

	switch mediaType {
	case "image":
		contentType = mimeJPEG
	case "video":
		contentType = mimeMP4
	default:
		return domain.UploadUrl{}, fmt.Errorf("invalid media type: %s, must be 'image' or 'video'", mediaType)
	}

	urls, err := s.awsService.GenerateUploadURLs(ctx, userID, 1, contentType, presignTTL, &purpose)
	if err != nil {
		return domain.UploadUrl{}, commonlogger.LogError(s.logger, "failed to generate feedback attachment upload url", err, zap.String("userID", userID), zap.String("mediaType", mediaType))
	}

	if len(urls) != 1 {
		return domain.UploadUrl{}, fmt.Errorf("failed to generate feedback attachment upload url: expected 1 url, got %d", len(urls))
	}

	return domain.UploadUrl{
		Key:       urls[0].Key,
		UploadUrl: urls[0].URL,
		Headers:   urls[0].Headers,
		MaxBytes:  maxUploadBytes,
	}, nil
}

func (s *service) TranscribeInstagramReel(ctx context.Context, reelURL string) (string, error) {
	// Validate Instagram reel URL
	if !isValidInstagramReelURL(reelURL) {
		return "", fmt.Errorf("invalid Instagram reel URL: %s", reelURL)
	}

	// Create temporary directory for video and audio files
	tempDir, err := os.MkdirTemp("", "reel-transcribe-*")
	if err != nil {
		return "", commonlogger.LogError(s.logger, "create temp directory", err, zap.String("reelURL", reelURL))
	}
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			s.logger.Warn("failed to cleanup temp directory",
				zap.String("tempDir", tempDir),
				zap.Error(cleanupErr))
		}
	}()

	videoPath := filepath.Join(tempDir, "video.mp4")
	audioPath := filepath.Join(tempDir, "audio.mp3")

	// Download video using yt-dlp
	if err := s.downloadVideo(ctx, reelURL, videoPath); err != nil {
		return "", commonlogger.LogError(s.logger, "download video", err, zap.String("reelURL", reelURL))
	}

	// Extract audio using FFmpeg
	if err := s.extractAudio(ctx, videoPath, audioPath); err != nil {
		return "", commonlogger.LogError(s.logger, "extract audio", err, zap.String("videoPath", videoPath))
	}

	// Read audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return "", commonlogger.LogError(s.logger, "read audio file", err, zap.String("audioPath", audioPath))
	}

	// Transcribe using OpenAI
	transcript, err := s.openaiService.TranscribeAudio(ctx, audioData, "audio.mp3")
	if err != nil {
		return "", commonlogger.LogError(s.logger, "transcribe audio", err, zap.String("reelURL", reelURL))
	}

	s.logger.Info("successfully transcribed Instagram reel",
		zap.String("reelURL", reelURL),
		zap.Int("transcriptLength", len(transcript)))

	return transcript, nil
}

func isValidInstagramReelURL(url string) bool {
	// Instagram reel URL patterns:
	// https://www.instagram.com/reel/...
	// https://instagram.com/reel/...
	// https://www.instagram.com/p/... (posts can also be reels)
	pattern := `^https?://(www\.)?instagram\.com/(reel|p)/[A-Za-z0-9_-]+`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

func (s *service) downloadVideo(ctx context.Context, reelURL, outputPath string) error {
	// Use yt-dlp to download the video
	// yt-dlp -f "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best" -o <output_path> <url>
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		"-o", outputPath,
		reelURL,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %w, output: %s", err, string(output))
	}

	// Verify the file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("downloaded video file not found at %s", outputPath)
	}

	s.logger.Debug("video downloaded successfully",
		zap.String("reelURL", reelURL),
		zap.String("outputPath", outputPath))

	return nil
}

func (s *service) extractAudio(ctx context.Context, videoPath, audioPath string) error {
	// Use FFmpeg to extract audio
	// ffmpeg -i <video_path> -vn -acodec libmp3lame -ab 192k <audio_path>
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", videoPath,
		"-vn",
		"-acodec", "libmp3lame",
		"-ab", "192k",
		"-y", // Overwrite output file if it exists
		audioPath,
	)

	// FFmpeg outputs to stderr, so we capture that
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	// Verify the file was created
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return fmt.Errorf("extracted audio file not found at %s", audioPath)
	}

	s.logger.Debug("audio extracted successfully",
		zap.String("videoPath", videoPath),
		zap.String("audioPath", audioPath))

	return nil
}
