package service

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"testing"
	"time"

	"github.com/umbrllacrp/g_motion_crm/internal/config"
)

func TestSetupRejectsInvalidSetupSecret(t *testing.T) {
	svc := NewAuth(nil, config.AuthConfig{SetupSecret: "expected-secret"})

	err := svc.Setup(context.Background(), "wrong-secret", "newpassword12345")
	if !errors.Is(err, ErrInvalidSetupSecret) {
		t.Fatalf("Setup error = %v, want ErrInvalidSetupSecret", err)
	}
}

func TestIssueTokenIncludesOwnerSubject(t *testing.T) {
	svc := NewAuth(nil, config.AuthConfig{JWTSecret: "test-secret"})

	token, err := svc.issueToken("authenticated", time.Minute, 1)
	if err != nil {
		t.Fatalf("issueToken error = %v", err)
	}
	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken error = %v", err)
	}
	if claims.SessionVersion != 1 {
		t.Fatalf("session version = %d, want 1", claims.SessionVersion)
	}
	if claims.Subject != "owner" {
		t.Fatalf("subject = %q, want owner", claims.Subject)
	}
}

func TestSetupChecksPasswordAfterSetupSecret(t *testing.T) {
	svc := NewAuth(nil, config.AuthConfig{SetupSecret: "expected-secret"})

	err := svc.Setup(context.Background(), "expected-secret", "short")
	if !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("Setup error = %v, want ErrWeakPassword", err)
	}
}

// Embed DBTX so an unexpected database operation fails instead of being silently mocked.
type sessionDB struct {
	db.DBTX
	version int64
	err     error
}

func (d sessionDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return sessionRow{d.version, d.err}
}

type sessionRow struct {
	version int64
	err     error
}

func (r sessionRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[5].(*int64) = r.version
	return nil
}

func TestSessionValid(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version int64
		dbErr   error
		want    bool
	}{
		{"current", 3, nil, true},
		{"revoked", 2, nil, false},
		{"missing version", 0, nil, false},
		{"database failure", 3, errors.New("database unavailable"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewAuth(db.New(sessionDB{version: 3, err: tc.dbErr}), config.AuthConfig{})
			valid, err := svc.SessionValid(context.Background(), &Claims{SessionVersion: tc.version})
			if valid != tc.want || !errors.Is(err, tc.dbErr) {
				t.Fatalf("got (%v, %v), want (%v, %v)", valid, err, tc.want, tc.dbErr)
			}
		})
	}
}

func TestLegacyTokenHasNoValidSession(t *testing.T) {
	svc := NewAuth(db.New(sessionDB{version: 1}), config.AuthConfig{JWTSecret: "test-secret"})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "owner", "stage": "authenticated", "exp": time.Now().Add(time.Minute).Unix(),
	}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	valid, err := svc.SessionValid(context.Background(), claims)
	if err != nil || valid {
		t.Fatalf("legacy session: valid=%v, err=%v", valid, err)
	}
}
