package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidTOTP        = errors.New("invalid TOTP code")
	ErrTOTPNotEnabled     = errors.New("TOTP not enabled")
	ErrTOTPAlreadyEnabled = errors.New("TOTP already enabled")
)

type AuthService struct {
	q          *db.Queries
	jwtSecret  []byte
	totpIssuer string
}

type Claims struct {
	Stage string `json:"stage"`
	jwt.RegisteredClaims
}

type LoginResult struct {
	Token        string
	TOTPRequired bool
}

func NewAuth(q *db.Queries, cfg config.AuthConfig) *AuthService {
	return &AuthService{
		q:          q,
		jwtSecret:  []byte(cfg.JWTSecret),
		totpIssuer: cfg.TOTPIssuer,
	}
}

func (s *AuthService) Login(ctx context.Context, password string) (LoginResult, error) {
	setting, err := s.q.GetSetting(ctx)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(setting.OwnerPasswordHash), []byte(password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if setting.OwnerTotpSecret.Valid {
		token, err := s.issueToken("totp_pending", 5*time.Minute)
		if err != nil {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{Token: token, TOTPRequired: true}, nil
	}

	token, err := s.issueToken("authenticated", 24*time.Hour)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	return LoginResult{Token: token, TOTPRequired: false}, nil
}

func (s *AuthService) VerifyTOTP(ctx context.Context, code string) (string, error) {
	setting, err := s.q.GetSetting(ctx)
	if err != nil {
		return "", err
	}

	if !setting.OwnerTotpSecret.Valid {
		return "", ErrTOTPNotEnabled
	}

	if !totp.Validate(code, setting.OwnerTotpSecret.String) {
		return "", ErrInvalidTOTP
	}

	return s.issueToken("authenticated", 24*time.Hour)
}

func (s *AuthService) SetUpTOTP(ctx context.Context) (secret, keyURI string, err error) {
	setting, err := s.q.GetSetting(ctx)
	if err != nil {
		return "", "", err
	}

	if setting.OwnerTotpSecret.Valid {
		return "", "", ErrTOTPAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.totpIssuer,
		AccountName: "owner",
	})
	if err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *AuthService) EnableTOTP(ctx context.Context, secret, code string) error {
	if !totp.Validate(code, secret) {
		return ErrInvalidTOTP
	}

	_, err := s.q.UpdateOwnerTOTP(ctx, toText(&secret))
	return err
}

func (s *AuthService) DisableTOTP(ctx context.Context) error {
	_, err := s.q.UpdateOwnerTOTP(ctx, toText(nil))
	return err
}

func (s *AuthService) SetPassword(ctx context.Context, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.q.UpdateOwnerPassword(ctx, string(hash))
	return err
}

func (s *AuthService) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}

func (s *AuthService) issueToken(stage string, ttl time.Duration) (string, error) {
	now := time.Now()
	c := Claims{
		Stage: stage,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.jwtSecret)
}
