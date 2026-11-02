package service

import (
	"errors"
	"io"
	"net/http"
	"net/url"
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

func TestValidateTallyDocumentURL(t *testing.T) {
	for _, tc := range []struct {
		url     string
		allowed bool
	}{
		{"https://tally.so/file", true},
		{"https://storage.tally.so/file", true},
		{"https://STORAGE.TALLY.SO:443/file", true},
		{"http://storage.tally.so/file", false},
		{"https://example.com/file", false},
		{"https://tally.so.example.com/file", false},
		{"https://eviltally.so/file", false},
		{"https://127.0.0.1/file", false},
		{"https://169.254.169.254/file", false},
		{"https://tally.so@localhost/file", false},
		{"https://user:password@tally.so/file", false},
		{"https://tally.so:8443/file", false},
		{"/relative", false},
	} {
		t.Run(tc.url, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			if err != nil {
				t.Fatal(err)
			}
			err = validateTallyDocumentURL(u)
			if (err == nil) != tc.allowed {
				t.Fatalf("allowed=%v, err=%v", tc.allowed, err)
			}
		})
	}
}

type documentRoundTripper func(*http.Request) (*http.Response, error)

func (f documentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestTallyDownloadRedirects(t *testing.T) {
	for _, tc := range []struct {
		name         string
		targets      []string
		wantRequests int
		wantError    bool
	}{
		{"allowed subdomain", []string{"https://storage.tally.so/file"}, 2, false},
		{"relative redirect", []string{"/file"}, 2, false},
		{"foreign host", []string{"https://example.com/file"}, 1, true},
		{"HTTP downgrade", []string{"http://storage.tally.so/file"}, 1, true},
		{"internal address", []string{"https://127.0.0.1/file"}, 1, true},
		{"blocked second hop", []string{"https://storage.tally.so/file", "https://169.254.169.254/file"}, 2, true},
		{"redirect loop", []string{"/1", "/2", "/3", "/4", "/5", "/6", "/7", "/8", "/9", "/10"}, 10, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := newTallyDownloadClient()
			requests := 0
			client.Transport = documentRoundTripper(func(req *http.Request) (*http.Response, error) {
				requests++
				res := &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("file")), Request: req}
				if requests <= len(tc.targets) {
					res.StatusCode = 302
					res.Header.Set("Location", tc.targets[requests-1])
				}
				return res, nil
			})
			res, err := client.Get("https://tally.so/start")
			if res != nil {
				res.Body.Close()
			}
			if (err != nil) != tc.wantError {
				t.Fatalf("err=%v, wantError=%v", err, tc.wantError)
			}
			if requests != tc.wantRequests {
				t.Fatalf("requests=%d, want %d", requests, tc.wantRequests)
			}
		})
	}
}
