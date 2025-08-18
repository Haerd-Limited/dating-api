package config

import (
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port                       string `mapstructure:"PORT" yaml:"port" validate:"required"`
	DatabaseURL                string `mapstructure:"DATABASE_URL" yaml:"database_url" validate:"required"`
	LogLevel                   string `mapstructure:"LOG_LEVEL" yaml:"log_level"`
	Env                        string `mapstructure:"ENV" yaml:"env" validate:"required"`
	JwtSecret                  string `mapstructure:"JWT_SECRET" yaml:"jwt_secret" validate:"required"`
	AWSRegion                  string `mapstructure:"AWS_REGION" yaml:"aws_region" validate:"required"`
	AWSAccessKeyID             string `mapstructure:"AWS_ACCESS_KEY_ID" yaml:"aws_access_key_id" validate:"required"`
	AWSSecretAccessKey         string `mapstructure:"AWS_SECRET_ACCESS_KEY" yaml:"aws_secret_access_key" validate:"required"`
	S3BucketName               string `mapstructure:"S3_BUCKET_NAME" yaml:"s3_bucket_name" validate:"required"`
	FirebaseServiceAccountPath string `mapstructure:"FIREBASE_SERVICE_ACCOUNT_PATH"`
	GoogleCredentialsJson      string `mapstructure:"GOOGLE_CREDENTIALS_JSON"`
}

// LoadConfig loads configuration from the OS environment and, if not in production,
// from a .env file at the root of the repository.
func LoadConfig() (*Config, error) {
	// Check if running in production.
	// When ENV is "prod", we assume all necessary environment variables are set.
	// Otherwise, load variables from the .env file.
	if os.Getenv("ENV") != "prod" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: no .env file found, relying on OS environment variables: %v", err)
		}
	}

	// Use Viper to read environment variables.
	viper.AutomaticEnv()

	// Set a default value for ENV if it hasn't been set.
	if viper.GetString("ENV") == "" {
		viper.Set("ENV", "dev")
	}

	// Create a Config instance with values from environment variables.
	cfg := Config{
		Port:                       viper.GetString("PORT"),
		DatabaseURL:                viper.GetString("DATABASE_URL"),
		LogLevel:                   viper.GetString("LOG_LEVEL"),
		JwtSecret:                  viper.GetString("JWT_SECRET"),
		Env:                        viper.GetString("ENV"),
		AWSRegion:                  viper.GetString("AWS_REGION"),
		AWSAccessKeyID:             viper.GetString("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:         viper.GetString("AWS_SECRET_ACCESS_KEY"),
		S3BucketName:               viper.GetString("S3_BUCKET_NAME"),
		FirebaseServiceAccountPath: viper.GetString("FIREBASE_SERVICE_ACCOUNT_PATH"),
		GoogleCredentialsJson:      viper.GetString("GOOGLE_CREDENTIALS_JSON"),
	}

	// Validate the config.
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
