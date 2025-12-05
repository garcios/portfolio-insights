// Package worker implements background workers.
package worker

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/domain"
	"github.com/garcios/portfolio-insights/services/marketdata-service/internal/metrics"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// CurrencyIngestionWorker handles the ingestion of currency rates from files.
type CurrencyIngestionWorker struct {
	repo        domain.MarketDataRepository
	minioClient *minio.Client
	bucketName  string
}

// NewCurrencyIngestionWorker creates a new currency ingestion worker.
func NewCurrencyIngestionWorker(repo domain.MarketDataRepository) (*CurrencyIngestionWorker, error) {
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

	return &CurrencyIngestionWorker{
		repo:        repo,
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}

// Start starts the currency ingestion worker.
func (w *CurrencyIngestionWorker) Start(ctx context.Context) {
	go func() {
		log.Println("Currency Worker: Starting ingestion worker...")

		// Check if bucket exists, create if not
		exists, err := w.minioClient.BucketExists(ctx, w.bucketName)
		if err != nil {
			log.Printf("Currency Worker: Error checking bucket existence: %v", err)
			return
		}
		if !exists {
			err = w.minioClient.MakeBucket(ctx, w.bucketName, minio.MakeBucketOptions{})
			if err != nil {
				log.Printf("Currency Worker: Error creating bucket: %v", err)
				return
			}
			log.Printf("Currency Worker: Created bucket: %s", w.bucketName)
		}

		// Listen for bucket notifications
		log.Printf("Currency Worker: Listening for events on bucket: %s", w.bucketName)
		for notificationInfo := range w.minioClient.ListenBucketNotification(ctx, w.bucketName, "", "currency_rates.csv", []string{
			"s3:ObjectCreated:Put",
		}) {
			if notificationInfo.Err != nil {
				log.Printf("Currency Worker: Error in event channel: %v", notificationInfo.Err)
				continue
			}

			for _, record := range notificationInfo.Records {
				log.Printf("Currency Worker: Received event: %s for object: %s", record.EventName, record.S3.Object.Key)
				if err := w.processFile(ctx, record.S3.Object.Key); err != nil {
					log.Printf("Currency Worker: Error processing file %s: %v", record.S3.Object.Key, err)
				}
			}
		}
	}()
}

func (w *CurrencyIngestionWorker) processFile(ctx context.Context, objectKey string) error {
	start := time.Now()
	status := "success"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordIngestionJob("currency", status, duration)
	}()

	log.Printf("Currency Worker: Starting ingestion for object: %s", objectKey)

	// Check if file exists
	_, err := w.minioClient.StatObject(ctx, w.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		status = "failure"
		return fmt.Errorf("failed to stat object %s/%s: %w", w.bucketName, objectKey, err)
	}

	// Get the file from MinIO
	object, err := w.minioClient.GetObject(ctx, w.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		status = "failure"
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer func() {
		if err := object.Close(); err != nil {
			log.Printf("Currency Worker: Error closing object: %v", err)
		}
	}()

	// Parse CSV
	reader := csv.NewReader(object)
	_, err = reader.Read() // Skip header
	if err != nil {
		status = "failure"
		return fmt.Errorf("failed to read header: %w", err)
	}

	var rows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			status = "failure"
			return fmt.Errorf("failed to read csv record: %w", err)
		}
		rows = append(rows, record)
	}

	if len(rows) > 0 {
		if err := w.batchInsert(ctx, rows); err != nil {
			status = "failure"
			return err
		}
		return nil
	}

	log.Println("Currency Worker: No rows to insert.")
	return nil
}

func (w *CurrencyIngestionWorker) batchInsert(ctx context.Context, rows [][]string) error {
	var rates []*domain.CurrencyRate

	for _, row := range rows {
		// Expected CSV format: base_currency,target_currency,rate,rate_date
		if len(row) < 4 {
			log.Printf("Currency Worker: Warning - Invalid row format (expected 4 columns): %v", row)
			continue
		}

		baseCurrency := row[0]
		targetCurrency := row[1]
		rateStr := row[2]
		rateDateStr := row[3]

		// Validate currency codes (should be 3 characters)
		if len(baseCurrency) != 3 || len(targetCurrency) != 3 {
			log.Printf("Currency Worker: Warning - Invalid currency code: %s/%s", baseCurrency, targetCurrency)
			continue
		}

		// Parse rate
		rate, err := strconv.ParseFloat(rateStr, 64)
		if err != nil {
			log.Printf("Currency Worker: Warning - Invalid rate for %s/%s: %s", baseCurrency, targetCurrency, rateStr)
			continue
		}

		// Parse date (expected format: YYYY-MM-DD)
		rateDate, err := time.Parse("2006-01-02", rateDateStr)
		if err != nil {
			log.Printf("Currency Worker: Warning - Invalid date for %s/%s: %s", baseCurrency, targetCurrency, rateDateStr)
			continue
		}

		rates = append(rates, &domain.CurrencyRate{
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			Rate:           rate,
			RateDate:       rateDate,
		})
	}

	// Batch insert with size limit
	batchSize := 1000
	for i := 0; i < len(rates); i += batchSize {
		end := i + batchSize
		if end > len(rates) {
			end = len(rates)
		}
		batch := rates[i:end]
		if err := w.repo.InsertCurrencyRates(batch); err != nil {
			return fmt.Errorf("failed to insert currency rates batch: %w", err)
		}
	}

	log.Printf("Currency Worker: Successfully ingested %d currency rates.", len(rates))
	metrics.RecordCurrenciesIngested(len(rates))
	return nil
}
