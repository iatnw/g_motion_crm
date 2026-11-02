package storage

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
)

func TestS3EndpointConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, publicEndpoint, wantHost string
		pathStyle                                bool
	}{
		{name: "AWS defaults"},
		{name: "local internal and browser endpoints", endpoint: "http://storage:9090", publicEndpoint: "http://localhost:9000", wantHost: "localhost:9000", pathStyle: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewS3Client(config.AWSConfig{Region: "us-east-1", AccessKeyID: "test", SecretAccessKey: "test-secret", BucketDocs: "documents", S3Endpoint: tc.endpoint, S3PublicEndpoint: tc.publicEndpoint})
			if err != nil {
				t.Fatal(err)
			}
			opts := client.client.Options()
			if opts.UsePathStyle != tc.pathStyle {
				t.Fatalf("UsePathStyle=%v", opts.UsePathStyle)
			}
			if tc.endpoint == "" {
				if opts.BaseEndpoint != nil {
					t.Fatal("AWS endpoint was overridden")
				}
				return
			}
			if opts.BaseEndpoint == nil || *opts.BaseEndpoint != tc.endpoint {
				t.Fatal("internal endpoint not configured")
			}
			signed, err := client.PresignedURL(context.Background(), "demo.txt", time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			u, err := url.Parse(signed)
			if err != nil {
				t.Fatal(err)
			}
			if u.Host != tc.wantHost || u.Path != "/documents/demo.txt" {
				t.Fatalf("incorrect browser URL: host=%s path=%s", u.Host, u.Path)
			}
		})
	}
}

func TestLocalStorageRoundTrip(t *testing.T) {
	endpoint := os.Getenv("LOCAL_STORAGE_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("requires local S3 simulator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := NewS3Client(config.AWSConfig{Region: os.Getenv("AWS_REGION"), AccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"), SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"), BucketDocs: os.Getenv("S3_BUCKET_DOCUMENTS"), S3Endpoint: endpoint, S3PublicEndpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	key := "storage-tests/" + uuid.NewString() + ".txt"
	body := []byte("Fictional CRM storage test")
	if err := client.Upload(ctx, key, bytes.NewReader(body), int64(len(body)), "text/plain"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Delete(context.Background(), key); err != nil {
			t.Error(err)
		}
	})
	info, err := client.Head(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != int64(len(body)) || info.ContentType != "text/plain" {
		t.Fatalf("incorrect metadata: %+v", info)
	}
	// Exercise the browser's direct-upload path as well as server-side uploads.
	signed, err := client.PresignedUploadURL(ctx, key, "text/plain", int64(len(body)), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signed, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")
	httpClient := &http.Client{Timeout: 15 * time.Second}
	res, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		t.Fatalf("presigned upload status %d", res.StatusCode)
	}
	signed, err = client.PresignedURL(ctx, key, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, signed, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err = httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 || !bytes.Equal(got, body) {
		t.Fatalf("download mismatch, status %d", res.StatusCode)
	}
}
