package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type DealerHandler struct {
	dealer *service.DealerService
}

func NewDealerHandler(dealer *service.DealerService) *DealerHandler {
	return &DealerHandler{dealer: dealer}
}

type dealerRequest struct {
	Name               string  `json:"name"`
	ContactPerson      *string `json:"contact_person"`
	Phone              *string `json:"phone"`
	Address            *string `json:"address"`
	Notes              *string `json:"notes"`
	ContactPersonEmail *string `json:"contact_person_email"`
}

func (h *DealerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dealerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	dealer, err := h.dealer.Create(r.Context(), service.CreateDealerInput{
		Name:               req.Name,
		ContactPerson:      req.ContactPerson,
		Phone:              req.Phone,
		Address:            req.Address,
		Notes:              req.Notes,
		ContactPersonEmail: req.ContactPersonEmail,
	})
	if err != nil {
		slog.Error("create dealer", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toDealerResponse(dealer))
}

func (h *DealerHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	dealers, err := h.dealer.List(r.Context(), limit+1, offset)
	if err != nil {
		slog.Error("list dealers", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	dealers, hasMore := paginate(dealers, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"data":     toDealerResponses(dealers),
		"limit":    limit,
		"offset":   offset,
		"has_more": hasMore,
	})
}

func (h *DealerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	dealer, err := h.dealer.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get dealer", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toDealerResponse(dealer))
}

func (h *DealerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.dealer.Delete(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("delete dealer", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DealerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req dealerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	dealer, err := h.dealer.Update(r.Context(), service.UpdateDealerInput{
		ID:                 id,
		Name:               req.Name,
		ContactPerson:      req.ContactPerson,
		Phone:              req.Phone,
		Address:            req.Address,
		Notes:              req.Notes,
		ContactPersonEmail: req.ContactPersonEmail,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("update dealer", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toDealerResponse(dealer))
}

func (h *DealerHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	stats, err := h.dealer.GetStats(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get dealer stats", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toDealerStatResponse(stats))
}
