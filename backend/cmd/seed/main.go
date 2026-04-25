// seed sets the initial owner password.
// Usage: go run ./cmd/seed <password>
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/seed <password>")
		os.Exit(1)
	}
	password := os.Args[1]

	if os.Getenv("DATABASE_URL") == "" {
		_ = godotenv.Load()
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`UPDATE settings SET owner_password_hash = $1, updated_at = NOW() WHERE id = 1`, string(hash))
	if err != nil {
		log.Fatalf("update: %v", err)
	}

	fmt.Println("Password set successfully.")
}
