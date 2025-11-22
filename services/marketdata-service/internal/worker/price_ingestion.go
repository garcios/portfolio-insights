package worker

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type PriceIngestionWorker struct {
	db          *sql.DB
	minioClient *minio.Client
	bucketName  string
	fileName    string
}

func NewPriceIngestionWorker(db *sql.DB) (*PriceIngestionWorker, error) {
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
		db:          db,
		minioClient: minioClient,
		bucketName:  bucketName,
		fileName:    fileName,
	}, nil
}

func (w *PriceIngestionWorker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run immediately on start
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

	// 1. Check if object exists
	_, err := w.minioClient.StatObject(ctx, w.bucketName, w.fileName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object %s/%s: %w", w.bucketName, w.fileName, err)
	}

	// 2. Load Asset Map (Symbol -> ID)
	assetMap, err := w.loadAssetMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to load asset map: %w", err)
	}

	// 3. Get object
	object, err := w.minioClient.GetObject(ctx, w.bucketName, w.fileName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// 4. Parse CSV
	reader := csv.NewReader(object)

	// Read header
	_, err = reader.Read()
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

func (w *PriceIngestionWorker) loadAssetMap(ctx context.Context) (map[string]string, error) {
	rows, err := w.db.QueryContext(ctx, "SELECT symbol, id FROM marketdata.assets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assetMap := make(map[string]string)
	for rows.Next() {
		var symbol, id string
		if err := rows.Scan(&symbol, &id); err != nil {
			return nil, err
		}
		assetMap[symbol] = id
	}
	return assetMap, nil
}

func (w *PriceIngestionWorker) batchInsert(ctx context.Context, rows [][]string, assetMap map[string]string) error {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Batch size
	batchSize := 1000
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		if err := w.insertBatch(ctx, tx, batch, assetMap); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Price Worker: Successfully ingested %d prices.", len(rows))
	return nil
}

func (w *PriceIngestionWorker) insertBatch(ctx context.Context, tx *sql.Tx, batch [][]string, assetMap map[string]string) error {
	valueStrings := make([]string, 0, len(batch))
	valueArgs := make([]interface{}, 0, len(batch)*3)

	for _, row := range batch {
		// CSV: symbol, price, timestamp (optional)
		if len(row) < 2 {
			continue
		}

		symbol := row[0]
		priceStr := row[1]

		// Lookup Asset ID
		assetID, ok := assetMap[symbol]
		if !ok {
			log.Printf("Price Worker: Warning - Symbol %s not found in assets table. Skipping.", symbol)
			continue
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Printf("Price Worker: Warning - Invalid price for %s: %s", symbol, priceStr)
			continue
		}

		timestamp := time.Now()
		if len(row) > 2 && row[2] != "" {
			// Try parsing timestamp, fallback to now if fail
			// Assuming ISO8601/RFC3339 for simplicity: "2023-10-27T10:00:00Z"
			if t, err := time.Parse(time.RFC3339, row[2]); err == nil {
				timestamp = t
			}
		}

		n := len(valueStrings) * 3
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", n+1, n+2, n+3))
		valueArgs = append(valueArgs, assetID, price, timestamp)
	}

	if len(valueStrings) == 0 {
		return nil
	}

	stmt := fmt.Sprintf(`
		INSERT INTO marketdata.asset_prices (asset_id, price, timestamp)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}
