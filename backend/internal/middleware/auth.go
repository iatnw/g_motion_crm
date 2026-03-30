package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type contextKey string

const claimsKey contextKey = "claims"

func Auth(parseToken func(string) (*service.Claims, error), requiredStage string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")

			claims, err := parseToken(tokenStr)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if claims.Stage != requiredStage {
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
