// localstorage initializes development buckets in the local S3-compatible server.
package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/umbrllacrp/g_motion_crm/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.App.Env != "development" || cfg.AWS.S3Endpoint == "" {
		log.Fatal("localstorage requires APP_ENV=development and S3_ENDPOINT")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWS.Region), awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey, "")))
	if err != nil {
		log.Fatal(err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) { o.BaseEndpoint = aws.String(cfg.AWS.S3Endpoint); o.UsePathStyle = true })
	for _, bucket := range []string{cfg.AWS.BucketDocs, cfg.AWS.BucketBackups} {
		if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err == nil {
			continue
		}
		if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
			log.Fatal(err)
		}
		log.Printf("Created local bucket %s", bucket)
	}
}
