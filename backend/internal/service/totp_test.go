package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
	"github.com/umbrllacrp/g_motion_crm/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type totpDB struct {
	db.DBTX
	hash     string
	secret   pgtype.Text
	readErr  error
	writeErr error
	rows     int64
	writes   int
}

type totpRow struct{ d *totpDB }

func (r totpRow) Scan(dest ...any) error {
	if r.d.readErr != nil {
		return r.d.readErr
	}
	*dest[1].(*string) = r.d.hash
	*dest[2].(*pgtype.Text) = r.d.secret
	return nil
}
func (d *totpDB) QueryRow(context.Context, string, ...any) pgx.Row { return totpRow{d} }
func (d *totpDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	d.writes++
	if d.rows == 0 {
		return pgconn.NewCommandTag("UPDATE 0"), d.writeErr
	}
	return pgconn.NewCommandTag("UPDATE 1"), d.writeErr
}

func TestDisableTOTPRequiresReauthentication(t *testing.T) {
	const password = "test-password-12345"
	const secret = "JBSWY3DPEHPK3PXP"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name, password, code string
		disabled             bool
		readErr, writeErr    error
		rows                 int64
		want                 error
		writes               int
	}{
		{name: "valid credentials", password: password, code: code, rows: 1, writes: 1},
		{name: "wrong password", password: "wrong", code: code, want: ErrInvalidCredentials},
		{name: "wrong code", password: password, code: "invalid", want: ErrInvalidCredentials},
		{name: "missing password", code: code, want: ErrInvalidCredentials},
		{name: "missing code", password: password, want: ErrInvalidCredentials},
		{name: "not enabled", password: password, code: code, disabled: true, want: ErrTOTPNotEnabled},
		{name: "read failure", readErr: dbErr, want: dbErr},
		{name: "write failure", password: password, code: code, rows: 1, writes: 1, writeErr: dbErr, want: dbErr},
		{name: "settings changed during validation", password: password, code: code, writes: 1, want: ErrInvalidCredentials},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &totpDB{hash: string(hash), secret: pgtype.Text{String: secret, Valid: !tc.disabled}, readErr: tc.readErr, writeErr: tc.writeErr, rows: tc.rows}
			svc := NewAuth(db.New(store), config.AuthConfig{})
			err := svc.DisableTOTP(context.Background(), tc.password, tc.code)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if store.writes != tc.writes {
				t.Fatalf("writes = %d, want %d", store.writes, tc.writes)
			}
		})
	}
}

func TestEnableTOTPCannotReplaceExistingSecret(t *testing.T) {
	const secret = "JBSWY3DPEHPK3PXP"
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	store := &totpDB{rows: 0}
	svc := NewAuth(db.New(store), config.AuthConfig{})
	if err := svc.EnableTOTP(context.Background(), secret, code); !errors.Is(err, ErrTOTPAlreadyEnabled) {
		t.Fatalf("got %v, want already enabled", err)
	}
}
