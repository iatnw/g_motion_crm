package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/umbrllacrp/g_motion_crm/internal/storage"
)

type BackupService struct {
	store       *storage.BackupStorage
	databaseURL string
}

func NewBackup(store *storage.BackupStorage, databaseURL string) *BackupService {
	return &BackupService{store: store, databaseURL: databaseURL}
}

func (s *BackupService) Run(ctx context.Context) error {
	key := fmt.Sprintf("backups/%s.sql.gz", time.Now().UTC().Format("2006-01-02T15-04-05Z"))

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	cmd := exec.CommandContext(ctx, "pg_dump", "--format=plain", "--no-acl", "--no-owner", s.databaseURL)
	cmd.Stdout = gz

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("gzip close: %w", err)
	}

	backupSize := int64(buf.Len())
	if err := s.store.Upload(ctx, key, &buf, backupSize); err != nil {
		return fmt.Errorf("s3 upload: %w", err)
	}

	slog.Info("database backup complete", "key", key, "bytes", backupSize)
	return nil
}

// StartScheduler runs an initial backup then one every 24 hours.
// It stops when ctx is cancelled.
func (s *BackupService) StartScheduler(ctx context.Context) {
	go func() {
		if err := s.Run(ctx); err != nil {
			slog.Error("initial backup failed", "error", err)
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.Run(ctx); err != nil {
					slog.Error("scheduled backup failed", "error", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
