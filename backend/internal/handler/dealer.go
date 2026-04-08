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

type DealerHandler struct {
	dealer *service.DealerService
}

func NewDealerHandler(dealer *service.DealerService) *DealerHandler {
	return &DealerHandler{dealer: dealer}
}

type dealerRequest struct {
	Name          string  `json:"name"`
	ContactPerson *string `json:"contact_person"`
	Phone         *string `json:"phone"`
	Address       *string `json:"address"`
	Notes         *string `json:"notes"`
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
		Name:          req.Name,
		ContactPerson: nullString(req.ContactPerson),
		Phone:         nullString(req.Phone),
		Address:       nullString(req.Address),
		Notes:         nullString(req.Notes),
	})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dealer)
}

func (h *DealerHandler) List(w http.ResponseWriter, r *http.Request) {
	dealers, err := h.dealer.List(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dealers)
}

func (h *DealerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	dealer, err := h.dealer.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dealer)
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
		ID:            id,
		Name:          req.Name,
		ContactPerson: nullString(req.ContactPerson),
		Phone:         nullString(req.Phone),
		Address:       nullString(req.Address),
		Notes:         nullString(req.Notes),
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
	json.NewEncoder(w).Encode(dealer)
}
