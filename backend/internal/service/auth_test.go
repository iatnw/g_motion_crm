package service

import (
	"context"
	"errors"
	"testing"

	"github.com/umbrllacrp/g_motion_crm/internal/config"
)

func TestSetupRejectsInvalidSetupSecret(t *testing.T) {
	svc := NewAuth(nil, config.AuthConfig{SetupSecret: "expected-secret"})

	err := svc.Setup(context.Background(), "wrong-secret", "newpassword12345")
	if !errors.Is(err, ErrInvalidSetupSecret) {
		t.Fatalf("Setup error = %v, want ErrInvalidSetupSecret", err)
	}
}

func TestSetupChecksPasswordAfterSetupSecret(t *testing.T) {
	svc := NewAuth(nil, config.AuthConfig{SetupSecret: "expected-secret"})

	err := svc.Setup(context.Background(), "expected-secret", "short")
	if !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("Setup error = %v, want ErrWeakPassword", err)
	}
}
