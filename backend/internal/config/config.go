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
	Env     string
	Port    string
	BaseURL string
}

type DatabaseConfig struct {
	DatabaseURL string
}

type AWSConfig struct{}

type AuthConfig struct {
	JWTSecret  string
	TOTPIssuer string
}

type TwilioConfig struct{}

type TallyConfig struct{}

func Load() (*Config, error) {
	if os.Getenv("APP_ENV") == "" {
		godotenv.Load()
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

	return &Config{
		App: AppConfig{
			Env:     os.Getenv("APP_ENV"),
			Port:    port,
			BaseURL: os.Getenv("APP_BASE_URL"),
		},
		Database: DatabaseConfig{
			DatabaseURL: dbURL,
		},
		Auth: AuthConfig{
			JWTSecret:  jwtSecret,
			TOTPIssuer: os.Getenv("TOTP_ISSUER"),
		},
	}, nil
}
