package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/storage"
)

type DocumentService struct {
	q  *db.Queries
	s3 *storage.S3Client
}

func NewDocument(q *db.Queries, s3 *storage.S3Client) *DocumentService {
	return &DocumentService{q: q, s3: s3}
}

// Upload saves a file from an io.Reader to S3 and records it in the DB.
func (s *DocumentService) Upload(
	ctx context.Context,
	r io.Reader,
	filename string,
	size int64,
	contentType string,
	clientID uuid.NullUUID,
	dealID uuid.NullUUID,
	docType db.DocumentType,
) (db.Document, error) {
	key := buildKey(clientID, dealID, filename)

	if err := s.s3.Upload(ctx, key, r, size, contentType); err != nil {
		return db.Document{}, fmt.Errorf("s3 upload: %w", err)
	}

	doc, err := s.q.CreateDocument(ctx, db.CreateDocumentParams{
		ClientID:      nullUUIDToPg(clientID),
		DealID:        nullUUIDToPg(dealID),
		S3Key:         key,
		FileName:      filename,
		FileSizeBytes: int64ToInt4(size),
		DocumentType:  docType,
	})
	if err != nil {
		return db.Document{}, fmt.Errorf("create document record: %w", err)
	}

	return doc, nil
}

// UploadFromURL downloads a file from a remote URL, uploads to S3, and records in DB.
func (s *DocumentService) UploadFromURL(
	ctx context.Context,
	fileURL string,
	filename string,
	clientID uuid.NullUUID,
	dealID uuid.NullUUID,
	docType db.DocumentType,
) (db.Document, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return db.Document{}, fmt.Errorf("parse file URL: %w", err)
	}
	if u.Scheme != "https" {
		return db.Document{}, fmt.Errorf("only https URLs are supported")
	}
	if !strings.HasSuffix(u.Host, ".tally.so") && u.Host != "tally.so" {
		return db.Document{}, fmt.Errorf("only Tally URLs are supported")
	}

	client := http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return db.Document{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return db.Document{}, fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	key := buildKey(clientID, dealID, filename)
	if err := s.s3.Upload(ctx, key, resp.Body, resp.ContentLength, resp.Header.Get("Content-Type")); err != nil {
		return db.Document{}, fmt.Errorf("s3 upload: %w", err)
	}

	doc, err := s.q.CreateDocument(ctx, db.CreateDocumentParams{
		ClientID:      nullUUIDToPg(clientID),
		DealID:        nullUUIDToPg(dealID),
		S3Key:         key,
		FileName:      filename,
		FileSizeBytes: int64ToInt4(resp.ContentLength),
		DocumentType:  docType,
	})
	if err != nil {
		return db.Document{}, fmt.Errorf("create document: %w", err)
	}

	return doc, nil
}

// PresignedURL returns a time-limited download URL for a document.
func (s *DocumentService) PresignedURL(ctx context.Context, id uuid.UUID) (string, error) {
	doc, err := s.q.GetDocument(ctx, toUUID(id))
	if err != nil {
		return "", err
	}
	return s.s3.PresignedURL(ctx, doc.S3Key, 15*time.Minute)
}

func (s *DocumentService) ListByClient(ctx context.Context, clientID uuid.UUID) ([]db.Document, error) {
	return s.q.ListDocumentsByClient(ctx, toUUID(clientID))
}

func (s *DocumentService) ListByDeal(ctx context.Context, dealID uuid.UUID) ([]db.Document, error) {
	return s.q.ListDocumentsByDeal(ctx, toUUID(dealID))
}

func (s *DocumentService) Delete(ctx context.Context, id uuid.UUID) error {
	doc, err := s.q.DeleteDocument(ctx, toUUID(id))
	if err != nil {
		return err
	}
	return s.s3.Delete(ctx, doc.S3Key)
}

func buildKey(clientID uuid.NullUUID, dealID uuid.NullUUID, filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	ts := time.Now().Unix()
	unique := uuid.New().String()[:8]
	name := fmt.Sprintf("%s_%d_%s%s", base, ts, unique, ext)

	if dealID.Valid {
		return fmt.Sprintf("deals/%s/%s", dealID.UUID, name)
	}
	if clientID.Valid {
		return fmt.Sprintf("clients/%s/%s", clientID.UUID, name)
	}
	return fmt.Sprintf("documents/%s", name)
}

func nullUUIDToPg(id uuid.NullUUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id.UUID, Valid: id.Valid}
}

func int64ToInt4(n int64) pgtype.Int4 {
	if n <= 0 || n > 2147483647 {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(n), Valid: true}
}
