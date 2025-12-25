package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/coocood/freecache"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // <-- Add this line to register the Postgres driver
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/auth"
	authstorage "github.com/Haerd-Limited/dating-api/internal/auth/storage"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/communication"
	"github.com/Haerd-Limited/dating-api/internal/config"
	"github.com/Haerd-Limited/dating-api/internal/conversation"
	"github.com/Haerd-Limited/dating-api/internal/conversation/score"
	conversationstorage "github.com/Haerd-Limited/dating-api/internal/conversation/storage"
	"github.com/Haerd-Limited/dating-api/internal/discover"
	discoverstorage "github.com/Haerd-Limited/dating-api/internal/discover/storage"
	"github.com/Haerd-Limited/dating-api/internal/feedback"
	feedbackstorage "github.com/Haerd-Limited/dating-api/internal/feedback/storage"
	"github.com/Haerd-Limited/dating-api/internal/http/router"
	insightsvc "github.com/Haerd-Limited/dating-api/internal/insights"
	insightstorage "github.com/Haerd-Limited/dating-api/internal/insights/storage"
	"github.com/Haerd-Limited/dating-api/internal/interaction"
	interactionstorage "github.com/Haerd-Limited/dating-api/internal/interaction/storage"
	"github.com/Haerd-Limited/dating-api/internal/lookup"
	lookupstorage "github.com/Haerd-Limited/dating-api/internal/lookup/storage"
	"github.com/Haerd-Limited/dating-api/internal/matching"
	matchingstorage "github.com/Haerd-Limited/dating-api/internal/matching/storage"
	"github.com/Haerd-Limited/dating-api/internal/media"
	"github.com/Haerd-Limited/dating-api/internal/notification"
	notificationstorage "github.com/Haerd-Limited/dating-api/internal/notification/storage"
	"github.com/Haerd-Limited/dating-api/internal/onboarding"
	"github.com/Haerd-Limited/dating-api/internal/openai"
	"github.com/Haerd-Limited/dating-api/internal/preference"
	preferencestorage "github.com/Haerd-Limited/dating-api/internal/preference/storage"
	"github.com/Haerd-Limited/dating-api/internal/profile"
	profilestorage "github.com/Haerd-Limited/dating-api/internal/profile/storage"
	"github.com/Haerd-Limited/dating-api/internal/realtime"
	"github.com/Haerd-Limited/dating-api/internal/safety"
	safetystorage "github.com/Haerd-Limited/dating-api/internal/safety/storage"
	"github.com/Haerd-Limited/dating-api/internal/uow"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	"github.com/Haerd-Limited/dating-api/internal/verification"
	verificationstorage "github.com/Haerd-Limited/dating-api/internal/verification/storage"
	commonanalytics "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/analytics"
	commondb "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/db"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/ids"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/render"
	s3Storage "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Shutdown on Ctrl+C
	go listenForShutdown(cancel)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	logger := commonlogger.New(cfg)

	// Connect to the PostgreSQL database using sqlx.
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Sugar().Fatalf("failed to connect to database: %v", err)
	}

	err = commondb.RunMigrations(db.DB, nil)
	if err != nil {
		logger.Sugar().Fatalf("failed to run migrations: %v", err)
	}

	// Init analytics emitter (non-blocking event sink)
	commonanalytics.InitDefault(db, logger, 2048)
	defer commonanalytics.ShutdownDefault()

	// In bytes, where 1024 * 1024 represents a single Megabyte, and 100 * 1024*1024 represents 100 Megabytes.
	cacheSize := 100 * 1024 * 1024 // todo: use for intentions and stuff. no need to read db since never change
	cache := freecache.NewCache(cacheSize)

	debug.SetGCPercent(20)

	s3Presigner, err := s3Storage.NewPresigner(ctx, cfg.Env, cfg.AWSRegion, cfg.S3BucketName)
	if err != nil {
		logger.Sugar().Fatalf("failed to create S3 presigner: %v", err)
	}

	s3Uploader, err := s3Storage.NewS3Uploader(cfg.S3BucketName, cfg.AWSRegion)
	if err != nil {
		logger.Sugar().Fatalf("failed to create S3 uploader: %v", err)
	}

	s3Reader, err := s3Storage.NewS3Reader(ctx, logger, cfg.AWSRegion, cfg.S3BucketName)
	if err != nil {
		logger.Sugar().Fatalf("failed to create S3 reader: %v", err)
	}

	verificationRepo := verificationstorage.NewVerificationRepository(db)
	lookupRepo := lookupstorage.NewLookupRepository(db)
	profileRepo := profilestorage.NewProfileRepository(db)
	preferenceRepo := preferencestorage.NewPreferenceRepository(db)
	discoverRepo := discoverstorage.NewDiscoverRepository(db)
	conversationRepo := conversationstorage.NewConversationRepository(db)
	interactionRepo := interactionstorage.NewInteractionRepository(db)
	safetyRepo := safetystorage.NewRepository(db)
	feedbackRepo := feedbackstorage.NewRepository(db)
	userRepo := storage.NewUserRepository(db)
	authRepo := authstorage.NewAuthRepository(db)
	matchingRepo := matchingstorage.NewMatchingRepository(db, logger)
	deviceTokenRepo := notificationstorage.NewDeviceTokenRepository(db)
	unitOfWork := uow.New(db.DB)

	hub := realtime.NewHub()
	flake := ids.NewSnowflake(1)

	rek, err := aws.NewRek(ctx, cfg.AWSRekognitionRegion)
	if err != nil {
		logger.Sugar().Fatalf("failed to create rek: %v", err)
	}

	matchingService := matching.NewMatchingService(logger, matchingRepo)
	awsService := aws.NewAwsService(logger, s3Uploader, s3Presigner, s3Reader, cfg.Env)
	openaiService := openai.NewOpenAIService(cfg.OpenAIAPIKey, logger)
	lookupService := lookup.NewLookupService(logger, lookupRepo)
	profileService := profile.NewProfileService(logger, profileRepo, lookupRepo, verificationRepo, openaiService, awsService, unitOfWork)
	preferenceService := preference.NewPreferenceService(logger, preferenceRepo)
	discoverService := discover.NewDiscoverService(logger, profileService, matchingService, discoverRepo)
	scoreService := score.NewScoreService(logger, conversationRepo, unitOfWork)

	notificationService, err := notification.NewService(ctx, logger, deviceTokenRepo, notification.Config{
		ServiceAccountPath: cfg.FirebaseServiceAccountPath,
		CredentialsJSON:    cfg.GoogleCredentialsJson,
		ProjectID:          cfg.FirebaseProjectID,
	})
	if err != nil {
		logger.Sugar().Fatalf("failed to initialise notification service: %v", err)
	}

	notificationService.StartWeeklyRefreshScheduler(ctx)
	verificationService := verification.NewVerificationService(rek.Client, cfg.AWSRekognitionRegion, verificationRepo, awsService, profileService, logger, hub, notificationService)
	conversationService := conversation.NewConversationService(logger, conversationRepo, profileService, flake, hub, interactionRepo, scoreService, unitOfWork, notificationService)
	safetyService := safety.NewService(logger, safetyRepo, conversationRepo, unitOfWork, hub)
	interactionService := interaction.NewInteractionService(logger, profileService, conversationService, interactionRepo, discoverService, safetyService, unitOfWork, hub, notificationService)
	userService := user.NewUserService(logger, userRepo, awsService, cache, unitOfWork, profileService, preferenceService)
	communicationService := communication.NewService(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioNumber)
	authService := auth.NewAuthService(logger, cfg.JwtSecret, userService, authRepo, awsService, communicationService, cfg.Env)
	mediaService := media.NewMediaService(logger, awsService, openaiService)
	notificationPhoneNumbers := parsePhoneNumbers(cfg.NotificationPhoneNumbers)
	backendEngineerPhoneNumbers := parsePhoneNumbers(cfg.BackendEngineerPhoneNumbers)
	frontendEngineerPhoneNumbers := parsePhoneNumbers(cfg.FrontendEngineerPhoneNumbers)
	feedbackService := feedback.NewService(logger, feedbackRepo, unitOfWork, communicationService, notificationPhoneNumbers)

	// Initialize error notification handler for SMS alerts
	render.InitErrorNotificationHandler(
		communicationService,
		backendEngineerPhoneNumbers,
		frontendEngineerPhoneNumbers,
		logger,
	)

	onboardingService := onboarding.NewOnboardingService(
		logger,
		userService,
		authService,
		mediaService,
		profileService,
		lookupRepo,
		communicationService,
		verificationService,
		notificationPhoneNumbers,
		cfg.EnablePreregCap,
		cfg.MaxParticipants,
		cfg.MaxMaleParticipants,
		cfg.MaxFemaleParticipants,
	)

	// Insights
	insRepo := insightstorage.NewRepository(db)
	insSvc := insightsvc.NewService(logger, insRepo)

	// Start background weekly scheduler for insights (runs inside API process).
	runInsightsWeekly(ctx, logger, insSvc, insRepo)

	mux := router.New(
		logger,
		cfg.JwtSecret,
		authService,
		onboardingService,
		profileService,
		discoverService,
		interactionService,
		conversationService,
		mediaService,
		lookupService,
		hub,
		verificationService,
		matchingService,
		notificationService,
		safetyService,
		insSvc,
		userService,
		feedbackService,
		cfg.AdminAPIKey,
	)

	// Start server with context
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	go func() {
		logger.Sugar().Infof("Server starting on port %s", cfg.Port)

		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Sugar().Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait for cancel (SIGINT / SIGTERM)
	<-ctx.Done()

	// Graceful shutdown
	logger.Sugar().Info("Shutting down gracefully...")

	if err := server.Shutdown(context.Background()); err != nil {
		logger.Sugar().Fatalf("Server forced to shutdown: %v", err)
	}
}

func listenForShutdown(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	fmt.Println("Shutdown signal received")
	cancel()
}

// runInsightsWeekly starts a goroutine that runs every day at 02:00 UTC.
// On Mondays it computes last week's [Mon..Mon) window and stores a snapshot.
func runInsightsWeekly(ctx context.Context, logger *zap.Logger, ins insightsvc.Service, repo insightstorage.Repository) {
	go func() {
		for {
			now := time.Now().UTC()
			// Next trigger at 02:00 UTC.
			next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, time.UTC)
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(next)):
			}

			// Only run the weekly snapshot on Mondays (Monday == 0 with this mapping)
			today := time.Now().UTC()

			weekday := (int(today.Weekday()) + 6) % 7
			if weekday != 0 {
				continue
			}

			weekEnd := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC) // this Monday start
			weekStart := weekEnd.AddDate(0, 0, -7)

			global, err := ins.GetPublicWeekly(ctx, weekStart)
			if err != nil {
				logger.Sugar().Warnw("insights weekly compute failed", "error", err)
				continue
			}

			if err := repo.InsertGlobalWeeklySnapshot(ctx, "weekly_highlights", weekStart, weekEnd, global); err != nil {
				logger.Sugar().Warnw("insights weekly snapshot insert failed", "error", err)
			} else {
				logger.Sugar().Infof("Insights weekly snapshot stored for %s", weekStart.Format("2006-01-02"))
			}
		}
	}()
}

// parsePhoneNumbers parses a comma-separated string of phone numbers and returns a slice of trimmed, non-empty numbers.
func parsePhoneNumbers(phoneNumbersStr string) []string {
	if phoneNumbersStr == "" {
		return nil
	}

	numbers := strings.Split(phoneNumbersStr, ",")
	result := make([]string, 0, len(numbers))

	for _, phone := range numbers {
		phone = strings.TrimSpace(phone)
		if phone != "" {
			result = append(result, phone)
		}
	}

	return result
}
