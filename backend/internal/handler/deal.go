package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type DealHandler struct {
	deal *service.DealService
}

func NewDealHandler(deal *service.DealService) *DealHandler {
	return &DealHandler{deal: deal}
}

type dealRequest struct {
	ClientID               string   `json:"client_id"`
	BrokerID               *string  `json:"broker_id"`
	DealerID               *string  `json:"dealer_id"`
	DealDate               *string  `json:"deal_date"`
	DealNumber             *string  `json:"deal_number"`
	Vin                    *string  `json:"vin"`
	Make                   *string  `json:"make"`
	Model                  *string  `json:"model"`
	TrimLevel              *string  `json:"trim_level"`
	DownPayment            *float64 `json:"down_payment"`
	DownPaymentMethod      *string  `json:"down_payment_method"`
	MonthlyPayment         *float64 `json:"monthly_payment"`
	BrokerFee              *float64 `json:"broker_fee"`
	BrokerPercent          *float64 `json:"broker_percent"`
	BrokerCommission       *float64 `json:"broker_commission"`
	BrokerCommissionStatus *string  `json:"broker_commission_status"`
	YourNet                *float64 `json:"your_net"`
	BrokerFeeStatus        *string  `json:"broker_fee_status"`
	DealStatus             *string  `json:"deal_status"`
	CollectAmount          *float64 `json:"collect_amount"`
	DocFee                 *float64 `json:"doc_fee"`
	BrokerPaymentMethod    *string  `json:"broker_payment_method"`
	Notes                  *string  `json:"notes"`
}

func (h *DealHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.ClientID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return
	}

	brokerID, err := nullUUID(req.BrokerID)
	if err != nil {
		http.Error(w, "invalid broker_id", http.StatusBadRequest)
		return
	}

	dealerID, err := nullUUID(req.DealerID)
	if err != nil {
		http.Error(w, "invalid dealer_id", http.StatusBadRequest)
		return
	}

	dealDate, err := nullTime(req.DealDate)
	if err != nil {
		http.Error(w, "invalid deal_date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	downPaymentMethod, err := nullDownPaymentMethod(req.DownPaymentMethod)
	if err != nil {
		http.Error(w, "invalid down_payment_method", http.StatusBadRequest)
		return
	}
	brokerCommissionStatus, err := nullPaymentStatus(req.BrokerCommissionStatus)
	if err != nil {
		http.Error(w, "invalid broker_commission_status", http.StatusBadRequest)
		return
	}
	brokerFeeStatus, err := nullPaymentStatus(req.BrokerFeeStatus)
	if err != nil {
		http.Error(w, "invalid broker_fee_status", http.StatusBadRequest)
		return
	}
	dealStatus, err := nullDealStatus(req.DealStatus)
	if err != nil {
		http.Error(w, "invalid deal_status", http.StatusBadRequest)
		return
	}
	brokerPaymentMethod, err := nullBrokerPaymentMethod(req.BrokerPaymentMethod)
	if err != nil {
		http.Error(w, "invalid broker_payment_method", http.StatusBadRequest)
		return
	}

	deal, err := h.deal.Create(r.Context(), service.CreateDealInput{
		ClientID:               clientID,
		BrokerID:               brokerID,
		DealerID:               dealerID,
		DealDate:               dealDate,
		DealNumber:             req.DealNumber,
		Vin:                    req.Vin,
		Make:                   req.Make,
		Model:                  req.Model,
		TrimLevel:              req.TrimLevel,
		DownPayment:            req.DownPayment,
		DownPaymentMethod:      downPaymentMethod,
		MonthlyPayment:         req.MonthlyPayment,
		BrokerFee:              req.BrokerFee,
		BrokerPercent:          req.BrokerPercent,
		BrokerCommission:       req.BrokerCommission,
		BrokerCommissionStatus: brokerCommissionStatus,
		YourNet:                req.YourNet,
		BrokerFeeStatus:        brokerFeeStatus,
		DealStatus:             dealStatus,
		CollectAmount:          req.CollectAmount,
		DocFee:                 req.DocFee,
		BrokerPaymentMethod:    brokerPaymentMethod,
		Notes:                  req.Notes,
	})

	if err != nil {
		if errors.Is(err, service.ErrBrokerCommissionMismatch) {
			http.Error(w, "broker_commission does not match broker_percent and broker_fee", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toDealResponse(deal))
}

func (h *DealHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req dealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.ClientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	clientID, err := uuid.Parse(req.ClientID)
	if err != nil {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return
	}

	brokerID, err := nullUUID(req.BrokerID)
	if err != nil {
		http.Error(w, "invalid broker_id", http.StatusBadRequest)
		return
	}

	dealerID, err := nullUUID(req.DealerID)
	if err != nil {
		http.Error(w, "invalid dealer_id", http.StatusBadRequest)
		return
	}

	dealDate, err := nullTime(req.DealDate)
	if err != nil {
		http.Error(w, "invalid deal_date, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	downPaymentMethod, err := nullDownPaymentMethod(req.DownPaymentMethod)
	if err != nil {
		http.Error(w, "invalid down_payment_method", http.StatusBadRequest)
		return
	}
	brokerCommissionStatus, err := nullPaymentStatus(req.BrokerCommissionStatus)
	if err != nil {
		http.Error(w, "invalid broker_commission_status", http.StatusBadRequest)
		return
	}
	brokerFeeStatus, err := nullPaymentStatus(req.BrokerFeeStatus)
	if err != nil {
		http.Error(w, "invalid broker_fee_status", http.StatusBadRequest)
		return
	}
	dealStatus, err := nullDealStatus(req.DealStatus)
	if err != nil {
		http.Error(w, "invalid deal_status", http.StatusBadRequest)
		return
	}
	brokerPaymentMethod, err := nullBrokerPaymentMethod(req.BrokerPaymentMethod)
	if err != nil {
		http.Error(w, "invalid broker_payment_method", http.StatusBadRequest)
		return
	}

	deal, err := h.deal.Update(r.Context(), service.UpdateDealInput{
		ID:                     id,
		ClientID:               clientID,
		BrokerID:               brokerID,
		DealerID:               dealerID,
		DealDate:               dealDate,
		DealNumber:             req.DealNumber,
		Vin:                    req.Vin,
		Make:                   req.Make,
		Model:                  req.Model,
		TrimLevel:              req.TrimLevel,
		DownPayment:            req.DownPayment,
		DownPaymentMethod:      downPaymentMethod,
		MonthlyPayment:         req.MonthlyPayment,
		BrokerFee:              req.BrokerFee,
		BrokerPercent:          req.BrokerPercent,
		BrokerCommission:       req.BrokerCommission,
		BrokerCommissionStatus: brokerCommissionStatus,
		YourNet:                req.YourNet,
		BrokerFeeStatus:        brokerFeeStatus,
		DealStatus:             dealStatus,
		CollectAmount:          req.CollectAmount,
		DocFee:                 req.DocFee,
		BrokerPaymentMethod:    brokerPaymentMethod,
		Notes:                  req.Notes,
	})
	if err != nil {
		if errors.Is(err, service.ErrBrokerCommissionMismatch) {
			http.Error(w, "broker_commission does not match broker_percent and broker_fee", http.StatusBadRequest)
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDealResponse(deal))
}

func (h *DealHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deal, err := h.deal.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDealResponse(deal))
}

func (h *DealHandler) GetShow(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deal, err := h.deal.GetShow(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDealShowResponse(deal))
}

func (h *DealHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	limit, offset := parsePagination(r)

	if clientIDStr := r.URL.Query().Get("client_id"); clientIDStr != "" {
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			http.Error(w, "invalid client id", http.StatusBadRequest)
			return
		}
		deals, err := h.deal.ListByClient(r.Context(), clientID, limit+1, offset)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		deals, hasMore := paginate(deals, limit)
		json.NewEncoder(w).Encode(map[string]any{
			"data":     toDealResponses(deals),
			"limit":    limit,
			"offset":   offset,
			"has_more": hasMore,
		})
		return
	}

	deals, err := h.deal.List(r.Context(), limit+1, offset)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	deals, hasMore := paginate(deals, limit)
	json.NewEncoder(w).Encode(map[string]any{
		"data":     toDealResponses(deals),
		"limit":    limit,
		"offset":   offset,
		"has_more": hasMore,
	})
}

func (h *DealHandler) UpdateBrokerFeeStatus(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "broker fee status is required", http.StatusBadRequest)
		return
	}
	status, err := nullPaymentStatus(&req.Status)
	if err != nil {
		http.Error(w, "invalid broker fee status", http.StatusBadRequest)
		return
	}

	deal, err := h.deal.UpdateBrokerFeeStatus(r.Context(), id, status.PaymentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDealResponse(deal))
}

func (h *DealHandler) UpdateBrokerCommissionStatus(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "broker commission status is required", http.StatusBadRequest)
		return
	}
	status, err := nullPaymentStatus(&req.Status)
	if err != nil {
		http.Error(w, "invalid broker commission status", http.StatusBadRequest)
		return
	}

	deal, err := h.deal.UpdateBrokerCommissionStatus(r.Context(), id, status.PaymentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDealResponse(deal))
}

func (h *DealHandler) NextNumber(w http.ResponseWriter, r *http.Request) {
	num, err := h.deal.PreviewNextDealNumber(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"deal_number": num})
}

func (h *DealHandler) MarkPaymentReminderEmailSent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	deal, err := h.deal.MarkPaymentReminderEmailSent(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toDealResponse(deal))
}
