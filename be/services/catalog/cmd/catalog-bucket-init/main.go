// catalog-bucket-init creates the local development bucket; it is not a production migration.
package main

import (
	"context"
	"log"
	"os"
	"time"

	s3store "github.com/kainguyen/badminton-hub/be/services/catalog/internal/storage/objectstore/s3"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	region := os.Getenv("CATALOG_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	store, err := s3store.New(ctx, s3store.Config{
		Bucket: os.Getenv("CATALOG_S3_BUCKET"), Region: region, Endpoint: os.Getenv("CATALOG_S3_ENDPOINT"),
		AccessKey: os.Getenv("CATALOG_S3_ACCESS_KEY"), SecretKey: os.Getenv("CATALOG_S3_SECRET_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	for {
		attemptCtx, stop := context.WithTimeout(ctx, 3*time.Second)
		err := store.EnsureBucket(attemptCtx)
		stop()
		if err == nil {
			log.Print("local Catalog image bucket is ready")
			return
		}
		select {
		case <-ctx.Done():
			log.Fatalf("initialize local image bucket: %v", err)
		case <-time.After(2 * time.Second):
		}
	}
}
