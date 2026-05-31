package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
)

type auditRecorder struct {
	logs []db.CreateAuditLogParams
}

func (r *auditRecorder) CreateAuditLog(_ context.Context, arg db.CreateAuditLogParams) (db.AuditLog, error) {
	r.logs = append(r.logs, arg)
	return db.AuditLog{}, nil
}

func TestAuditMutationsRecordsSuccessfulWrite(t *testing.T) {
	store := &auditRecorder{}
	r := chi.NewRouter()
	r.Use(AuditMutations(store))
	r.Put("/clients/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/clients/1b2a177f-4060-452a-9be3-b1bdb68d1119", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	req.Header.Set("User-Agent", "audit-test")
	r.ServeHTTP(httptest.NewRecorder(), req)

	if len(store.logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(store.logs))
	}
	log := store.logs[0]
	if log.EntityType != "clients" || log.Action != "put" {
		t.Fatalf("audit log = %#v, want clients put", log)
	}
	if log.Actor != "owner" {
		t.Fatalf("actor = %q, want owner fallback", log.Actor)
	}
	if !log.EntityID.Valid {
		t.Fatal("entity id is null, want route id")
	}
	if log.IpAddress.String != "192.0.2.10" || log.UserAgent.String != "audit-test" {
		t.Fatalf("request metadata = %#v, want remote IP and user agent", log)
	}
}

func TestAuditActorUsesClaimsSubject(t *testing.T) {
	ctx := context.WithValue(context.Background(), claimsKey, &service.Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "account-123"},
	})

	if got := auditActor(ctx); got != "account-123" {
		t.Fatalf("auditActor = %q, want account-123", got)
	}
}

func TestAuditEntityIDFoundByValue(t *testing.T) {
	store := &auditRecorder{}
	r := chi.NewRouter()
	r.Use(AuditMutations(store))
	r.Delete("/deals/{dealID}/documents/{docID}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(
		http.MethodDelete,
		"/deals/1b2a177f-4060-452a-9be3-b1bdb68d1119/documents/2c3b288g-5171-563b-ace4-c2cec79e2230",
		nil,
	))

	if len(store.logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(store.logs))
	}
	if !store.logs[0].EntityID.Valid {
		t.Fatal("entity id is null, want first UUID param regardless of name")
	}
}

func TestAuditMutationsSkipsFailedWriteAndRead(t *testing.T) {
	store := &auditRecorder{}
	r := chi.NewRouter()
	r.Use(AuditMutations(store))
	r.Post("/clients", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})
	r.Get("/clients", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/clients", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/clients", nil))

	if len(store.logs) != 0 {
		t.Fatalf("audit logs = %d, want 0", len(store.logs))
	}
}
