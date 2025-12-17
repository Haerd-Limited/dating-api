package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type UploadSlot struct {
	Key         string            `json:"key"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"` // headers the client must send (e.g., Content-Type)
	ContentType string            `json:"contentType"`
}

type presigner struct {
	bucket    string
	region    string
	presigner *s3.PresignClient
	env       string
}

type Presigner interface {
	GenerateUploadURLs(
		ctx context.Context,
		userID string, count int,
		contentType string,
		ttl time.Duration,
		purpose *string) ([]UploadSlot, error)
}

func NewPresigner(ctx context.Context, env, region, bucket string, loadOpts ...func(*config.LoadOptions) error) (Presigner, error) {
	if region == "" || bucket == "" {
		return nil, errors.New("region and bucket are required")
	}

	cfg, err := config.LoadDefaultConfig(ctx, append(loadOpts, config.WithRegion(region))...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &presigner{
		bucket:    bucket,
		region:    region,
		presigner: s3.NewPresignClient(client),
		env:       env,
	}, nil
}

// GenerateUploadURLs returns 'count' presigned PUT URLs under users/{userID}/photos
// All URLs are signed to require the provided contentType header.
func (p *presigner) GenerateUploadURLs(ctx context.Context, userID string, count int, contentType string, ttl time.Duration, purpose *string) ([]UploadSlot, error) {
	if count <= 0 {
		return nil, errors.New("count must be > 0")
	}

	if contentType == "" {
		return nil, errors.New("contentType is required (e.g., image/jpeg)")
	}

	ext := extForContentType(contentType)
	if ext == "" {
		return nil, fmt.Errorf("unsupported contentType: %s", contentType)
	}

	slots := make([]UploadSlot, 0, count)
	KeyBase := fmt.Sprintf("%s/users/%s", p.env, sanitize(userID))

	for i := 0; i < count; i++ {
		var key string

		if purpose != nil && *purpose == "voicenote" {
			key = fmt.Sprintf("%s/messages/voice-notes/%s.%s", KeyBase, uuid.NewString(), ext)
		} else if purpose != nil && *purpose == "feedback" {
			if strings.Contains(contentType, "video") {
				key = fmt.Sprintf("%s/feedback-attachments/videos/%s.%s", KeyBase, uuid.NewString(), ext)
			} else {
				key = fmt.Sprintf("%s/feedback-attachments/images/%s.%s", KeyBase, uuid.NewString(), ext)
			}
		} else if purpose != nil && *purpose == "verification-video" {
			key = fmt.Sprintf("%s/verification/videos/%s.%s", KeyBase, uuid.NewString(), ext)
		} else if strings.Contains(contentType, "audio") {
			key = fmt.Sprintf("%s/prompts/%s.%s", KeyBase, uuid.NewString(), ext)
		} else {
			key = fmt.Sprintf("%s/profile-photos/%s.%s", KeyBase, uuid.NewString(), ext)
		}

		req := &s3.PutObjectInput{
			Bucket:      aws.String(p.bucket),
			Key:         aws.String(key),
			ContentType: aws.String(contentType),
			// No ACL here; keep bucket public access blocked.
			// Do not set SSE headers if bucket has default encryption.
			// Optional: Metadata: map[string]string{"user-id": userID},
		}

		out, err := p.presigner.PresignPutObject(ctx, req, func(po *s3.PresignOptions) {
			po.Expires = ttl
		})
		if err != nil {
			return nil, fmt.Errorf("presign put for key %s: %w", key, err)
		}

		slots = append(slots, UploadSlot{
			Key:         key,
			URL:         out.URL,
			Headers:     map[string]string{"Content-Type": contentType},
			ContentType: contentType,
		})
	}

	return slots, nil
}

func extForContentType(ct string) string {
	switch strings.ToLower(ct) {
	// photo
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/heic", "image/heif":
		return "heic"
		// audio
	case "audio/mp4", "audio/m4a", "audio/x-m4a":
		return "m4a"
	case "audio/aac":
		return "aac"
	case "audio/mpeg", "audio/mp3":
		return "mp3"
	case "audio/wav", "audio/x-wav":
		return "wav"
	case "audio/webm":
		return "webm"
	case "audio/ogg", "audio/opus":
		return "ogg"
		// video
	case "video/mp4":
		return "mp4"
	case "video/quicktime", "video/x-quicktime":
		return "mov"
	case "video/webm":
		return "webm"
	case "video/x-msvideo", "video/avi":
		return "avi"
	}

	return ""
}

func sanitize(s string) string {
	// minimal sanitization to avoid path traversal or odd characters in keys
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_' || r == '.':
			return r
		default:
			return '-'
		}
	}, s)
}
