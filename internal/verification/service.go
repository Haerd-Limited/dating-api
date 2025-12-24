// internal/verification/service.go
package verification

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aws/aws-sdk-go-v2/aws"
	rek "github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/google/uuid"
	"go.uber.org/zap"

	realtimedto "github.com/Haerd-Limited/dating-api/internal/api/realtime/dto"
	internalaws "github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/entity"
	"github.com/Haerd-Limited/dating-api/internal/notification"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	profiledomain "github.com/Haerd-Limited/dating-api/internal/profile/domain"
	"github.com/Haerd-Limited/dating-api/internal/realtime"
	"github.com/Haerd-Limited/dating-api/internal/verification/domain"
	"github.com/Haerd-Limited/dating-api/internal/verification/storage"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

type Service interface {
	StartPhotoVerification(ctx context.Context, userID string) (domain.StartResult, error)
	CompletePhotoVerification(ctx context.Context, userID, sessionID string) (domain.CompleteResult, error)
	StartVideoVerification(ctx context.Context, userID string) (domain.StartVideoResult, error)
	SubmitVideoVerification(ctx context.Context, userID, videoS3Key string) (domain.SubmitVideoResult, error)
	// Admin video verification methods
	ListVideoAttempts(ctx context.Context, filter domain.VideoAttemptFilter) ([]domain.VideoAttempt, error)
	GetVideoAttempt(ctx context.Context, attemptID string) (*domain.VideoAttempt, error)
	GetVideoDownloadURL(ctx context.Context, videoS3Key string) (string, error)
	ApproveVideoAttempt(ctx context.Context, attemptID string, notes *string) error
	RejectVideoAttempt(ctx context.Context, attemptID string, rejectionReason string, notes *string) error
}

type service struct {
	logger              *zap.Logger
	rek                 internalaws.RekClient
	region              string
	verificationRepo    storage.VerificationRepository
	awsService          internalaws.Service
	profileService      profile.Service
	notificationService notification.Service
	hub                 realtime.Broadcaster
}

func NewVerificationService(
	rek internalaws.RekClient,
	region string,
	verificationRepo storage.VerificationRepository,
	awsService internalaws.Service,
	profileService profile.Service,
	logger *zap.Logger,
	hub realtime.Broadcaster,
	notificationService notification.Service,
) Service {
	return &service{
		rek:                 rek,
		region:              region,
		verificationRepo:    verificationRepo,
		awsService:          awsService,
		profileService:      profileService,
		logger:              logger,
		hub:                 hub,
		notificationService: notificationService,
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

		return domain.StartResult{}, commonlogger.LogError(s.logger, "create face liveness session", err, zap.String("userID", userID), zap.String("clientToken", tok))
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

		return domain.StartResult{}, commonlogger.LogError(s.logger, "create attempt", err, zap.String("userID", userID), zap.String("sessionID", sessionID), zap.String("clientToken", tok))
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
		return domain.CompleteResult{}, commonlogger.LogError(s.logger, "get face liveness result", err, zap.String("userID", userID), zap.String("sessionID", sessionID))
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
		return domain.CompleteResult{}, commonlogger.LogError(s.logger, "get verification attempt", err, zap.String("userID", userID), zap.String("sessionID", sessionID))
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
			return domain.CompleteResult{}, commonlogger.LogError(s.logger, "marshal reason codes", mErr, zap.String("userID", userID), zap.String("sessionID", sessionID))
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

			return domain.CompleteResult{}, commonlogger.LogError(s.logger, "mark attempt", err, zap.String("userID", userID), zap.String("sessionID", sessionID), zap.String("attemptID", verificationAttempt.ID))
		}

		reason := "liveness_failed"
		s.sendVerificationStatusEvent(userID, entity.VerificationStatusFailed, verificationAttempt.ID, &reason)

		return domain.CompleteResult{Status: entity.VerificationStatusFailed, PhotoVerified: false, Reasons: []string{"liveness_failed"}}, nil
	}

	// Best frame bytes for matching
	var bestFrame []byte
	if res.ReferenceImage != nil && len(res.ReferenceImage.Bytes) > 0 {
		bestFrame = res.ReferenceImage.Bytes
	} else {
		return domain.CompleteResult{}, errors.New("no reference image from liveness result")
	}
	// todo: On haerd-dating/verification/, set an S3 lifecycle rule to expire after 7–30 days. (Ops: S3 → Bucket → Management → Lifecycle rules.)
	bestKey, bErr := s.awsService.StoreBestFrame(ctx, bestFrame, userID)
	if bErr == nil {
		s.logger.Debug("best frame stored",
			zap.String("user_id", userID), zap.String("session_id", sessionID),
			zap.String("best_frame_key", bestKey),
		)

		verificationAttempt.BestFrameS3Key = null.StringFrom(bestKey)
		// don’t fail the whole flow if this write fails; it’s only for review/audit
	}

	// 2) Compare against user's stored private photos
	keys, err := s.verificationRepo.GetUserPrivatePhotoKeys(ctx, userID)
	if err != nil || len(keys) == 0 {
		return domain.CompleteResult{}, commonlogger.LogError(s.logger, "no user photos to compare", err, zap.String("userID", userID), zap.String("sessionID", sessionID))
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
			return domain.CompleteResult{}, commonlogger.LogError(s.logger, "mark attempt", err, zap.String("userID", userID), zap.String("sessionID", sessionID), zap.String("attemptID", verificationAttempt.ID))
		}

		err = s.verificationRepo.SetUserPhotoVerified(ctx, userID, verificationAttempt.ID)
		if err != nil {
			return domain.CompleteResult{}, commonlogger.LogError(s.logger, "set user photo verified", err, zap.String("userID", userID), zap.String("sessionID", sessionID), zap.String("attemptID", verificationAttempt.ID))
		}

		err = s.profileService.VerifyProfile(ctx, userID)
		if err != nil {
			return domain.CompleteResult{}, commonlogger.LogError(s.logger, "verify profile", err, zap.String("userID", userID), zap.String("sessionID", sessionID))
		}

		s.logger.Info("verification passed",
			zap.String("attempt_id", verificationAttempt.ID),
			zap.String("user_id", userID),
			zap.Float32("liveness", *res.Confidence),
			zap.Float64("best_similarity", bestSim),
		)

		s.sendVerificationStatusEvent(userID, entity.VerificationStatusPassed, verificationAttempt.ID, nil)

		return domain.CompleteResult{Status: entity.VerificationStatusPassed, MatchScore: bestSim, PhotoVerified: true}, nil
	}

	// todo(high-priority): create a dashboard for manual review.
	verificationAttempt.Status = entity.VerificationStatusNeedsReview
	confidence := float64(*res.Confidence)
	verificationAttempt.LivenessScore = null.Float64From(confidence)

	reasonCodes, mErr := json.Marshal([]string{"low_similarity"})
	if mErr != nil {
		return domain.CompleteResult{}, commonlogger.LogError(s.logger, "marshal reason codes", mErr, zap.String("userID", userID), zap.String("sessionID", sessionID))
	}

	verificationAttempt.ReasonCodes = null.JSONFrom(reasonCodes)
	verificationAttempt.MatchScore = null.Float64From(bestSim)

	err = s.verificationRepo.MarkAttempt(ctx, *verificationAttempt)
	if err != nil {
		return domain.CompleteResult{}, commonlogger.LogError(s.logger, "mark attempt", err, zap.String("userID", userID), zap.String("sessionID", sessionID), zap.String("attemptID", verificationAttempt.ID))
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

func (s *service) StartVideoVerification(ctx context.Context, userID string) (domain.StartVideoResult, error) {
	// Check for any ongoing video verification (pending or needs_review)
	ongoing, err := s.verificationRepo.CheckIfOngoingVideoAttemptExists(ctx, userID)
	if err == nil && ongoing != nil {
		// If there's an ongoing verification, check if it's pending (can reuse) or needs_review (must block)
		if ongoing.Status == entity.VerificationStatusNeedsReview {
			s.logger.Info("blocked video verification start - already under review",
				zap.String("user_id", userID),
				zap.String("attempt_id", ongoing.ID),
				zap.String("status", ongoing.Status),
			)

			return domain.StartVideoResult{}, ErrOngoingVideoVerification
		}

		// Status is pending - reuse existing attempt
		code, codeErr := s.verificationRepo.GetVerificationCode(ctx, ongoing.ID)
		if codeErr != nil || code == "" {
			// If code is missing, generate new one and update
			newCode, genErr := generateFourDigitCode()
			if genErr != nil {
				return domain.StartVideoResult{}, commonlogger.LogError(s.logger, "generate code", genErr, zap.String("userID", userID))
			}

			code = newCode
			// Update the attempt with the new code
			updateErr := s.verificationRepo.UpdateVerificationCode(ctx, ongoing.ID, code)
			if updateErr != nil {
				return domain.StartVideoResult{}, commonlogger.LogError(s.logger, "update verification code", updateErr, zap.String("userID", userID))
			}
		}

		s.logger.Info("start reused pending video attempt",
			zap.String("user_id", userID),
			zap.String("attempt_id", ongoing.ID),
		)

		// Generate presigned URL for video upload
		purpose := "verification-video"

		urls, urlErr := s.awsService.GenerateUploadURLs(ctx, userID, 1, "video/mp4", 20*time.Minute, &purpose)
		if urlErr != nil {
			return domain.StartVideoResult{}, commonlogger.LogError(s.logger, "generate upload url", urlErr, zap.String("userID", userID))
		}

		if len(urls) == 0 {
			return domain.StartVideoResult{}, errors.New("failed to generate upload url")
		}

		return domain.StartVideoResult{
			Code:      code,
			UploadURL: urls[0].URL,
			UploadKey: urls[0].Key,
		}, nil
	}

	// Generate random 4-digit code (1000-9999)
	code, err := generateFourDigitCode()
	if err != nil {
		return domain.StartVideoResult{}, commonlogger.LogError(s.logger, "generate code", err, zap.String("userID", userID))
	}

	// Generate presigned URL for video upload
	purpose := "verification-video"

	urls, err := s.awsService.GenerateUploadURLs(ctx, userID, 1, "video/mp4", 20*time.Minute, &purpose)
	if err != nil {
		return domain.StartVideoResult{}, commonlogger.LogError(s.logger, "generate upload url", err, zap.String("userID", userID))
	}

	if len(urls) == 0 {
		return domain.StartVideoResult{}, errors.New("failed to generate upload url")
	}

	// Create verification attempt
	// Note: After entity regeneration, VerificationCode field will be available
	// For now, create attempt and then update with code via raw SQL
	attempt := entity.VerificationAttempt{
		UserID:           userID,
		Type:             "video",
		Status:           entity.VerificationStatusPending,
		VerificationCode: null.StringFrom(code),
	}

	err = s.verificationRepo.CreateAttempt(ctx, attempt)
	if err != nil {
		s.logger.Error("persist video attempt failed",
			zap.String("user_id", userID),
			zap.String("code", code),
			zap.Error(err),
		)

		return domain.StartVideoResult{}, commonlogger.LogError(s.logger, "create video attempt", err, zap.String("userID", userID))
	}

	s.logger.Info("start created video attempt",
		zap.String("user_id", userID),
		zap.String("code", code),
	)

	return domain.StartVideoResult{
		Code:      code,
		UploadURL: urls[0].URL,
		UploadKey: urls[0].Key,
	}, nil
}

func (s *service) SubmitVideoVerification(ctx context.Context, userID, videoS3Key string) (domain.SubmitVideoResult, error) {
	// Update attempt with video S3 key and mark as needs_review
	err := s.verificationRepo.UpdateVideoAttemptWithKey(ctx, userID, videoS3Key)
	if err != nil {
		s.logger.Error("submit video verification failed",
			zap.String("user_id", userID),
			zap.String("video_s3_key", videoS3Key),
			zap.Error(err),
		)

		return domain.SubmitVideoResult{}, commonlogger.LogError(s.logger, "update video attempt", err, zap.String("userID", userID), zap.String("videoS3Key", videoS3Key))
	}

	// Set profile to UNDER_REVIEW status
	err = s.profileService.SetProfileUnderReview(ctx, userID)
	if err != nil {
		s.logger.Warn("failed to set profile under review after video submission",
			zap.String("userID", userID),
			zap.String("videoS3Key", videoS3Key),
			zap.Error(err))
		// Don't fail the whole operation if this fails
	}

	s.logger.Info("video verification submitted",
		zap.String("user_id", userID),
		zap.String("video_s3_key", videoS3Key),
	)

	return domain.SubmitVideoResult{
		Status: "submitted",
	}, nil
}

// generateFourDigitCode generates a random 4-digit code (1000-9999)
func generateFourDigitCode() (string, error) {
	const digits = "0123456789"

	b := make([]byte, 4)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random code: %w", err)
	}

	// Ensure first digit is 1-9 (so code is between 1000-9999)
	b[0] = digits[1+int(b[0])%9] // 1-9
	for i := 1; i < 4; i++ {
		b[i] = digits[int(b[i])%10] // 0-9
	}

	return string(b), nil
}

var (
	ErrVideoAttemptNotFound      = errors.New("video attempt not found")
	ErrInvalidVideoAttemptStatus = errors.New("invalid video attempt status")
	ErrRejectionReasonRequired   = errors.New("rejection reason is required")
	ErrOngoingVideoVerification  = errors.New("you already have a video verification in progress or under review")
)

func (s *service) ListVideoAttempts(ctx context.Context, filter domain.VideoAttemptFilter) ([]domain.VideoAttempt, error) {
	attempts, err := s.verificationRepo.ListVideoAttempts(ctx, filter)
	if err != nil {
		return nil, commonlogger.LogError(s.logger, "list video attempts", err)
	}

	result := make([]domain.VideoAttempt, 0, len(attempts))

	for _, attempt := range attempts {
		videoAttempt := s.entityToVideoAttempt(attempt)

		// Fetch user's profile photos for comparison
		photos, err := s.profileService.GetUserPhotos(ctx, attempt.UserID)
		if err != nil {
			s.logger.Warn("failed to get user photos for video attempt",
				zap.String("userID", attempt.UserID),
				zap.String("attemptID", attempt.ID),
				zap.Error(err))
			// Don't fail the whole operation if photos can't be fetched
			videoAttempt.Photos = []profiledomain.Photo{}
		} else {
			videoAttempt.Photos = photos
		}

		result = append(result, videoAttempt)
	}

	return result, nil
}

func (s *service) GetVideoAttempt(ctx context.Context, attemptID string) (*domain.VideoAttempt, error) {
	attempt, err := s.verificationRepo.GetVideoAttemptByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVideoAttemptNotFound
		}

		return nil, commonlogger.LogError(s.logger, "get video attempt", err, zap.String("attemptID", attemptID))
	}

	if attempt.Type != "video" {
		return nil, ErrVideoAttemptNotFound
	}

	videoAttempt := s.entityToVideoAttempt(attempt)

	// Fetch user's profile photos for comparison
	photos, err := s.profileService.GetUserPhotos(ctx, attempt.UserID)
	if err != nil {
		s.logger.Warn("failed to get user photos for video attempt",
			zap.String("userID", attempt.UserID),
			zap.String("attemptID", attemptID),
			zap.Error(err))
		// Don't fail the whole operation if photos can't be fetched
		videoAttempt.Photos = []profiledomain.Photo{}
	} else {
		videoAttempt.Photos = photos
	}

	return &videoAttempt, nil
}

func (s *service) GetVideoDownloadURL(ctx context.Context, videoS3Key string) (string, error) {
	if videoS3Key == "" {
		return "", errors.New("video S3 key is required")
	}

	// Generate presigned URL with 1 hour expiration
	url, err := s.awsService.GenerateDownloadURL(ctx, videoS3Key, 1*time.Hour)
	if err != nil {
		return "", commonlogger.LogError(s.logger, "generate download URL", err, zap.String("videoS3Key", videoS3Key))
	}

	return url, nil
}

func (s *service) ApproveVideoAttempt(ctx context.Context, attemptID string, notes *string) error {
	attempt, err := s.verificationRepo.GetVideoAttemptByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrVideoAttemptNotFound
		}

		return commonlogger.LogError(s.logger, "get video attempt", err, zap.String("attemptID", attemptID))
	}

	if attempt.Type != "video" {
		return ErrVideoAttemptNotFound
	}

	// Validate status - can only approve pending or needs_review attempts
	if attempt.Status != entity.VerificationStatusPending && attempt.Status != entity.VerificationStatusNeedsReview {
		return fmt.Errorf("%w: cannot approve attempt with status %s", ErrInvalidVideoAttemptStatus, attempt.Status)
	}

	// Update status to passed
	err = s.verificationRepo.UpdateVideoAttemptStatus(ctx, attemptID, entity.VerificationStatusPassed, nil)
	if err != nil {
		return commonlogger.LogError(s.logger, "update video attempt status", err, zap.String("attemptID", attemptID))
	}

	// Update user verification status (similar to photo verification)
	err = s.verificationRepo.SetUserPhotoVerified(ctx, attempt.UserID, attemptID)
	if err != nil {
		s.logger.Warn("failed to set user photo verified after video approval",
			zap.String("userID", attempt.UserID),
			zap.String("attemptID", attemptID),
			zap.Error(err))
		// Don't fail the whole operation if this fails
	}

	// Verify profile
	err = s.profileService.VerifyProfile(ctx, attempt.UserID)
	if err != nil {
		s.logger.Warn("failed to verify profile after video approval",
			zap.String("userID", attempt.UserID),
			zap.String("attemptID", attemptID),
			zap.Error(err))
		// Don't fail the whole operation if this fails
	}

	// Send notification to user
	if s.notificationService != nil {
		if err := s.notificationService.SendVerificationApprovedNotification(ctx, attempt.UserID); err != nil {
			s.logger.Warn("failed to send verification approved notification",
				zap.String("userID", attempt.UserID),
				zap.String("attemptID", attemptID),
				zap.Error(err))
			// Don't fail the whole operation if this fails
		}
	}

	s.logger.Info("video verification approved",
		zap.String("attempt_id", attemptID),
		zap.String("user_id", attempt.UserID),
		zap.String("notes", func() string {
			if notes != nil {
				return *notes
			}

			return ""
		}()))

	s.sendVerificationStatusEvent(attempt.UserID, entity.VerificationStatusPassed, attemptID, nil)

	return nil
}

func (s *service) RejectVideoAttempt(ctx context.Context, attemptID string, rejectionReason string, notes *string) error {
	if rejectionReason == "" {
		return ErrRejectionReasonRequired
	}

	attempt, err := s.verificationRepo.GetVideoAttemptByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrVideoAttemptNotFound
		}

		return commonlogger.LogError(s.logger, "get video attempt", err, zap.String("attemptID", attemptID))
	}

	if attempt.Type != "video" {
		return ErrVideoAttemptNotFound
	}

	// Validate status - can only reject pending or needs_review attempts
	if attempt.Status != entity.VerificationStatusPending && attempt.Status != entity.VerificationStatusNeedsReview {
		return fmt.Errorf("%w: cannot reject attempt with status %s", ErrInvalidVideoAttemptStatus, attempt.Status)
	}

	// Update status to failed with rejection reason
	err = s.verificationRepo.UpdateVideoAttemptStatus(ctx, attemptID, entity.VerificationStatusFailed, &rejectionReason)
	if err != nil {
		return commonlogger.LogError(s.logger, "update video attempt status", err, zap.String("attemptID", attemptID))
	}

	// Set profile status back to unverified
	err = s.profileService.SetProfileUnverified(ctx, attempt.UserID)
	if err != nil {
		s.logger.Warn("failed to set profile unverified after video rejection",
			zap.String("userID", attempt.UserID),
			zap.String("attemptID", attemptID),
			zap.Error(err))
		// Don't fail the whole operation if this fails
	}

	// Send notification to user
	if s.notificationService != nil {
		if err := s.notificationService.SendVerificationRejectedNotification(ctx, attempt.UserID, rejectionReason); err != nil {
			s.logger.Warn("failed to send verification rejected notification",
				zap.String("userID", attempt.UserID),
				zap.String("attemptID", attemptID),
				zap.Error(err))
			// Don't fail the whole operation if this fails
		}
	}

	s.logger.Info("video verification rejected",
		zap.String("attempt_id", attemptID),
		zap.String("user_id", attempt.UserID),
		zap.String("rejection_reason", rejectionReason),
		zap.String("notes", func() string {
			if notes != nil {
				return *notes
			}

			return ""
		}()))

	s.sendVerificationStatusEvent(attempt.UserID, entity.VerificationStatusFailed, attemptID, &rejectionReason)

	return nil
}

// entityToVideoAttempt converts an entity.VerificationAttempt to domain.VideoAttempt
func (s *service) entityToVideoAttempt(attempt *entity.VerificationAttempt) domain.VideoAttempt {
	var verificationCode string
	if attempt.VerificationCode.Valid {
		verificationCode = attempt.VerificationCode.String
	}

	var videoS3Key string
	if attempt.VideoS3Key.Valid {
		videoS3Key = attempt.VideoS3Key.String
	}

	var rejectionReason *string

	if attempt.ReasonCodes.Valid && len(attempt.ReasonCodes.JSON) > 0 {
		var reasonCodes []string
		if err := json.Unmarshal(attempt.ReasonCodes.JSON, &reasonCodes); err == nil && len(reasonCodes) > 0 {
			rejectionReason = &reasonCodes[0]
		}
	}

	return domain.VideoAttempt{
		ID:               attempt.ID,
		UserID:           attempt.UserID,
		VerificationCode: verificationCode,
		VideoS3Key:       videoS3Key,
		Status:           attempt.Status,
		RejectionReason:  rejectionReason,
		CreatedAt:        attempt.CreatedAt,
		UpdatedAt:        attempt.UpdatedAt,
	}
}

// sendVerificationStatusEvent sends a WebSocket event when verification status changes
func (s *service) sendVerificationStatusEvent(userID string, status string, attemptID string, reason *string) {
	eventData := map[string]interface{}{
		"status":     status,
		"attempt_id": attemptID,
	}
	if reason != nil {
		eventData["reason"] = *reason
	}

	evt := realtimedto.Event{
		ID:        realtime.NewEventID(),
		Type:      "verification.status_changed",
		ActorID:   userID,
		Ts:        time.Now(),
		ContextID: attemptID,
		Data:      eventData,
		Version:   1,
	}

	b, err := json.Marshal(evt)
	if err != nil {
		s.logger.Error("error marshalling verification status event", zap.Error(err))
		return
	}

	s.hub.BroadcastToUser(userID, b)
}
