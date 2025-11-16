package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/config"
	insightsvc "github.com/Haerd-Limited/dating-api/internal/insights"
	insightstorage "github.com/Haerd-Limited/dating-api/internal/insights/storage"
	commonlogger "github.com/Haerd-Limited/dating-api/pkg/commonlibrary/logger"
)

func main() {
	fromFlag := flag.String("from", "", "period start (YYYY-MM-DD, week)")
	toFlag := flag.String("to", "", "period end (YYYY-MM-DD, week)")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := commonlogger.New(cfg)

	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Sugar().Fatalf("connect db: %v", err)
	}

	repo := insightstorage.NewRepository(db)
	svc := insightsvc.NewService(logger, repo)

	var from time.Time

	var to time.Time

	if *fromFlag != "" {
		from, err = time.Parse("2006-01-02", *fromFlag)
		if err != nil {
			logger.Sugar().Fatalf("invalid from: %v", err)
		}
	} else {
		// default: last week Monday
		now := time.Now().UTC()
		offset := (int(now.Weekday()) + 6) % 7
		thisMonday := time.Date(now.Year(), now.Month(), now.Day()-offset, 0, 0, 0, 0, time.UTC)
		from = thisMonday.AddDate(0, 0, -7)
	}

	if *toFlag != "" {
		to, err = time.Parse("2006-01-02", *toFlag)
		if err != nil {
			logger.Sugar().Fatalf("invalid to: %v", err)
		}
	} else {
		to = from.AddDate(0, 0, 7)
	}

	runWeekly(ctx, logger, svc, repo, from, to)
}

func runWeekly(ctx context.Context, logger *zap.Logger, svc insightsvc.Service, repo insightstorage.Repository, from, to time.Time) {
	global, err := svc.GetPublicWeekly(ctx, from)
	if err != nil {
		logger.Sugar().Fatalf("compute weekly: %v", err)
	}

	if err := repo.InsertGlobalWeeklySnapshot(ctx, "weekly_highlights", from, to, global); err != nil {
		logger.Sugar().Fatalf("insert snapshot: %v", err)
	}

	fmt.Println("weekly insights snapshot stored", from.Format("2006-01-02"))
}
