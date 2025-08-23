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
	"syscall"

	"github.com/coocood/freecache"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // <-- Add this line to register the Postgres driver

	"github.com/Haerd-Limited/dating-api/internal/auth"
	authstorage "github.com/Haerd-Limited/dating-api/internal/auth/storage"
	"github.com/Haerd-Limited/dating-api/internal/aws"
	"github.com/Haerd-Limited/dating-api/internal/config"
	"github.com/Haerd-Limited/dating-api/internal/http/router"
	"github.com/Haerd-Limited/dating-api/internal/onboarding"
	onboardingstorage "github.com/Haerd-Limited/dating-api/internal/onboarding/storage"
	"github.com/Haerd-Limited/dating-api/internal/user"
	"github.com/Haerd-Limited/dating-api/internal/user/storage"
	commondb "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/db"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
	s3Storage "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/storage"
)

func main() {
	// Load environment variables from .env file (only in development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env file")
	}

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

	if err := commondb.RunMigrations(db.DB, nil); err != nil {
		logger.Sugar().Fatalf("failed to run migrations: %v", err)
	}

	// In bytes, where 1024 * 1024 represents a single Megabyte, and 100 * 1024*1024 represents 100 Megabytes.
	cacheSize := 100 * 1024 * 1024
	cache := freecache.NewCache(cacheSize)

	debug.SetGCPercent(20)

	// notificationRepo := notificationStorage.NewNotificationRepository(db)

	s3Uploader, err := s3Storage.NewS3Uploader(cfg.S3BucketName, cfg.AWSRegion)
	if err != nil {
		logger.Sugar().Fatalf("failed to create S3 uploader: %v", err)
	}

	awsService := aws.NewAwsService(logger, s3Uploader)

	userRepo := storage.NewUserRepository(db)
	userService := user.NewUserService(logger, userRepo, awsService, cache)

	authRepo := authstorage.NewAuthRepository(db)
	authService := auth.NewAuthService(logger, cfg.JwtSecret, userService, authRepo, awsService)

	onboardingRepo := onboardingstorage.NewOnboardingRepository(db)
	onboardingService := onboarding.NewOnboardingService(logger, onboardingRepo, userService, authService)

	/*notificationService, err := notification.NewNotificationService(logger, notificationRepo, cfg.GoogleCredentialsJson)
	if err != nil {
		logger.Sugar().Fatalf("failed to create notification service: %v", err)
	}*/

	mux := router.New(
		logger,
		cfg.JwtSecret,
		authService,
		userService,
		onboardingService,
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
