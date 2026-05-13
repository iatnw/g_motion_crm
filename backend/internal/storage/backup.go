package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
)

type BackupStorage struct {
	client *s3.Client
	bucket string
}

func NewBackupStorage(cfg config.AWSConfig) (*BackupStorage, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}
	return &BackupStorage{
		client: s3.NewFromConfig(awsCfg),
		bucket: cfg.BucketBackups,
	}, nil
}

func (b *BackupStorage) Upload(ctx context.Context, key string, body io.Reader, size int64) error {
	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(b.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/gzip"),
	})
	return err
}
