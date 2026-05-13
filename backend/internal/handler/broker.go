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

type BrokerHandler struct {
	broker *service.BrokerService
}

func NewBrokerHandler(broker *service.BrokerService) *BrokerHandler {
	return &BrokerHandler{broker: broker}
}

type brokerRequest struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
	Notes     *string `json:"notes"`
}

func (h *BrokerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req brokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" {
		http.Error(w, "first_name is required", http.StatusBadRequest)
		return
	}

	broker, err := h.broker.Create(r.Context(), service.CreateBrokerInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Email:     req.Email,
		Notes:     req.Notes,
	})
	if err != nil {
		slog.Error("create broker", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toBrokerResponse(broker))
}

func (h *BrokerHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	brokers, err := h.broker.List(r.Context(), limit+1, offset)
	if err != nil {
		slog.Error("list brokers", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	brokers, hasMore := paginate(brokers, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"data":     toBrokerResponses(brokers),
		"limit":    limit,
		"offset":   offset,
		"has_more": hasMore,
	})
}

func (h *BrokerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	broker, err := h.broker.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get broker", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toBrokerResponse(broker))
}

func (h *BrokerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.broker.Delete(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("delete broker", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BrokerHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	stats, err := h.broker.GetStats(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get broker stats", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toBrokerStatResponse(stats))
}

func (h *BrokerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req brokerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" {
		http.Error(w, "first_name is required", http.StatusBadRequest)
		return
	}

	broker, err := h.broker.Update(r.Context(), service.UpdateBrokerInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Email:     req.Email,
		Notes:     req.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("update broker", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toBrokerResponse(broker))
}
