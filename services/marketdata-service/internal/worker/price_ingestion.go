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

type PriceIngestionWorker struct {
	repo        domain.MarketDataRepository
	minioClient *minio.Client
	bucketName  string
	fileName    string
}

func NewPriceIngestionWorker(repo domain.MarketDataRepository) (*PriceIngestionWorker, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	fileName := "price.csv"

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

	return &PriceIngestionWorker{
		repo:        repo,
		minioClient: minioClient,
		bucketName:  bucketName,
		fileName:    fileName,
	}, nil
}

func (w *PriceIngestionWorker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		if err := w.processFile(ctx); err != nil {
			log.Printf("Price Worker: Error processing file: %v", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.processFile(ctx); err != nil {
					log.Printf("Price Worker: Error processing file: %v", err)
				}
			}
		}
	}()
}

func (w *PriceIngestionWorker) processFile(ctx context.Context) error {
	log.Println("Price Worker: Starting ingestion...")

	_, err := w.minioClient.StatObject(ctx, w.bucketName, w.fileName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object %s/%s: %w", w.bucketName, w.fileName, err)
	}

	assetMap, err := w.repo.GetAllAssetIDs()
	if err != nil {
		return fmt.Errorf("failed to load asset map: %w", err)
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
		return w.batchInsert(ctx, rows, assetMap)
	}

	log.Println("Price Worker: No rows to insert.")
	return nil
}

func (w *PriceIngestionWorker) batchInsert(ctx context.Context, rows [][]string, assetMap map[string]string) error {
	var prices []*domain.AssetPrice

	for _, row := range rows {
		if len(row) < 2 {
			continue
		}

		symbol := row[0]
		priceStr := row[1]

		assetID, ok := assetMap[symbol]
		if !ok {
			log.Printf("Price Worker: Warning - Symbol %s not found in assets table. Skipping.", symbol)
			continue
		}

		priceVal, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Printf("Price Worker: Warning - Invalid price for %s: %s", symbol, priceStr)
			continue
		}

		timestamp := time.Now()
		if len(row) > 2 && row[2] != "" {
			// Try multiple date/time formats
			formats := []string{
				time.RFC3339,          // 2006-01-02T15:04:05Z07:00
				"2006-01-02",          // YYYY-MM-DD
				"2006-01-02 15:04:05", // YYYY-MM-DD HH:MM:SS
				time.RFC3339Nano,      // 2006-01-02T15:04:05.999999999Z07:00
			}

			parsed := false
			for _, format := range formats {
				if t, err := time.Parse(format, row[2]); err == nil {
					timestamp = t
					parsed = true
					break
				}
			}

			if !parsed {
				log.Printf("Price Worker: Warning - Could not parse timestamp '%s' for %s, using current time", row[2], symbol)
			}
		}

		prices = append(prices, &domain.AssetPrice{
			AssetID:   assetID,
			Price:     priceVal,
			Timestamp: timestamp,
		})
	}

	batchSize := 1000
	for i := 0; i < len(prices); i += batchSize {
		end := i + batchSize
		if end > len(prices) {
			end = len(prices)
		}
		batch := prices[i:end]
		if err := w.repo.InsertPrices(batch); err != nil {
			return err
		}
	}

	log.Printf("Price Worker: Successfully ingested %d prices.", len(prices))
	return nil
}
