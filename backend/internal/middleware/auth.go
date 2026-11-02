package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type contextKey string

const claimsKey contextKey = "claims"

func Auth(parseToken func(string) (*service.Claims, error), sessionValid func(context.Context, *service.Claims) (bool, error), requiredStage string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := parseToken(cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Stage != requiredStage {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			valid, err := sessionValid(r.Context(), claims)
			if err != nil {
				slog.Error("validate session", "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) *service.Claims {
	c, _ := ctx.Value(claimsKey).(*service.Claims)
	return c
}
