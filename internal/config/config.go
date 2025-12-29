package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port                         string `mapstructure:"PORT" yaml:"port" validate:"required"`
	DatabaseURL                  string `mapstructure:"DATABASE_URL" yaml:"database_url" validate:"required"`
	LogLevel                     string `mapstructure:"LOG_LEVEL" yaml:"log_level"`
	Env                          string `mapstructure:"ENV" yaml:"env" validate:"required"`
	JwtSecret                    string `mapstructure:"JWT_SECRET" yaml:"jwt_secret" validate:"required"`
	AdminAPIKey                  string `mapstructure:"ADMIN_API_KEY" yaml:"admin_api_key"`
	AWSRegion                    string `mapstructure:"AWS_REGION" yaml:"aws_region" validate:"required"`
	AWSRekognitionRegion         string `mapstructure:"AWS_REKOGNITION_REGION" yaml:"aws_rekognition_region" validate:"required"`
	AWSAccessKeyID               string `mapstructure:"AWS_ACCESS_KEY_ID" yaml:"aws_access_key_id" validate:"required"`
	AWSSecretAccessKey           string `mapstructure:"AWS_SECRET_ACCESS_KEY" yaml:"aws_secret_access_key" validate:"required"`
	S3BucketName                 string `mapstructure:"S3_BUCKET_NAME" yaml:"s3_bucket_name" validate:"required"`
	ExpoAccessToken              string `mapstructure:"EXPO_ACCESS_TOKEN"`
	TwilioAccountSID             string `mapstructure:"TWILIO_ACCOUNT_SID" yaml:"twilio_account_sid"`
	TwilioAuthToken              string `mapstructure:"TWILIO_AUTH_TOKEN" yaml:"twilio_auth_token"`
	TwilioNumber                 string `mapstructure:"TWILIO_NUMBER" yaml:"twilio_number"`
	NotificationPhoneNumbers     string `mapstructure:"NOTIFICATION_PHONE_NUMBERS" yaml:"notification_phone_numbers"`           // Comma-separated list
	BackendEngineerPhoneNumbers  string `mapstructure:"BACKEND_ENGINEER_PHONE_NUMBERS" yaml:"backend_engineer_phone_numbers"`   // Comma-separated list
	FrontendEngineerPhoneNumbers string `mapstructure:"FRONTEND_ENGINEER_PHONE_NUMBERS" yaml:"frontend_engineer_phone_numbers"` // Comma-separated list
	OpenAIAPIKey                 string `mapstructure:"OPENAI_API_KEY" yaml:"openai_api_key" validate:"required"`
	// Feature flags and limits for preregistration caps
	EnablePreregCap       bool `mapstructure:"ENABLE_PREREG_CAP" yaml:"enable_prereg_cap"`
	MaxParticipants       int  `mapstructure:"MAX_PARTICIPANTS" yaml:"max_participants"`
	MaxMaleParticipants   int  `mapstructure:"MAX_MALE_PARTICIPANTS" yaml:"max_male_participants"`
	MaxFemaleParticipants int  `mapstructure:"MAX_FEMALE_PARTICIPANTS" yaml:"max_female_participants"`
}

// LoadConfig loads from OS env; if ENV=local (or unset) it will attempt to load .env first.
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	// Sensible default; can be overridden by real ENV
	viper.SetDefault("ENV", "local")
	// Defaults for prereg limits
	viper.SetDefault("ENABLE_PREREG_CAP", true)
	viper.SetDefault("MAX_PARTICIPANTS", 1500)
	viper.SetDefault("MAX_MALE_PARTICIPANTS", 750)
	viper.SetDefault("MAX_FEMALE_PARTICIPANTS", 750)

	// If ENV explicitly set to "local" (or not set in OS), try .env without failing hard.
	rawEnv := os.Getenv("ENV")
	if rawEnv == "" || strings.EqualFold(rawEnv, "local") {
		if err := godotenv.Load(); err != nil {
			// Not fatal—just informational in local dev
			log.Printf("no .env file found (ok if running in CI/containers): %v", err)
		}
	}

	// Create a Config instance with values from environment variables.
	cfg := Config{
		Port:                         viper.GetString("PORT"),
		DatabaseURL:                  viper.GetString("DATABASE_URL"),
		LogLevel:                     viper.GetString("LOG_LEVEL"),
		JwtSecret:                    viper.GetString("JWT_SECRET"),
		AdminAPIKey:                  viper.GetString("ADMIN_API_KEY"),
		Env:                          viper.GetString("ENV"),
		AWSRegion:                    viper.GetString("AWS_REGION"),
		AWSAccessKeyID:               viper.GetString("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey:           viper.GetString("AWS_SECRET_ACCESS_KEY"),
		S3BucketName:                 viper.GetString("S3_BUCKET_NAME"),
		ExpoAccessToken:              viper.GetString("EXPO_ACCESS_TOKEN"),
		TwilioAccountSID:             viper.GetString("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:              viper.GetString("TWILIO_AUTH_TOKEN"),
		TwilioNumber:                 viper.GetString("TWILIO_NUMBER"),
		NotificationPhoneNumbers:     viper.GetString("NOTIFICATION_PHONE_NUMBERS"),
		BackendEngineerPhoneNumbers:  viper.GetString("BACKEND_ENGINEER_PHONE_NUMBERS"),
		FrontendEngineerPhoneNumbers: viper.GetString("FRONTEND_ENGINEER_PHONE_NUMBERS"),
		AWSRekognitionRegion:         viper.GetString("AWS_REKOGNITION_REGION"),
		OpenAIAPIKey:                 viper.GetString("OPENAI_API_KEY"),
		EnablePreregCap:              viper.GetBool("ENABLE_PREREG_CAP"),
		MaxParticipants:              viper.GetInt("MAX_PARTICIPANTS"),
		MaxMaleParticipants:          viper.GetInt("MAX_MALE_PARTICIPANTS"),
		MaxFemaleParticipants:        viper.GetInt("MAX_FEMALE_PARTICIPANTS"),
	}

	// Validate the config.
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
