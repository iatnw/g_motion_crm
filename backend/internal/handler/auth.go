package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Password string `json:"password"`
}
type loginResponse struct {
	Token        string `json:"token"`
	TOTPRequired bool   `json:"totp_required"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	result, err := h.auth.Login(r.Context(), req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{
		Token:        result.Token,
		TOTPRequired: result.TOTPRequired,
	})
}

type verifyTOTPRequest struct {
	Code string `json:"code"`
}
type verifyTOTPResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) VerifyTOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	token, err := h.auth.VerifyTOTP(r.Context(), req.Code)
	if err != nil {
		if errors.Is(err, service.ErrInvalidTOTP) {
			http.Error(w, "invalid totp code", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(verifyTOTPResponse{Token: token})
}

type setupTOTPResponse struct {
	Secret string `json:"secret"`
	KeyURI string `json:"key_uri"`
}

func (h *AuthHandler) SetUpTOTP(w http.ResponseWriter, r *http.Request) {
	secret, keyURI, err := h.auth.SetUpTOTP(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrTOTPAlreadyEnabled) {
			http.Error(w, "totp already enabled", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(setupTOTPResponse{
		Secret: secret,
		KeyURI: keyURI,
	})
}

type enableTOTPRequest struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

func (h *AuthHandler) EnableTOTP(w http.ResponseWriter, r *http.Request) {
	var req enableTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.auth.EnableTOTP(r.Context(), req.Secret, req.Code); err != nil {
		if errors.Is(err, service.ErrInvalidTOTP) {
			http.Error(w, "invalid totp code", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.DisableTOTP(r.Context()); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type setPasswordRequest struct {
	Password string `json:"password"`
}

func (h *AuthHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	var req setPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.auth.SetPassword(r.Context(), req.Password); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
