package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

func TestAuthSessionValidation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stage string
		valid bool
		err   error
		want  int
	}{
		{"current authenticated session", "authenticated", true, nil, 204},
		{"revoked authenticated session", "authenticated", false, nil, 401},
		{"current pending session", "totp_pending", true, nil, 204},
		{"revoked pending session", "totp_pending", false, nil, 401},
		{"database unavailable", "authenticated", false, errors.New("database unavailable"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := &service.Claims{Stage: tc.stage, SessionVersion: 1}
			handler := Auth(func(string) (*service.Claims, error) { return claims, nil },
				func(context.Context, *service.Claims) (bool, error) { return tc.valid, tc.err }, tc.stage)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if GetClaims(r.Context()) != claims {
						t.Error("missing claims")
					}
					w.WriteHeader(http.StatusNoContent)
				}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{Name: "token", Value: "test-token"})
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != tc.want {
				t.Fatalf("got %d, want %d", res.Code, tc.want)
			}
		})
	}
}
