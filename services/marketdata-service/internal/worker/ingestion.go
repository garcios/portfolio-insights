package worker

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type IngestionWorker struct {
	repo        domain.MarketDataRepository
	minioClient *minio.Client
	bucketName  string
	fileName    string
}

func NewIngestionWorker(repo domain.MarketDataRepository) (*IngestionWorker, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	fileName := "asset.csv"

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
		fileName:    fileName,
	}, nil
}

func (w *IngestionWorker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		if err := w.processFile(ctx); err != nil {
			log.Printf("Error processing file: %v", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.processFile(ctx); err != nil {
					log.Printf("Error processing file: %v", err)
				}
			}
		}
	}()
}

func (w *IngestionWorker) processFile(ctx context.Context) error {
	log.Println("Starting asset ingestion...")

	_, err := w.minioClient.StatObject(ctx, w.bucketName, w.fileName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object %s/%s: %w", w.bucketName, w.fileName, err)
	}

	object, err := w.minioClient.GetObject(ctx, w.bucketName, w.fileName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

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
