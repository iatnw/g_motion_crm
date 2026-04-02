package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	sqlDB := stdlib.OpenDBFromPool(pool)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	queries := db.New(sqlDB)

	authService := service.NewAuth(queries, cfg.Auth)
	brokerService := service.NewBroker(queries)
	dealService := service.NewDeal(queries)
	clientService := service.NewClient(queries)

	authHandler := handler.NewAuthHandler(authService)
	brokerHandler := handler.NewBrokerHandler(brokerService)
	dealHandler := handler.NewDealHandler(dealService)
	clientHandler := handler.NewClientHandler(clientService)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			log.Printf("failed to write health response: %v", err)
		}
	})

	r.Post("/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService.ParseToken, "totp_pending"))
		r.Post("/auth/totp/verify", authHandler.VerifyTOTP)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService.ParseToken, "authenticated"))

		r.Post("/auth/totp/setup", authHandler.SetUpTOTP)
		r.Post("/auth/totp/enable", authHandler.EnableTOTP)
		r.Post("/auth/totp/disable", authHandler.DisableTOTP)
		r.Post("/auth/password", authHandler.SetPassword)

		r.Get("/clients", clientHandler.List)
		r.Post("/clients", clientHandler.Create)
		r.Get("/clients/{id}", clientHandler.Get)
		r.Put("/clients/{id}", clientHandler.Update)

		r.Get("/brokers", brokerHandler.List)
		r.Post("/brokers", brokerHandler.Create)
		r.Get("/brokers/{id}", brokerHandler.Get)
		r.Put("/brokers/{id}", brokerHandler.Update)

		r.Get("/deals", dealHandler.List)
		r.Post("/deals", dealHandler.Create)
		r.Get("/deals/{id}", dealHandler.Get)
		r.Put("/deals/{id}", dealHandler.Update)
		r.Patch("/deals/{id}/broker-fee-status", dealHandler.UpdateBrokerFeeStatus)
		r.Patch("/deals/{id}/broker-commission-status", dealHandler.UpdateBrokerCommissionStatus)
		r.Patch("/deals/{id}/referral-status", dealHandler.UpdateReferralStatus)
	})

	addr := fmt.Sprintf(":%s", cfg.App.Port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("sever failed %v", err)
	}
}
