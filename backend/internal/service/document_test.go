package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

func TestDirectUploadTokenRoundTrip(t *testing.T) {
	svc := &DocumentService{encryptionKey: []byte("test-key")}
	want := directUploadToken{
		Key:          "clients/test/file.pdf",
		Filename:     "file.pdf",
		ContentType:  "application/pdf",
		Size:         1234,
		DocumentType: string(db.DocumentTypeOther),
		ClientID:     "550e8400-e29b-41d4-a716-446655440000",
		ExpiresAt:    time.Now().Add(time.Minute).Unix(),
	}

	token, err := svc.signDirectUpload(want)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	got, err := svc.parseDirectUpload(token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if got != want {
		t.Fatalf("token claims = %+v, want %+v", got, want)
	}
}

func TestDirectUploadTokenRejectsTampering(t *testing.T) {
	svc := &DocumentService{encryptionKey: []byte("test-key")}
	token, err := svc.signDirectUpload(directUploadToken{
		Key:          "clients/test/file.pdf",
		Filename:     "file.pdf",
		Size:         1234,
		DocumentType: string(db.DocumentTypeOther),
		ExpiresAt:    time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	parts := strings.Split(token, ".")
	parts[0] = parts[0] + "a"
	_, err = svc.parseDirectUpload(strings.Join(parts, "."))
	if !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("parse tampered token error = %v, want ErrInvalidUpload", err)
	}
}

func TestDirectUploadTokenRejectsExpiredToken(t *testing.T) {
	svc := &DocumentService{encryptionKey: []byte("test-key")}
	token, err := svc.signDirectUpload(directUploadToken{
		Key:          "clients/test/file.pdf",
		Filename:     "file.pdf",
		Size:         1234,
		DocumentType: string(db.DocumentTypeOther),
		ExpiresAt:    time.Now().Add(-time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = svc.parseDirectUpload(token)
	if !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("parse expired token error = %v, want ErrInvalidUpload", err)
	}
}

func TestValidDocumentTypeRejectsUnknownType(t *testing.T) {
	if validDocumentType(db.DocumentType("unknown")) {
		t.Fatal("expected unknown document type to be rejected")
	}
}
