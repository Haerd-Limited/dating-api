// internal/verification/service.go
package verification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	internalaws "github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	"github.com/Haerd-Limited/dating-api/internal/verification/domain"
	"github.com/Haerd-Limited/dating-api/internal/verification/storage"
	"github.com/aarondl/null/v8"
	"github.com/aws/aws-sdk-go-v2/aws"
	rek "github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	StartPhotoVerification(ctx context.Context, userID string) (domain.StartResult, error)
	CompletePhotoVerification(ctx context.Context, userID, sessionID string) (domain.CompleteResult, error)
}

type service struct {
	logger           *zap.Logger
	rek              internalaws.RekClient
	region           string
	verificationRepo storage.VerificationRepository
	awsService       internalaws.Service
	profileService   profile.Service
}

func NewVerificationService(
	rek internalaws.RekClient,
	region string,
	verificationRepo storage.VerificationRepository,
	awsService internalaws.Service,
	profileService profile.Service,
	logger *zap.Logger,
) Service {
	return &service{
		rek:              rek,
		region:           region,
		verificationRepo: verificationRepo,
		awsService:       awsService,
		profileService:   profileService,
		logger:           logger,
	}
}

func (s *service) StartPhotoVerification(ctx context.Context, userID string) (domain.StartResult, error) {
	existing, _ := s.verificationRepo.CheckIfPendingAttemptsExist(ctx, userID)
	if existing != nil {
		s.logger.Info("start reused pending attempt",
			zap.String("user_id", userID),
			zap.String("session_id", existing.SessionID.String),
		)
		return domain.StartResult{SessionID: existing.SessionID.String, Region: s.region}, nil
	}

	tok := uuid.NewString()
	out, err := s.rek.CreateFaceLivenessSession(ctx, &rek.CreateFaceLivenessSessionInput{
		ClientRequestToken: aws.String(tok),
	})
	if err != nil {
		s.logger.Error("create liveness session failed",
			zap.String("user_id", userID),
			zap.String("client_token", tok),
			zap.Error(err),
		)
		return domain.StartResult{}, fmt.Errorf("create face liveness session: %w", err)
	}
	sessionID := aws.ToString(out.SessionId)

	// persist attempt
	err = s.verificationRepo.CreateAttempt(ctx, entity.VerificationAttempt{
		UserID:      userID,
		Type:        entity.VerificationTypePhoto,
		Status:      entity.VerificationStatusPending,
		SessionID:   null.StringFrom(sessionID),
		ClientToken: null.StringFrom(tok),
	})
	if err != nil {
		s.logger.Error("persist attempt failed",
			zap.String("user_id", userID),
			zap.String("session_id", sessionID),
			zap.String("client_token", tok),
			zap.Error(err),
		)
		return domain.StartResult{}, fmt.Errorf("create attempt: %w", err)
	}

	s.logger.Info("start created attempt",
		zap.String("user_id", userID),
		zap.String("session_id", sessionID),
		zap.String("client_token", tok),
		zap.String("region", s.region),
	)

	return domain.StartResult{
		SessionID: sessionID,
		Region:    s.region,
	}, nil
}

func (s *service) CompletePhotoVerification(ctx context.Context, userID, sessionID string) (domain.CompleteResult, error) {
	// 1) Fetch liveness result
	res, err := s.rek.GetFaceLivenessSessionResults(ctx, &rek.GetFaceLivenessSessionResultsInput{
		SessionId: aws.String(sessionID),
	})
	if err != nil {
		s.logger.Error("get liveness results failed",
			zap.String("user_id", userID), zap.String("session_id", sessionID), zap.Error(err))
		return domain.CompleteResult{}, fmt.Errorf("get face liveness result: %w", err)
	}
	if res.Confidence == nil {
		s.logger.Warn("liveness missing confidence",
			zap.String("user_id", userID), zap.String("session_id", sessionID))
		return domain.CompleteResult{}, errors.New("no confidence from liveness result")
	}

	verificationAttempt, err := s.verificationRepo.GetVerificationAttemptByUserIDAndSessionID(ctx, userID, sessionID)
	if err != nil {
		s.logger.Error("load attempt failed",
			zap.String("user_id", userID), zap.String("session_id", sessionID), zap.Error(err))
		return domain.CompleteResult{}, fmt.Errorf("get verification attempt: %w", err)
	}

	if *res.Confidence < domain.LivenessPassThreshold {
		confidence := float64(*res.Confidence)
		s.logger.Info("liveness failed",
			zap.String("user_id", userID), zap.String("session_id", sessionID),
			zap.Float64("liveness", confidence),
			zap.Float64("threshold", domain.LivenessPassThreshold),
		)
		verificationAttempt.Status = entity.VerificationStatusFailed
		verificationAttempt.LivenessScore = null.Float64From(confidence)
		reasonCodes, mErr := json.Marshal([]string{"liveness_low_confidence"})
		if mErr != nil {
			return domain.CompleteResult{}, fmt.Errorf("marshal reason codes: %w", mErr)
		}
		verificationAttempt.ReasonCodes = null.JSONFrom(reasonCodes)

		err = s.verificationRepo.MarkAttempt(ctx, *verificationAttempt)
		if err != nil {
			s.logger.Info("attempt marked failed",
				zap.String("attempt_id", verificationAttempt.ID),
				zap.String("user_id", userID),
				zap.Float64("liveness", confidence),
				zap.Strings("reasons", []string{"liveness_low_confidence"}),
			)
			return domain.CompleteResult{}, fmt.Errorf("mark attempt: %w", err)
		}
		return domain.CompleteResult{Status: entity.VerificationStatusFailed, PhotoVerified: false, Reasons: []string{"liveness_failed"}}, nil
	}

	// Best frame bytes for matching
	var bestFrame []byte
	if res.ReferenceImage != nil && len(res.ReferenceImage.Bytes) > 0 {
		bestFrame = res.ReferenceImage.Bytes
	} else {
		return domain.CompleteResult{}, errors.New("no reference image from liveness result")
	}
	//todo: On haerd-dating/verification/, set an S3 lifecycle rule to expire after 7–30 days. (Ops: S3 → Bucket → Management → Lifecycle rules.)
	bestKey, bErr := s.awsService.StoreBestFrame(ctx, bestFrame, userID)
	if bErr == nil {
		s.logger.Debug("best frame stored",
			zap.String("user_id", userID), zap.String("session_id", sessionID),
			zap.String("best_frame_key", bestKey),
		)
		verificationAttempt.BestFrameS3Key = null.StringFrom(bestKey)
		// don’t fail the whole flow if this write fails; it’s only for review/audit
	}

	// 2) Compare against user’s stored private photos
	keys, err := s.verificationRepo.GetUserPrivatePhotoKeys(ctx, userID)
	if err != nil || len(keys) == 0 {
		return domain.CompleteResult{}, fmt.Errorf("no user photos to compare: %w", err)
	}

	var bestSim float64
	for _, key := range keys {
		targetImgBytes, gErr := s.awsService.GetObjectBytes(ctx, key)
		if gErr != nil {
			continue
		}

		out, cErr := s.rek.CompareFaces(ctx, &rek.CompareFacesInput{
			SourceImage:         &types.Image{Bytes: bestFrame},
			TargetImage:         &types.Image{Bytes: targetImgBytes},
			SimilarityThreshold: aws.Float32(float32(domain.FaceMatchThreshold)),
		})
		if cErr != nil {
			continue
		}
		// Rekognition returns zero or more face matches
		for _, m := range out.FaceMatches {
			if m.Similarity != nil && float64(*m.Similarity) > bestSim {
				bestSim = float64(*m.Similarity)
			}
		}
	}
	s.logger.Info("compare summary",
		zap.String("user_id", userID), zap.String("session_id", sessionID),
		zap.Float32("liveness", *res.Confidence),
		zap.Float64("best_similarity", bestSim),
		zap.Float64("threshold", domain.FaceMatchThreshold),
		zap.Int("photos_compared", len(keys)),
	)

	if bestSim >= domain.FaceMatchThreshold {
		verificationAttempt.Status = entity.VerificationStatusPassed
		confidence := float64(*res.Confidence)
		verificationAttempt.LivenessScore = null.Float64From(confidence)
		verificationAttempt.MatchScore = null.Float64From(bestSim)

		err = s.verificationRepo.MarkAttempt(ctx, *verificationAttempt)
		if err != nil {
			return domain.CompleteResult{}, fmt.Errorf("mark attempt: %w", err)
		}

		err = s.verificationRepo.SetUserPhotoVerified(ctx, userID, verificationAttempt.ID)
		if err != nil {
			return domain.CompleteResult{}, fmt.Errorf("set user photo verified: %w", err)
		}
		err = s.profileService.VerifyProfile(ctx, userID)
		if err != nil {
			return domain.CompleteResult{}, fmt.Errorf("verify profile: %w", err)
		}
		s.logger.Info("verification passed",
			zap.String("attempt_id", verificationAttempt.ID),
			zap.String("user_id", userID),
			zap.Float32("liveness", *res.Confidence),
			zap.Float64("best_similarity", bestSim),
		)
		return domain.CompleteResult{Status: entity.VerificationStatusPassed, MatchScore: bestSim, PhotoVerified: true}, nil
	}

	//todo: create a dashboard for manual review.
	verificationAttempt.Status = entity.VerificationStatusNeedsReview
	confidence := float64(*res.Confidence)
	verificationAttempt.LivenessScore = null.Float64From(confidence)
	reasonCodes, mErr := json.Marshal([]string{"low_similarity"})
	if mErr != nil {
		return domain.CompleteResult{}, fmt.Errorf("marshal reason codes: %w", mErr)
	}
	verificationAttempt.ReasonCodes = null.JSONFrom(reasonCodes)
	verificationAttempt.MatchScore = null.Float64From(bestSim)
	err = s.verificationRepo.MarkAttempt(ctx, *verificationAttempt)
	if err != nil {
		return domain.CompleteResult{}, fmt.Errorf("mark attempt: %w", err)
	}
	s.logger.Info("verification needs review",
		zap.String("attempt_id", verificationAttempt.ID),
		zap.String("user_id", userID),
		zap.Float32("liveness", *res.Confidence),
		zap.Float64("best_similarity", bestSim),
		zap.Strings("reasons", []string{"low_similarity"}),
	)

	return domain.CompleteResult{
		Status:        entity.VerificationStatusNeedsReview,
		MatchScore:    bestSim,
		PhotoVerified: false,
		Reasons:       []string{"face_not_close_enough"},
	}, nil
}
