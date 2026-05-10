package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type AuthHandler struct {
	auth   *service.AuthService
	secure bool // true in production — sets Secure flag on cookies
}

func NewAuthHandler(auth *service.AuthService, secure bool) *AuthHandler {
	return &AuthHandler{auth: auth, secure: secure}
}

func (h *AuthHandler) setTokenCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

type loginRequest struct {
	Password string `json:"password"`
}
type loginResponse struct {
	TOTPRequired bool `json:"totp_required"`
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
		slog.Error("login", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// totp_pending tokens live 5 minutes, authenticated tokens 24 hours.
	maxAge := 24 * 60 * 60
	if result.TOTPRequired {
		maxAge = 5 * 60
	}
	h.setTokenCookie(w, result.Token, maxAge)
	writeJSON(w, http.StatusOK, loginResponse{TOTPRequired: result.TOTPRequired})
}

type verifyTOTPRequest struct {
	Code string `json:"code"`
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
			http.Error(w, "invalid TOTP code", http.StatusUnauthorized)
			return
		}
		slog.Error("verify totp", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.setTokenCookie(w, token, 24*60*60)
	w.WriteHeader(http.StatusNoContent)
}

type setupTOTPResponse struct {
	Secret string `json:"secret"`
	KeyURI string `json:"key_uri"`
}

func (h *AuthHandler) SetUpTOTP(w http.ResponseWriter, r *http.Request) {
	secret, keyURI, err := h.auth.SetUpTOTP(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrTOTPAlreadyEnabled) {
			http.Error(w, "TOTP already enabled", http.StatusConflict)
			return
		}
		slog.Error("setup totp", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, setupTOTPResponse{Secret: secret, KeyURI: keyURI})
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
			http.Error(w, "invalid TOTP code", http.StatusUnauthorized)
			return
		}
		slog.Error("enable totp", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) SetupRequired(w http.ResponseWriter, r *http.Request) {
	required, err := h.auth.SetupRequired(r.Context())
	if err != nil {
		slog.Error("setup required", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"required": required})
}

func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.auth.Setup(r.Context(), req.Password); err != nil {
		if errors.Is(err, service.ErrSetupAlreadyDone) {
			http.Error(w, "setup already completed", http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrWeakPassword) {
			http.Error(w, "password must be at least 12 characters", http.StatusBadRequest)
			return
		}
		slog.Error("setup", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.setTokenCookie(w, "", -1)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.DisableTOTP(r.Context()); err != nil {
		slog.Error("disable totp", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type setPasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	Password        string `json:"password"`
}

func (h *AuthHandler) SetPassword(w http.ResponseWriter, r *http.Request) {
	var req setPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.auth.SetPassword(r.Context(), req.CurrentPassword, req.Password); err != nil {
		if errors.Is(err, service.ErrWeakPassword) {
			http.Error(w, "password must be at least 12 characters", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrInvalidCurrentPassword) {
			http.Error(w, "invalid current password", http.StatusUnauthorized)
			return
		}
		slog.Error("set password", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
