package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/storage"
)

const (
	maxDirectUploadSize = 50 << 20
	uploadTokenTTL      = 15 * time.Minute
)

var ErrInvalidUpload = errors.New("invalid upload")

type DocumentService struct {
	q             *db.Queries
	s3            *storage.S3Client
	encryptionKey []byte
}

type directUploadToken struct {
	Key          string `json:"key"`
	Filename     string `json:"filename"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	DocumentType string `json:"document_type"`
	ClientID     string `json:"client_id,omitempty"`
	DealID       string `json:"deal_id,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}

func NewDocument(q *db.Queries, s3 *storage.S3Client, encryptionKey []byte) *DocumentService {
	return &DocumentService{q: q, s3: s3, encryptionKey: encryptionKey}
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
	folder, label := s.resolveLabel(ctx, clientID, dealID)
	key := buildKey(folder, label, filename)

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
	if err := validateTallyDocumentURL(u); err != nil {
		return db.Document{}, err
	}

	client := newTallyDownloadClient()
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

	folder, label := s.resolveLabel(ctx, clientID, dealID)
	key := buildKey(folder, label, filename)
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

// Apply the same restrictions to the initial URL and every redirect target.
func validateTallyDocumentURL(u *url.URL) error {
	if u.Scheme != "https" {
		return fmt.Errorf("only https URLs are supported")
	}
	if u.User != nil {
		return fmt.Errorf("URL credentials are not supported")
	}
	host := strings.ToLower(u.Hostname())
	if host != "tally.so" && !strings.HasSuffix(host, ".tally.so") {
		return fmt.Errorf("only Tally URLs are supported")
	}
	if u.Port() != "" && u.Port() != "443" {
		return fmt.Errorf("only the standard HTTPS port is supported")
	}
	return nil
}

func newTallyDownloadClient() *http.Client {
	return &http.Client{
		Timeout: 90 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return validateTallyDocumentURL(req.URL)
		},
	}
}

// PresignUpload returns a presigned S3 PUT URL so the browser can upload directly.
func (s *DocumentService) PresignUpload(
	ctx context.Context,
	filename string,
	contentType string,
	size int64,
	docType db.DocumentType,
	clientID uuid.NullUUID,
	dealID uuid.NullUUID,
) (uploadURL string, uploadToken string, err error) {
	if err := s.validateDirectUpload(ctx, filename, size, docType, clientID, dealID); err != nil {
		return "", "", err
	}
	folder, label := s.resolveLabel(ctx, clientID, dealID)
	key := buildKey(folder, label, filename)
	uploadURL, err = s.s3.PresignedUploadURL(ctx, key, contentType, size, uploadTokenTTL)
	if err != nil {
		return "", "", err
	}
	uploadToken, err = s.signDirectUpload(directUploadToken{
		Key:          key,
		Filename:     filename,
		ContentType:  contentType,
		Size:         size,
		DocumentType: string(docType),
		ClientID:     nullUUIDString(clientID),
		DealID:       nullUUIDString(dealID),
		ExpiresAt:    time.Now().Add(uploadTokenTTL).Unix(),
	})
	return uploadURL, uploadToken, err
}

// ConfirmUpload creates the DB record after the browser has uploaded directly to S3.
func (s *DocumentService) ConfirmUpload(
	ctx context.Context,
	uploadToken string,
) (db.Document, error) {
	claims, err := s.parseDirectUpload(uploadToken)
	if err != nil {
		return db.Document{}, err
	}
	info, err := s.s3.Head(ctx, claims.Key)
	if err != nil {
		return db.Document{}, fmt.Errorf("verify uploaded object: %w", err)
	}
	if info.Size != claims.Size {
		return db.Document{}, fmt.Errorf("%w: uploaded object size does not match signed size", ErrInvalidUpload)
	}
	if info.ContentType != claims.ContentType {
		return db.Document{}, fmt.Errorf("%w: uploaded object content type does not match signed content type", ErrInvalidUpload)
	}
	clientID, err := parseNullUUID(claims.ClientID)
	if err != nil {
		return db.Document{}, err
	}
	dealID, err := parseNullUUID(claims.DealID)
	if err != nil {
		return db.Document{}, err
	}
	return s.q.CreateDocument(ctx, db.CreateDocumentParams{
		ClientID:      nullUUIDToPg(clientID),
		DealID:        nullUUIDToPg(dealID),
		S3Key:         claims.Key,
		FileName:      claims.Filename,
		FileSizeBytes: int64ToInt4(claims.Size),
		DocumentType:  db.DocumentType(claims.DocumentType),
	})
}

func (s *DocumentService) validateDirectUpload(ctx context.Context, filename string, size int64, docType db.DocumentType, clientID, dealID uuid.NullUUID) error {
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("%w: filename is required", ErrInvalidUpload)
	}
	if size <= 0 || size > maxDirectUploadSize {
		return fmt.Errorf("%w: file size must be between 1 and %d bytes", ErrInvalidUpload, maxDirectUploadSize)
	}
	if !validDocumentType(docType) {
		return fmt.Errorf("%w: invalid document type", ErrInvalidUpload)
	}
	if clientID.Valid == dealID.Valid {
		return fmt.Errorf("%w: exactly one document owner is required", ErrInvalidUpload)
	}
	if clientID.Valid {
		if _, err := s.q.GetClient(ctx, toUUID(clientID.UUID)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: client not found", ErrInvalidUpload)
			}
			return fmt.Errorf("get client: %w", err)
		}
	}
	if dealID.Valid {
		if _, err := s.q.GetDeal(ctx, toUUID(dealID.UUID)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: deal not found", ErrInvalidUpload)
			}
			return fmt.Errorf("get deal: %w", err)
		}
	}
	return nil
}

func (s *DocumentService) signDirectUpload(claims directUploadToken) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	sig := hmac.New(sha256.New, s.encryptionKey)
	_, _ = sig.Write(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(sig.Sum(nil)), nil
}

func (s *DocumentService) parseDirectUpload(token string) (directUploadToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return directUploadToken{}, fmt.Errorf("%w: malformed upload token", ErrInvalidUpload)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return directUploadToken{}, fmt.Errorf("%w: malformed upload token", ErrInvalidUpload)
	}
	gotSig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return directUploadToken{}, fmt.Errorf("%w: malformed upload token", ErrInvalidUpload)
	}
	wantSig := hmac.New(sha256.New, s.encryptionKey)
	_, _ = wantSig.Write(payload)
	if !hmac.Equal(gotSig, wantSig.Sum(nil)) {
		return directUploadToken{}, fmt.Errorf("%w: invalid upload token signature", ErrInvalidUpload)
	}
	var claims directUploadToken
	if err := json.Unmarshal(payload, &claims); err != nil {
		return directUploadToken{}, fmt.Errorf("%w: malformed upload token", ErrInvalidUpload)
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return directUploadToken{}, fmt.Errorf("%w: upload token expired", ErrInvalidUpload)
	}
	if claims.Key == "" || claims.Filename == "" || claims.Size <= 0 || claims.Size > maxDirectUploadSize || !validDocumentType(db.DocumentType(claims.DocumentType)) {
		return directUploadToken{}, fmt.Errorf("%w: invalid upload token claims", ErrInvalidUpload)
	}
	if (claims.ClientID == "") == (claims.DealID == "") {
		return directUploadToken{}, fmt.Errorf("%w: invalid upload token owner", ErrInvalidUpload)
	}
	return claims, nil
}

func validDocumentType(docType db.DocumentType) bool {
	switch docType {
	case db.DocumentTypeCreditAppPdf,
		db.DocumentTypeDriverLicense,
		db.DocumentTypeInsurance,
		db.DocumentTypeBrokerAgreement,
		db.DocumentTypeDealerDocs,
		db.DocumentTypeCheckCopy,
		db.DocumentTypeOther:
		return true
	default:
		return false
	}
}

func nullUUIDString(id uuid.NullUUID) string {
	if !id.Valid {
		return ""
	}
	return id.UUID.String()
}

func parseNullUUID(value string) (uuid.NullUUID, error) {
	if value == "" {
		return uuid.NullUUID{}, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.NullUUID{}, fmt.Errorf("%w: invalid owner id", ErrInvalidUpload)
	}
	return uuid.NullUUID{UUID: id, Valid: true}, nil
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

// resolveLabel returns an S3 folder and human-readable label for the upload.
// For deals it uses the deal number; for clients it uses the client name.
func (s *DocumentService) resolveLabel(ctx context.Context, clientID, dealID uuid.NullUUID) (folder, label string) {
	if dealID.Valid {
		deal, err := s.q.GetDeal(ctx, toUUID(dealID.UUID))
		if err == nil && deal.DealNumber.Valid {
			return "deals", sanitizeName(deal.DealNumber.String)
		}
		return "deals", dealID.UUID.String()[:8]
	}
	if clientID.Valid {
		client, err := s.q.GetClient(ctx, toUUID(clientID.UUID))
		if err == nil {
			name := strings.TrimSpace(client.FirstName + "-" + client.LastName)
			if name != "-" {
				return "clients", sanitizeName(name)
			}
		}
		return "clients", clientID.UUID.String()[:8]
	}
	return "documents", "misc"
}

func buildKey(folder, label, filename string) string {
	ext := filepath.Ext(filename)
	base := sanitizeName(strings.TrimSuffix(filepath.Base(filename), ext))
	date := time.Now().UTC().Format("2006-01-02")
	unique := uuid.New().String()[:6]
	return fmt.Sprintf("%s/%s/%s_%s_%s%s", folder, label, date, base, unique, ext)
}

var reSanitize = regexp.MustCompile(`[^a-z0-9]+`)

func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = reSanitize.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 60 {
		s = s[:60]
	}
	return s
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
