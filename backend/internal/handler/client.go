package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type ClientHandler struct {
	client *service.ClientService
}

func NewClientHandler(client *service.ClientService) *ClientHandler {
	return &ClientHandler{client: client}
}

type clientRequest struct {
	FirstName             string   `json:"first_name"`
	LastName              string   `json:"last_name"`
	Phone                 string   `json:"phone"`
	Email                 string   `json:"email"`
	Status                string   `json:"status"`
	Source                *string  `json:"source"`
	Notes                 *string  `json:"notes"`
	DateOfBirth           *string  `json:"date_of_birth"`
	Address               *string  `json:"address"`
	City                  *string  `json:"city"`
	State                 *string  `json:"state"`
	Zip                   *string  `json:"zip"`
	Ssn                   *string  `json:"ssn"`
	HousingStatus         *string  `json:"housing_status"`
	HousingDuration       *string  `json:"housing_duration"`
	HousingMonthlyPayment *float64 `json:"housing_monthly_payment"`
	EmployerName          *string  `json:"employer_name"`
	EmployerAddress       *string  `json:"employer_address"`
	EmployerCity          *string  `json:"employer_city"`
	EmployerState         *string  `json:"employer_state"`
	EmployerZip           *string  `json:"employer_zip"`
	Occupation            *string  `json:"occupation"`
	EmployerPhone         *string  `json:"employer_phone"`
	EmploymentStatus      *string  `json:"employment_status"`
	EmploymentDuration    *string  `json:"employment_duration"`
	GrossMonthlyIncome    *float64 `json:"gross_monthly_income"`
	OtherMonthlyIncome    *float64 `json:"other_monthly_income"`
	OtherIncomeSource     *string  `json:"other_income_source"`
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req clientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Phone == "" || req.Email == "" {
		http.Error(w, "first_name, last_name, phone and email are required", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = string(db.ClientStatusActive)
	}

	dob, err := nullTime(req.DateOfBirth)
	if err != nil {
		http.Error(w, "invalid date_of_birth, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	source, err := nullClientsSource(req.Source)
	if err != nil {
		http.Error(w, "invalid source", http.StatusBadRequest)
		return
	}
	employmentStatus, err := nullEmploymentStatus(req.EmploymentStatus)
	if err != nil {
		http.Error(w, "invalid employment_status", http.StatusBadRequest)
		return
	}

	client, err := h.client.Create(r.Context(), service.CreateClientInput{
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Phone:                 req.Phone,
		Email:                 req.Email,
		Status:                db.ClientStatus(req.Status),
		Source:                source,
		Notes:                 req.Notes,
		DateOfBirth:           dob,
		Address:               req.Address,
		City:                  req.City,
		State:                 req.State,
		Zip:                   req.Zip,
		Ssn:                   req.Ssn,
		HousingStatus:         req.HousingStatus,
		HousingDuration:       req.HousingDuration,
		HousingMonthlyPayment: req.HousingMonthlyPayment,
		EmployerName:          req.EmployerName,
		EmployerAddress:       req.EmployerAddress,
		EmployerCity:          req.EmployerCity,
		EmployerState:         req.EmployerState,
		EmployerZip:           req.EmployerZip,
		Occupation:            req.Occupation,
		EmployerPhone:         req.EmployerPhone,
		EmploymentStatus:      employmentStatus,
		EmploymentDuration:    req.EmploymentDuration,
		GrossMonthlyIncome:    req.GrossMonthlyIncome,
		OtherMonthlyIncome:    req.OtherMonthlyIncome,
		OtherIncomeSource:     req.OtherIncomeSource,
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			http.Error(w, "client with this phone or email already exists", http.StatusConflict)
			return
		}
		slog.Error("create client", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, toClientResponse(client))
}

func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	client, err := h.client.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		slog.Error("get client", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toClientResponse(client))
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	clients, err := h.client.List(r.Context(), limit+1, offset)
	if err != nil {
		slog.Error("list clients", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	clients, hasMore := paginate(clients, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"data":     toClientResponses(clients),
		"limit":    limit,
		"offset":   offset,
		"has_more": hasMore,
	})
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req clientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.FirstName == "" || req.LastName == "" || req.Phone == "" || req.Email == "" {
		http.Error(w, "first_name, last_name, phone and email are required", http.StatusBadRequest)
		return
	}
	if req.Status == "" {
		req.Status = string(db.ClientStatusActive)
	}

	dob, err := nullTime(req.DateOfBirth)
	if err != nil {
		http.Error(w, "invalid date_of_birth, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	source, err := nullClientsSource(req.Source)
	if err != nil {
		http.Error(w, "invalid source", http.StatusBadRequest)
		return
	}
	employmentStatus, err := nullEmploymentStatus(req.EmploymentStatus)
	if err != nil {
		http.Error(w, "invalid employment_status", http.StatusBadRequest)
		return
	}

	client, err := h.client.Update(r.Context(), service.UpdateClientInput{
		ID:                    id,
		FirstName:             req.FirstName,
		LastName:              req.LastName,
		Phone:                 req.Phone,
		Email:                 req.Email,
		Status:                db.ClientStatus(req.Status),
		Source:                source,
		Notes:                 req.Notes,
		DateOfBirth:           dob,
		Address:               req.Address,
		City:                  req.City,
		State:                 req.State,
		Zip:                   req.Zip,
		Ssn:                   req.Ssn,
		HousingStatus:         req.HousingStatus,
		HousingDuration:       req.HousingDuration,
		HousingMonthlyPayment: req.HousingMonthlyPayment,
		EmployerName:          req.EmployerName,
		EmployerAddress:       req.EmployerAddress,
		EmployerCity:          req.EmployerCity,
		EmployerState:         req.EmployerState,
		EmployerZip:           req.EmployerZip,
		Occupation:            req.Occupation,
		EmployerPhone:         req.EmployerPhone,
		EmploymentStatus:      employmentStatus,
		EmploymentDuration:    req.EmploymentDuration,
		GrossMonthlyIncome:    req.GrossMonthlyIncome,
		OtherMonthlyIncome:    req.OtherMonthlyIncome,
		OtherIncomeSource:     req.OtherIncomeSource,
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			http.Error(w, "client with this phone or email already exists", http.StatusConflict)
			return
		}
		slog.Error("update client", "id", id, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, toClientResponse(client))
}
