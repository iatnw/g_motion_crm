package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"github.com/umbrllacrp/g_motion_crm/internal/handler"
	"github.com/umbrllacrp/g_motion_crm/internal/middleware"
	"github.com/umbrllacrp/g_motion_crm/internal/service"
	"github.com/umbrllacrp/g_motion_crm/internal/storage"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)

	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("failed to set goose dialect", "error", err)
		os.Exit(1)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	queries := db.New(sqlDB)

	encryptionKey, err := hex.DecodeString(cfg.App.EncryptionKey)
	if err != nil {
		slog.Error("invalid ENCRYPTION_KEY: must be a hex-encoded 32-byte key", "error", err)
		os.Exit(1)
	}
	if len(encryptionKey) != 32 {
		slog.Error("invalid ENCRYPTION_KEY: must decode to exactly 32 bytes", "got", len(encryptionKey))
		os.Exit(1)
	}

	s3Client, err := storage.NewS3Client(cfg.AWS)
	if err != nil {
		slog.Error("failed to init S3 client", "error", err)
		os.Exit(1)
	}

	authService := service.NewAuth(queries, cfg.Auth)
	brokerService := service.NewBroker(queries)
	dealService := service.NewDeal(queries)
	clientService := service.NewClient(queries, encryptionKey)
	dealerService := service.NewDealer(queries)
	documentService := service.NewDocument(queries, s3Client)
	tallyService := service.NewTally(clientService, documentService, dealService, queries, sqlDB)

	authHandler := handler.NewAuthHandler(authService)
	brokerHandler := handler.NewBrokerHandler(brokerService)
	dealHandler := handler.NewDealHandler(dealService)
	clientHandler := handler.NewClientHandler(clientService)
	dealerHandler := handler.NewDealerHandler(dealerService)
	documentHandler := handler.NewDocumentHandler(documentService)
	tallyHandler := handler.NewTallyHandler(tallyService, cfg.Tally.SigningSecret)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	frontendURL := cfg.App.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	r.Use(corsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			slog.Error("failed to write health response", "error", err)
		}
	})

	// Public webhook — Tally sends to this without auth token
	r.Post("/webhooks/tally", tallyHandler.Webhook)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, time.Minute))
		r.Post("/auth/login", authHandler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, time.Minute))
		r.Use(middleware.Auth(authService.ParseToken, "totp_pending"))
		r.Post("/auth/totp/verify", authHandler.VerifyTOTP)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService.ParseToken, "authenticated"))

		r.Post("/auth/totp/setup", authHandler.SetUpTOTP)
		r.Post("/auth/totp/enable", authHandler.EnableTOTP)
		r.Post("/auth/totp/disable", authHandler.DisableTOTP)
		r.Post("/auth/password", authHandler.SetPassword)

		// Clients
		r.Get("/clients", clientHandler.List)
		r.Post("/clients", clientHandler.Create)
		r.Get("/clients/{id}", clientHandler.Get)
		r.Put("/clients/{id}", clientHandler.Update)
		r.Get("/clients/{id}/documents", documentHandler.ListByClient)
		r.Post("/clients/{id}/documents", documentHandler.UploadForClient)

		// Brokers
		r.Get("/brokers", brokerHandler.List)
		r.Post("/brokers", brokerHandler.Create)
		r.Get("/brokers/{id}", brokerHandler.Get)
		r.Put("/brokers/{id}", brokerHandler.Update)

		// Dealers
		r.Get("/dealers", dealerHandler.List)
		r.Post("/dealers", dealerHandler.Create)
		r.Get("/dealers/{id}", dealerHandler.Get)
		r.Put("/dealers/{id}", dealerHandler.Update)

		// Deals
		r.Get("/deals", dealHandler.List)
		r.Post("/deals", dealHandler.Create)
		r.Get("/deals/{id}", dealHandler.Get)
		r.Put("/deals/{id}", dealHandler.Update)
		r.Patch("/deals/{id}/broker-fee-status", dealHandler.UpdateBrokerFeeStatus)
		r.Patch("/deals/{id}/broker-commission-status", dealHandler.UpdateBrokerCommissionStatus)
		r.Patch("/deals/{id}/referral-status", dealHandler.UpdateReferralStatus)
		r.Get("/deals/{id}/documents", documentHandler.ListByDeal)
		r.Post("/deals/{id}/documents", documentHandler.UploadForDeal)

		// Documents
		r.Get("/documents/{id}/url", documentHandler.GetURL)
		r.Delete("/documents/{id}", documentHandler.Delete)
	})

	addr := fmt.Sprintf(":%s", cfg.App.Port)
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
