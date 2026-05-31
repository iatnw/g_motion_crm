package service

import (
	"context"
	"errors"
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

	token, err := svc.issueToken("authenticated", time.Minute)
	if err != nil {
		t.Fatalf("issueToken error = %v", err)
	}
	claims, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken error = %v", err)
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
