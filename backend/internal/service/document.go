package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
		if deleteErr := s.s3.Delete(ctx, key); deleteErr != nil {
			return db.Document{}, errors.Join(
				fmt.Errorf("create document record: %w", err),
				fmt.Errorf("cleanup uploaded object: %w", deleteErr),
			)
		}
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

	client := http.Client{Timeout: 90 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return db.Document{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return db.Document{}, fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return db.Document{}, fmt.Errorf("download file: unexpected status %d", resp.StatusCode)
	}

	const maxDownloadSize = 50 * 1024 * 1024 // 50 MB

	// Read maxDownloadSize+1 bytes so we can detect bodies that exceed the limit,
	// including cases where Content-Length is missing or misreported.
	buf, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadSize+1))
	if err != nil {
		return db.Document{}, fmt.Errorf("read file: %w", err)
	}
	if int64(len(buf)) > maxDownloadSize {
		slog.Error("tally file too large", "url", fileURL, "size", len(buf), "limit", maxDownloadSize)
		return db.Document{}, fmt.Errorf("file too large: body exceeds %d bytes", maxDownloadSize)
	}

	key := buildKey(clientID, dealID, filename)
	if err := s.s3.Upload(ctx, key, bytes.NewReader(buf), int64(len(buf)), resp.Header.Get("Content-Type")); err != nil {
		return db.Document{}, fmt.Errorf("s3 upload: %w", err)
	}

	doc, err := s.q.CreateDocument(ctx, db.CreateDocumentParams{
		ClientID:      nullUUIDToPg(clientID),
		DealID:        nullUUIDToPg(dealID),
		S3Key:         key,
		FileName:      filename,
		FileSizeBytes: int64ToInt4(int64(len(buf))),
		DocumentType:  docType,
	})
	if err != nil {
		if deleteErr := s.s3.Delete(ctx, key); deleteErr != nil {
			return db.Document{}, errors.Join(
				fmt.Errorf("create document: %w", err),
				fmt.Errorf("cleanup uploaded object: %w", deleteErr),
			)
		}
		return db.Document{}, fmt.Errorf("create document: %w", err)
	}

	return doc, nil
}

// PresignUpload returns a presigned S3 PUT URL so the browser can upload directly.
func (s *DocumentService) PresignUpload(
	ctx context.Context,
	filename string,
	contentType string,
	clientID uuid.NullUUID,
	dealID uuid.NullUUID,
) (uploadURL string, s3Key string, err error) {
	key := buildKey(clientID, dealID, filename)
	uploadURL, err = s.s3.PresignedUploadURL(ctx, key, contentType, 15*time.Minute)
	return uploadURL, key, err
}

// ConfirmUpload creates the DB record after the browser has uploaded directly to S3.
func (s *DocumentService) ConfirmUpload(
	ctx context.Context,
	s3Key string,
	filename string,
	size int64,
	clientID uuid.NullUUID,
	dealID uuid.NullUUID,
	docType db.DocumentType,
) (db.Document, error) {
	return s.q.CreateDocument(ctx, db.CreateDocumentParams{
		ClientID:      nullUUIDToPg(clientID),
		DealID:        nullUUIDToPg(dealID),
		S3Key:         s3Key,
		FileName:      filename,
		FileSizeBytes: int64ToInt4(size),
		DocumentType:  docType,
	})
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
	doc, err := s.q.GetDocument(ctx, toUUID(id))
	if err != nil {
		return err
	}
	if err := s.s3.Delete(ctx, doc.S3Key); err != nil {
		return err
	}
	_, err = s.q.DeleteDocument(ctx, toUUID(id))
	return err
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
