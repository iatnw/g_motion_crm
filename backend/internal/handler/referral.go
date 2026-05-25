package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type ReferralHandler struct {
	referral *service.ReferralService
}

func NewReferralHandler(referral *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{referral: referral}
}

type referralRequest struct {
	DealID    *string  `json:"deal_id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phone     *string  `json:"phone"`
	Amount    *float64 `json:"amount"`
	Status    *string  `json:"status"`
	Notes     *string  `json:"notes"`
}

func (h *ReferralHandler) Create(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "dealID"))
	if err != nil {
		http.Error(w, "invalid deal_id", http.StatusBadRequest)
		return
	}

	var req referralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" {
		http.Error(w, "first_name is required", http.StatusBadRequest)
		return
	}

	status, err := nullReferralStatus(req.Status)
	if err != nil {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	referral, err := h.referral.Create(r.Context(), service.CreateReferralInput{
		DealID:    &dealID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Amount:    req.Amount,
		Status:    status,
		Notes:     req.Notes,
	})
	if err != nil {
		slog.Error("create referral", "deal_id", dealID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toReferralResponse(referral))
}

func (h *ReferralHandler) CreateDirect(w http.ResponseWriter, r *http.Request) {
	var req referralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" {
		http.Error(w, "first_name is required", http.StatusBadRequest)
		return
	}

	status, err := nullReferralStatus(req.Status)
	if err != nil {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	var dealID *uuid.UUID
	if req.DealID != nil && *req.DealID != "" {
		id, err := uuid.Parse(*req.DealID)
		if err != nil {
			http.Error(w, "invalid deal_id", http.StatusBadRequest)
			return
		}
		dealID = &id
	}

	referral, err := h.referral.Create(r.Context(), service.CreateReferralInput{
		DealID:    dealID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Amount:    req.Amount,
		Status:    status,
		Notes:     req.Notes,
	})
	if err != nil {
		slog.Error("create referral", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toReferralResponse(referral))
}

func (h *ReferralHandler) List(w http.ResponseWriter, r *http.Request) {
	if dealIDsParam := strings.TrimSpace(r.URL.Query().Get("deal_ids")); dealIDsParam != "" {
		parts := strings.Split(dealIDsParam, ",")
		dealIDs := make([]uuid.UUID, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := uuid.Parse(part)
			if err != nil {
				http.Error(w, "invalid deal_ids", http.StatusBadRequest)
				return
			}
			dealIDs = append(dealIDs, id)
		}

		if len(dealIDs) == 0 {
			writeJSON(w, http.StatusOK, []ReferralResponse{})
			return
		}

		referrals, err := h.referral.ListByDeals(r.Context(), dealIDs)
		if err != nil {
			slog.Error("list referrals by deals", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, toReferralResponses(referrals))
		return
	}

	referrals, err := h.referral.ListAll(r.Context())
	if err != nil {
		slog.Error("list all referrals", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toReferralWithDealResponses(referrals))
}

func (h *ReferralHandler) ListByDeal(w http.ResponseWriter, r *http.Request) {
	dealID, err := uuid.Parse(chi.URLParam(r, "dealID"))
	if err != nil {
		http.Error(w, "invalid deal_id", http.StatusBadRequest)
		return
	}

	referrals, err := h.referral.List(r.Context(), dealID)
	if err != nil {
		slog.Error("list referrals by deal", "deal_id", dealID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toReferralResponses(referrals))
}

func (h *ReferralHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	referral, err := h.referral.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get referral", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toReferralResponse(referral))
}

func (h *ReferralHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req referralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" {
		http.Error(w, "first_name is required", http.StatusBadRequest)
		return
	}

	var dealID *uuid.UUID
	if req.DealID != nil && *req.DealID != "" {
		parsed, err := uuid.Parse(*req.DealID)
		if err != nil {
			http.Error(w, "invalid deal_id", http.StatusBadRequest)
			return
		}
		dealID = &parsed
	}

	referral, err := h.referral.Update(r.Context(), service.UpdateReferralInput{
		ID:        id,
		DealID:    dealID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Amount:    req.Amount,
		Notes:     req.Notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("update referral", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toReferralResponse(referral))
}

func (h *ReferralHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}
	status, err := nullReferralStatus(&req.Status)
	if err != nil {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	referral, err := h.referral.UpdateStatus(r.Context(), id, status.ReferralStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("update referral status", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toReferralResponse(referral))
}

func (h *ReferralHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.referral.Delete(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("delete referral", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
