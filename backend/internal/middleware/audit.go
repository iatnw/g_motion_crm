package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
)

type auditLogCreator interface {
	CreateAuditLog(context.Context, db.CreateAuditLogParams) (db.AuditLog, error)
}

// AuditMutations records successful authenticated writes without persisting request bodies.
func AuditMutations(q auditLogCreator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isMutation(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)
			if recorder.status < http.StatusOK || recorder.status >= http.StatusMultipleChoices {
				return
			}

			route := chi.RouteContext(r.Context()).RoutePattern()
			metadata, err := json.Marshal(map[string]any{
				"route":       route,
				"status_code": recorder.status,
			})
			if err != nil {
				slog.Error("marshal audit metadata", "error", err)
				return
			}

			if _, err := q.CreateAuditLog(r.Context(), db.CreateAuditLogParams{
				EntityType:    entityType(route),
				EntityID:      routeEntityID(r),
				Actor:         auditActor(r.Context()),
				Action:        strings.ToLower(r.Method),
				ChangedFields: metadata,
				IpAddress:     textValue(remoteIP(r.RemoteAddr)),
				UserAgent:     textValue(r.UserAgent()),
			}); err != nil {
				slog.Error("create audit log", "method", r.Method, "route", route, "error", err)
			}
		})
	}
}

func auditActor(ctx context.Context) string {
	claims := GetClaims(ctx)
	if claims == nil || claims.Subject == "" {
		return "owner"
	}
	return claims.Subject
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func isMutation(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func entityType(route string) string {
	route = strings.TrimPrefix(route, "/")
	if segment, _, ok := strings.Cut(route, "/"); ok {
		return segment
	}
	return route
}

func routeEntityID(r *http.Request) pgtype.UUID {
	for _, name := range []string{"id", "dealID"} {
		if id, err := uuid.Parse(chi.URLParam(r, name)); err == nil {
			return pgtype.UUID{Bytes: id, Valid: true}
		}
	}
	return pgtype.UUID{}
}

func textValue(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}
