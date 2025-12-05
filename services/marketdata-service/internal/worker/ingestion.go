package worker

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// IngestionWorker handles the ingestion of market data from files.
type IngestionWorker struct {
	repo        domain.MarketDataRepository
	minioClient *minio.Client
	bucketName  string
}

// NewIngestionWorker creates a new ingestion worker.
func NewIngestionWorker(repo domain.MarketDataRepository) (*IngestionWorker, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	if endpoint == "" {
		endpoint = "localhost:9000"
	}
	if bucketName == "" {
		bucketName = "market-data"
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &IngestionWorker{
		repo:        repo,
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}

// Start starts the ingestion worker.
func (w *IngestionWorker) Start(ctx context.Context) {
	go func() {
		log.Println("Starting ingestion worker...")

		// Check if bucket exists, create if not
		exists, err := w.minioClient.BucketExists(ctx, w.bucketName)
		if err != nil {
			log.Printf("Error checking bucket existence: %v", err)
			return
		}
		if !exists {
			err = w.minioClient.MakeBucket(ctx, w.bucketName, minio.MakeBucketOptions{})
			if err != nil {
				log.Printf("Error creating bucket: %v", err)
				return
			}
			log.Printf("Created bucket: %s", w.bucketName)
		}

		// Listen for bucket notifications
		log.Printf("Listening for events on bucket: %s", w.bucketName)
		for notificationInfo := range w.minioClient.ListenBucketNotification(ctx, w.bucketName, "", "asset.csv", []string{
			"s3:ObjectCreated:Put",
		}) {
			if notificationInfo.Err != nil {
				log.Printf("Error in event channel: %v", notificationInfo.Err)
				continue
			}

			for _, record := range notificationInfo.Records {
				log.Printf("Received event: %s for object: %s", record.EventName, record.S3.Object.Key)
				if err := w.processFile(ctx, record.S3.Object.Key); err != nil {
					log.Printf("Error processing file %s: %v", record.S3.Object.Key, err)
				}
			}
		}
	}()
}

func (w *IngestionWorker) processFile(ctx context.Context, objectKey string) error {
	log.Printf("Starting ingestion for object: %s", objectKey)

	_, err := w.minioClient.StatObject(ctx, w.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object %s/%s: %w", w.bucketName, objectKey, err)
	}

	object, err := w.minioClient.GetObject(ctx, w.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer func() {
		if err := object.Close(); err != nil {
			log.Printf("Ingestion Worker: Error closing object: %v", err)
		}
	}()

	reader := csv.NewReader(object)
	_, err = reader.Read() // Skip header
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	var rows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read csv record: %w", err)
		}
		rows = append(rows, record)
	}

	if len(rows) > 0 {
		return w.batchInsert(ctx, rows)
	}

	log.Println("No rows to insert.")
	return nil
}

func (w *IngestionWorker) batchInsert(ctx context.Context, rows [][]string) error {
	// Convert rows to domain.Asset
	var assets []*domain.Asset
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		assets = append(assets, &domain.Asset{
			Symbol:   row[0],
			Name:     row[1],
			Type:     row[2],
			Exchange: row[3],
			Currency: row[4],
		})
	}

	batchSize := 1000
	for i := 0; i < len(assets); i += batchSize {
		end := i + batchSize
		if end > len(assets) {
			end = len(assets)
		}
		batch := assets[i:end]
		if err := w.repo.UpsertAssets(batch); err != nil {
			return err
		}
	}

	log.Printf("Successfully ingested %d assets.", len(assets))
	return nil
}
