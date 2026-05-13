package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

const maxUploadSize = 50 << 20 // 50 MB

type DocumentHandler struct {
	doc *service.DocumentService
}

func NewDocumentHandler(doc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{doc: doc}
}

// UploadForClient handles multipart file upload for a client.
// POST /clients/{id}/documents
func (h *DocumentHandler) UploadForClient(w http.ResponseWriter, r *http.Request) {
	clientID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	h.upload(w, r, uuid.NullUUID{UUID: clientID, Valid: true}, uuid.NullUUID{})
}

// UploadForDeal handles multipart file upload for a deal.
// POST /deals/{id}/documents
func (h *DocumentHandler) UploadForDeal(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	h.upload(w, r, uuid.NullUUID{}, uuid.NullUUID{UUID: dealID, Valid: true})
}

func (h *DocumentHandler) upload(w http.ResponseWriter, r *http.Request, clientID, dealID uuid.NullUUID) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large or bad request", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	docTypeStr := r.FormValue("document_type")
	if docTypeStr == "" {
		docTypeStr = string(db.DocumentTypeOther)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	doc, err := h.doc.Upload(
		r.Context(),
		file,
		header.Filename,
		header.Size,
		contentType,
		clientID,
		dealID,
		db.DocumentType(docTypeStr),
	)
	if err != nil {
		slog.Error("upload document", "client_id", clientID, "deal_id", dealID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toDocumentResponse(doc))
}

// GetURL returns a 15-minute presigned download URL for a document.
// GET /documents/{id}/url
func (h *DocumentHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	url, err := h.doc.PresignedURL(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get document url", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// ListByClient returns all documents for a client.
// GET /clients/{id}/documents
func (h *DocumentHandler) ListByClient(w http.ResponseWriter, r *http.Request) {
	clientID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	docs, err := h.doc.ListByClient(r.Context(), clientID)
	if err != nil {
		slog.Error("list documents by client", "client_id", clientID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toDocumentResponses(docs))
}

// ListByDeal returns all documents for a deal.
// GET /deals/{id}/documents
func (h *DocumentHandler) ListByDeal(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	docs, err := h.doc.ListByDeal(r.Context(), dealID)
	if err != nil {
		slog.Error("list documents by deal", "deal_id", dealID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toDocumentResponses(docs))
}

// PresignUpload returns a presigned S3 PUT URL for direct browser upload.
// GET /documents/presign?filename=&content_type=&document_type=&client_id=&deal_id=
func (h *DocumentHandler) PresignUpload(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filename := q.Get("filename")
	contentType := q.Get("content_type")
	if filename == "" {
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var clientID, dealID uuid.NullUUID
	if id := q.Get("client_id"); id != "" {
		parsed, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "invalid client_id", http.StatusBadRequest)
			return
		}
		clientID = uuid.NullUUID{UUID: parsed, Valid: true}
	}
	if id := q.Get("deal_id"); id != "" {
		parsed, err := uuid.Parse(id)
		if err != nil {
			http.Error(w, "invalid deal_id", http.StatusBadRequest)
			return
		}
		dealID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	uploadURL, s3Key, err := h.doc.PresignUpload(r.Context(), filename, contentType, clientID, dealID)
	if err != nil {
		slog.Error("presign upload", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"upload_url": uploadURL, "key": s3Key})
}

// ConfirmUpload creates the DB record after browser has uploaded directly to S3.
// POST /documents/confirm
func (h *DocumentHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key          string `json:"key"`
		Filename     string `json:"filename"`
		Size         int64  `json:"size"`
		DocumentType string `json:"document_type"`
		ClientID     string `json:"client_id"`
		DealID       string `json:"deal_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Key == "" || req.Filename == "" {
		http.Error(w, "key and filename are required", http.StatusBadRequest)
		return
	}
	if req.DocumentType == "" {
		req.DocumentType = string(db.DocumentTypeOther)
	}

	var clientID, dealID uuid.NullUUID
	if req.ClientID != "" {
		parsed, err := uuid.Parse(req.ClientID)
		if err != nil {
			http.Error(w, "invalid client_id", http.StatusBadRequest)
			return
		}
		clientID = uuid.NullUUID{UUID: parsed, Valid: true}
	}
	if req.DealID != "" {
		parsed, err := uuid.Parse(req.DealID)
		if err != nil {
			http.Error(w, "invalid deal_id", http.StatusBadRequest)
			return
		}
		dealID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	doc, err := h.doc.ConfirmUpload(r.Context(), req.Key, req.Filename, req.Size, clientID, dealID, db.DocumentType(req.DocumentType))
	if err != nil {
		slog.Error("confirm upload", "key", req.Key, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toDocumentResponse(doc))
}

// Delete removes a document from S3 and DB.
// DELETE /documents/{id}
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.doc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("delete document", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
