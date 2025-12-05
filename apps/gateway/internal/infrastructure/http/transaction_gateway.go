// Package http implements HTTP gateways/clients.
package http

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/garcios/portfolio-insights/apps/gateway/internal/domain/gateway"
)

// TransactionHTTPGateway implements the TransactionFileGateway interface using HTTP
type TransactionHTTPGateway struct {
	baseURL string
	client  *http.Client
}

// NewTransactionHTTPGateway creates a new TransactionHTTPGateway
func NewTransactionHTTPGateway(baseURL string) gateway.TransactionFileGateway {
	return &TransactionHTTPGateway{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// UploadCSV uploads a CSV file for processing
func (g *TransactionHTTPGateway) UploadCSV(ctx context.Context, userID string, file io.Reader, filename string) error {
	// Create a pipe to stream the request body
	bodyReader, bodyWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(bodyWriter)

	// Create the request
	url := fmt.Sprintf("%s/upload-csv?user_id=%s", g.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the Content-Type header with the boundary
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Run the multipart writing in a goroutine
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			_ = bodyWriter.Close()
		}()
		defer func() {
			_ = multipartWriter.Close()
		}()

		// Create the form file field
		part, err := multipartWriter.CreateFormFile("file", filename)
		if err != nil {
			errChan <- fmt.Errorf("failed to create form file: %w", err)
			return
		}

		// Copy the file content to the multipart writer
		if _, err := io.Copy(part, file); err != nil {
			errChan <- fmt.Errorf("failed to copy file content: %w", err)
			return
		}

		close(errChan)
	}()

	// Execute the request
	resp, err := g.client.Do(req)

	// Check for errors from the goroutine
	select {
	case writeErr := <-errChan:
		if writeErr != nil {
			return writeErr
		}
	default:
	}

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
