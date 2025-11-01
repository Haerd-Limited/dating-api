package openai

import (
	"bytes"
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

//go:generate mockgen -source=service.go -destination=service_mock.go -package=openai
type Service interface {
	TranscribeAudio(ctx context.Context, audioData []byte, filename string) (string, error)
}

type service struct {
	client *openai.Client
	logger *zap.Logger
}

func NewOpenAIService(apiKey string, logger *zap.Logger) Service {
	return &service{
		client: openai.NewClient(apiKey),
		logger: logger,
	}
}

func (s *service) TranscribeAudio(ctx context.Context, audioData []byte, filename string) (string, error) {
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		Reader:   bytes.NewReader(audioData),
		FilePath: filename,
	}

	resp, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("openai transcription failed: %w", err)
	}

	return resp.Text, nil
}
