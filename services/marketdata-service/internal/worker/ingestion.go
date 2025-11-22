package worker

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type IngestionWorker struct {
	db          *sql.DB
	minioClient *minio.Client
	bucketName  string
	fileName    string
}

func NewIngestionWorker(db *sql.DB) (*IngestionWorker, error) {
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
		db:          db,
		minioClient: minioClient,
		bucketName:  bucketName,
		fileName:    fileName,
	}, nil
}

func (w *IngestionWorker) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Check every hour, or make configurable
		defer ticker.Stop()

		// Run immediately on start
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

	// Check if object exists
	_, err := w.minioClient.StatObject(ctx, w.bucketName, w.fileName, minio.StatObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to stat object %s/%s: %w", w.bucketName, w.fileName, err)
	}

	// Get object
	object, err := w.minioClient.GetObject(ctx, w.bucketName, w.fileName, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	// Parse CSV
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
		return w.batchInsert(ctx, rows)
	}

	log.Println("No rows to insert.")
	return nil
}

func (w *IngestionWorker) batchInsert(ctx context.Context, rows [][]string) error {
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

		if err := w.insertBatch(ctx, tx, batch); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully ingested %d assets.", len(rows))
	return nil
}

func (w *IngestionWorker) insertBatch(ctx context.Context, tx *sql.Tx, batch [][]string) error {
	valueStrings := make([]string, 0, len(batch))
	valueArgs := make([]interface{}, 0, len(batch)*5)

	for i, row := range batch {
		// Assuming CSV columns: symbol, name, type, exchange, currency
		if len(row) < 5 {
			continue // Skip invalid rows
		}

		n := i * 5
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5))
		valueArgs = append(valueArgs, row[0], row[1], row[2], row[3], row[4])
	}

	if len(valueStrings) == 0 {
		return nil
	}

	stmt := fmt.Sprintf(`
		INSERT INTO marketdata.assets (symbol, name, type, exchange, currency)
		VALUES %s
		ON CONFLICT (symbol) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			exchange = EXCLUDED.exchange,
			currency = EXCLUDED.currency,
			updated_at = NOW()
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, stmt, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}

// ConnectDB is a helper to establish DB connection if needed independently
func ConnectDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "garcios"
	}
	if password == "" {
		password = "Password123"
	}
	if dbname == "" {
		dbname = "portfolio"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
