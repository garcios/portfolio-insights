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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type CurrencyIngestionWorker struct {
	repo        domain.MarketDataRepository
	minioClient *minio.Client
	bucketName  string
	fileName    string
}

func NewCurrencyIngestionWorker(repo domain.MarketDataRepository) (*CurrencyIngestionWorker, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	fileName := "currency_rates.csv"

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
		fileName:    fileName,
	}, nil
}

func (w *CurrencyIngestionWorker) Start(ctx context.Context) {
	go func() {
		// Run daily at midnight
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// Run immediately on startup
		if err := w.processFile(ctx); err != nil {
			log.Printf("Currency Worker: Error processing file: %v", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.processFile(ctx); err != nil {
					log.Printf("Currency Worker: Error processing file: %v", err)
				}
			}
		}
	}()
}

func (w *CurrencyIngestionWorker) processFile(ctx context.Context) error {
	log.Println("Currency Worker: Starting ingestion...")

	// Check if file exists
	_, err := w.minioClient.StatObject(ctx, w.bucketName, w.fileName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object %s/%s: %w", w.bucketName, w.fileName, err)
	}

	// Get the file from MinIO
	object, err := w.minioClient.GetObject(ctx, w.bucketName, w.fileName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Parse CSV
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
	return nil
}
