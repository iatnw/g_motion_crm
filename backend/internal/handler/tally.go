package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type TallyHandler struct {
	tally         *service.TallyService
	signingSecret string
}

func NewTallyHandler(tally *service.TallyService, signingSecret string) *TallyHandler {
	return &TallyHandler{tally: tally, signingSecret: signingSecret}
}

// Webhook receives Tally form submissions.
// POST /webhooks/tally
func (h *TallyHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 5<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Signing secret must be configured — reject all requests without it.
	if h.signingSecret == "" {
		http.Error(w, "webhook not configured", http.StatusServiceUnavailable)
		return
	}
	sig := r.Header.Get("tally-signature")
	mac := hmac.New(sha256.New, []byte(h.signingSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	slog.Info("tally signature debug", "received", sig, "expected", expected)
	if !verifyTallySignature(body, sig, h.signingSecret) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload service.TallyWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if payload.EventType != "FORM_RESPONSE" {
		// Acknowledge other event types without processing.
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.tally.ProcessSubmission(r.Context(), payload); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func verifyTallySignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
