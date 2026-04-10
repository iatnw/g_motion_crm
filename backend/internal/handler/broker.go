package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	if req.FirstName == "" || req.LastName == "" {
		http.Error(w, "first_name and last_name are required", http.StatusBadRequest)
		return
	}

	broker, err := h.broker.Create(r.Context(), service.CreateBrokerInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     nullString(req.Phone),
		Email:     nullString(req.Email),
		Notes:     nullString(req.Notes),
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toBrokerResponse(broker))
}

func (h *BrokerHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	brokers, err := h.broker.List(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data":     toBrokerResponses(brokers),
		"limit":    limit,
		"offset":   offset,
		"has_more": len(brokers) == int(limit),
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
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toBrokerResponse(broker))
}

func (h *BrokerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.broker.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	if req.FirstName == "" || req.LastName == "" {
		http.Error(w, "first_name and last_name are required", http.StatusBadRequest)
		return
	}

	broker, err := h.broker.Update(r.Context(), service.UpdateBrokerInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     nullString(req.Phone),
		Email:     nullString(req.Email),
		Notes:     nullString(req.Notes),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toBrokerResponse(broker))
}
