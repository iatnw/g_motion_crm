package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	AWS      AWSConfig
	Auth     AuthConfig
	Twilio   TwilioConfig
	Tally    TallyConfig
}

type AppConfig struct {
	Env           string
	Port          string
	BaseURL       string
	FrontendURL   string
	EncryptionKey string
}

type DatabaseConfig struct {
	DatabaseURL string
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketDocs      string
	BucketBackups   string
}

type AuthConfig struct {
	JWTSecret  string
	TOTPIssuer string
}

type TwilioConfig struct{}

type TallyConfig struct {
	SigningSecret string
}

func Load() (*Config, error) {
	if os.Getenv("APP_ENV") == "" {
		err := godotenv.Load()
		if err != nil {
			return nil, err
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		return nil, fmt.Errorf("APP_PORT is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is required")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKeyID == "" {
		return nil, fmt.Errorf("AWS_ACCESS_KEY_ID is required")
	}

	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		return nil, fmt.Errorf("AWS_SECRET_ACCESS_KEY is required")
	}

	docsBucket := os.Getenv("S3_BUCKET_DOCUMENTS")
	if docsBucket == "" {
		return nil, fmt.Errorf("S3_BUCKET_DOCUMENTS is required")
	}
	backupsBucket := os.Getenv("S3_BUCKET_BACKUPS")
	if backupsBucket == "" {
		return nil, fmt.Errorf("S3_BUCKET_BACKUPS is required")
	}

	return &Config{
		App: AppConfig{
			Env:           os.Getenv("APP_ENV"),
			Port:          port,
			BaseURL:       os.Getenv("APP_BASE_URL"),
			FrontendURL:   os.Getenv("FRONTEND_URL"),
			EncryptionKey: encryptionKey,
		},
		Database: DatabaseConfig{
			DatabaseURL: dbURL,
		},
		Auth: AuthConfig{
			JWTSecret:  jwtSecret,
			TOTPIssuer: os.Getenv("TOTP_ISSUER"),
		},
		AWS: AWSConfig{
			Region:          region,
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
			BucketDocs:      docsBucket,
			BucketBackups:   backupsBucket,
		},
		Tally: TallyConfig{
			SigningSecret: os.Getenv("TALLY_SIGNING_SECRET"),
		},
	}, nil
}
