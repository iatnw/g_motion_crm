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
	authHandler := handler.NewAuthHandler(authService)

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/totp/verify", authHandler.VerifyTOTP)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService.ParseToken))
		r.Post("/auth/totp/setup", authHandler.SetUpTOTP)
		r.Post("/auth/totp/enable", authHandler.EnableTOTP)
		r.Post("/auth/totp/disable", authHandler.DisableTOTP)
		r.Post("/auth/password", authHandler.SetPassword)
	})

	addr := fmt.Sprintf(":%s", cfg.App.Port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("sever failed %v", err)
	}
}
