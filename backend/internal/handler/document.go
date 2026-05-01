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

const maxUploadSize = 20 << 20 // 20 MB

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toDocumentResponse(doc))
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDocumentResponses(docs))
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDocumentResponses(docs))
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
