package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
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
	TwilioAccountSID           string `mapstructure:"TWILIO_ACCOUNT_SID" yaml:"twilio_account_sid"`
	TwilioAuthToken            string `mapstructure:"TWILIO_AUTH_TOKEN" yaml:"twilio_auth_token"`
	TwilioNumber               string `mapstructure:"TWILIO_NUMBER" yaml:"twilio_number"`
}

// LoadConfig loads configuration from the OS environment unless you're in local. In that case from a .env file at the root of the repository.
func LoadConfig() (*Config, error) {
	// Use Viper to read environment variables.
	viper.AutomaticEnv()

	if viper.GetString("ENV") == "" {
		log.Printf("environment variable %s not set. Defaulting to local", viper.GetString("ENV"))
		if err := godotenv.Load(); err != nil {
			log.Fatalf("no .env file found. please place a .env file at the root of this repository: %v", err)
		}
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
		TwilioAccountSID:           viper.GetString("TWILIO_ACCOUNT_SID"),
		TwilioAuthToken:            viper.GetString("TWILIO_AUTH_TOKEN"),
		TwilioNumber:               viper.GetString("TWILIO_NUMBER"),
	}

	// Validate the config.
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
