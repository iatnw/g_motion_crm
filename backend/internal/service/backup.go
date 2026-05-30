package service

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
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

	reader, writer := io.Pipe()
	counter := &countingWriter{Writer: writer}
	dumpErr := make(chan error, 1)

	go func() {
		gz := gzip.NewWriter(counter)
		cmd := exec.CommandContext(ctx, "pg_dump", "--format=plain", "--no-acl", "--no-owner", s.databaseURL)
		cmd.Stdout = gz

		if err := cmd.Run(); err != nil {
			err = fmt.Errorf("pg_dump: %w", err)
			_ = writer.CloseWithError(err)
			dumpErr <- err
			return
		}
		if err := gz.Close(); err != nil {
			err = fmt.Errorf("gzip close: %w", err)
			_ = writer.CloseWithError(err)
			dumpErr <- err
			return
		}
		_ = writer.Close()
		dumpErr <- nil
	}()

	uploadErr := s.store.Upload(ctx, key, reader)
	if uploadErr != nil {
		_ = reader.CloseWithError(uploadErr)
	}
	if err := <-dumpErr; err != nil {
		return err
	}
	if uploadErr != nil {
		return fmt.Errorf("s3 upload: %w", uploadErr)
	}

	slog.Info("database backup complete", "key", key, "bytes", counter.Count)
	return nil
}

type countingWriter struct {
	io.Writer
	Count int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.Count += int64(n)
	return n, err
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
