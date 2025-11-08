package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/Haerd-Limited/dating-api/internal/config"
)

func New(cfg *config.Config) *zap.Logger {
	var (
		logger *zap.Logger
		err    error
	)

	switch strings.ToLower(cfg.Env) {
	case "local", "development", "dev", "test":
		devConfig := zap.NewDevelopmentConfig()
		devConfig.OutputPaths = []string{"stdout"}
		devConfig.ErrorOutputPaths = []string{"stdout"}

		logger, err = devConfig.Build()
	default:
		logger, err = zap.NewProduction()
	}

	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	return logger
}
