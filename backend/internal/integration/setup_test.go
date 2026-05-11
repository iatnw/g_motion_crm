package integration_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/handler"
	"github.com/umbrllacrp/g_motion_crm/internal/middleware"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
	"github.com/umbrllacrp/g_motion_crm/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

const testPassword = "testpassword12345"
const tallySigningSecret = "test-tally-secret"

var (
	srv  *httptest.Server
	pool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	mustSetenv("APP_ENV", "test")
	mustSetenv("APP_PORT", "8080")
	mustSetenv("APP_BASE_URL", "http://localhost:8080")
	mustSetenv("FRONTEND_URL", "http://localhost:3000")
	mustSetenv("DATABASE_URL", envOr("TEST_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gmotion_test?sslmode=disable"))
	mustSetenv("JWT_SECRET", "test-jwt-secret-for-integration-tests-only")
	mustSetenv("ENCRYPTION_KEY", "0000000000000000000000000000000000000000000000000000000000000000")
	mustSetenv("AWS_REGION", "us-east-1")
	mustSetenv("AWS_ACCESS_KEY_ID", "test")
	mustSetenv("AWS_SECRET_ACCESS_KEY", "test")
	mustSetenv("S3_BUCKET_DOCUMENTS", "test-docs")
	mustSetenv("S3_BUCKET_BACKUPS", "test-backups")
	mustSetenv("TALLY_SIGNING_SECRET", tallySigningSecret)

	cfg, err := config.Load()
	if err != nil {
		panic("load config: " + err.Error())
	}

	pool, err = pgxpool.New(context.Background(), cfg.Database.DatabaseURL)
	if err != nil {
		panic("connect to DB: " + err.Error())
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		panic("run migrations: " + err.Error())
	}

	truncateData()

	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	if _, err := pool.Exec(context.Background(),
		`UPDATE settings SET owner_password_hash = $1, owner_totp_secret = NULL`,
		string(hash),
	); err != nil {
		panic("seed settings: " + err.Error())
	}

	encryptionKey, _ := hex.DecodeString(cfg.App.EncryptionKey)
	s3Client, _ := storage.NewS3Client(cfg.AWS)
	queries := db.New(pool)

	authSvc := service.NewAuth(queries, cfg.Auth)
	brokerSvc := service.NewBroker(queries)
	dealSvc := service.NewDeal(queries, pool)
	clientSvc := service.NewClient(queries, encryptionKey)
	dealerSvc := service.NewDealer(queries)
	documentSvc := service.NewDocument(queries, s3Client)
	dashboardSvc := service.NewDashboard(queries)
	tallySvc := service.NewTally(clientSvc, documentSvc, dealSvc, queries, pool)
	referralSvc := service.NewReferral(queries)

	authH := handler.NewAuthHandler(authSvc, false)
	brokerH := handler.NewBrokerHandler(brokerSvc)
	dealH := handler.NewDealHandler(dealSvc)
	clientH := handler.NewClientHandler(clientSvc)
	dealerH := handler.NewDealerHandler(dealerSvc)
	documentH := handler.NewDocumentHandler(documentSvc)
	dashboardH := handler.NewDashboardHandler(dashboardSvc)
	tallyH := handler.NewTallyHandler(tallySvc, cfg.Tally.SigningSecret)
	referralH := handler.NewReferralHandler(referralSvc)

	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)

	r.Post("/webhooks/tally", tallyH.Webhook)

	r.Get("/auth/setup-required", authH.SetupRequired)
	r.Post("/auth/setup", authH.Setup)

	r.Post("/auth/login", authH.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc.ParseToken, "totp_pending"))
		r.Post("/auth/totp/verify", authH.VerifyTOTP)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc.ParseToken, "authenticated"))

		r.Get("/auth/me", authH.Me)
		r.Post("/auth/logout", authH.Logout)
		r.Post("/auth/password", authH.SetPassword)

		r.Get("/clients", clientH.List)
		r.Post("/clients", clientH.Create)
		r.Get("/clients/{id}", clientH.Get)
		r.Put("/clients/{id}", clientH.Update)

		r.Get("/brokers", brokerH.List)
		r.Post("/brokers", brokerH.Create)
		r.Get("/brokers/{id}", brokerH.Get)
		r.Put("/brokers/{id}", brokerH.Update)
		r.Delete("/brokers/{id}", brokerH.Delete)

		r.Get("/dealers", dealerH.List)
		r.Post("/dealers", dealerH.Create)
		r.Get("/dealers/{id}", dealerH.Get)
		r.Delete("/dealers/{id}", dealerH.Delete)

		r.Get("/deals", dealH.List)
		r.Post("/deals", dealH.Create)
		r.Get("/deals/{id}", dealH.Get)
		r.Put("/deals/{id}", dealH.Update)
		r.Patch("/deals/{id}/status", dealH.UpdateStatus)

		r.Get("/referrals", referralH.List)
		r.Post("/deals/{dealID}/referrals", referralH.Create)
		r.Get("/deals/{dealID}/referrals", referralH.ListByDeal)
		r.Delete("/referrals/{id}", referralH.Delete)

		r.Get("/dashboard/summary", dashboardH.Summary)
		r.Get("/documents/{id}/url", documentH.GetURL)
	})

	srv = httptest.NewServer(r)
}

func teardown() {
	srv.Close()
	pool.Close()
}

func truncateData() {
	_, err := pool.Exec(context.Background(), `
		TRUNCATE TABLE referrals, documents, tally_webhook_events, deals, clients, brokers, dealers
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		panic("truncate: " + err.Error())
	}
}

// login performs a login and returns the session cookie.
func login(t *testing.T) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"password": testPassword})
	res, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login: got %d, want 200", res.StatusCode)
	}
	for _, c := range res.Cookies() {
		if c.Name == "token" {
			return c
		}
	}
	t.Fatal("login: no token cookie in response")
	return nil
}

// do sends an authenticated request and returns the response.
func do(t *testing.T, method, path string, body any, cookie *http.Cookie) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, srv.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return res
}

// decodeJSON decodes the response body into dst.
func decodeJSON(t *testing.T, res *http.Response, dst any) {
	t.Helper()
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// tallySignature computes the HMAC-SHA256 base64 signature for a Tally webhook body.
func tallySignature(body []byte) string {
	mac := hmac.New(sha256.New, []byte(tallySigningSecret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func mustSetenv(key, val string) {
	if err := os.Setenv(key, val); err != nil {
		panic(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
