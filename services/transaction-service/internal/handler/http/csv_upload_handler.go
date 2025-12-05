// Package http implements HTTP handlers for the transaction service.
package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/garcios/portfolio-insights/services/transaction-service/internal/domain"
)

// CSVUploadHandler handles CSV upload requests.
type CSVUploadHandler struct {
	usecase domain.CSVUploadUsecase
}

// NewCSVUploadHandler creates a new CSV upload handler.
func NewCSVUploadHandler(usecase domain.CSVUploadUsecase) *CSVUploadHandler {
	return &CSVUploadHandler{
		usecase: usecase,
	}
}

// CSVUploadResponse represents the response for a CSV upload.
type CSVUploadResponse struct {
	TotalRecords      int                   `json:"total_records"`
	SuccessfulRecords int                   `json:"successful_records"`
	FailedRecords     int                   `json:"failed_records"`
	Errors            []CSVRowErrorResponse `json:"errors,omitempty"`
}

// CSVRowErrorResponse represents an error for a specific row in the CSV.
type CSVRowErrorResponse struct {
	RowNumber int               `json:"row_number"`
	Row       map[string]string `json:"row,omitempty"`
	Error     string            `json:"error"`
}

// UploadCSV handles the CSV file upload.
func (h *CSVUploadHandler) UploadCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from query parameter or header
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = r.Header.Get("X-User-ID")
	}
	if userID == "" {
		http.Error(w, "user_id is required (query param or X-User-ID header)", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("failed to close file: %v\n", err)
		}
	}()

	// Validate file type
	if header.Header.Get("Content-Type") != "text/csv" && !isCsvFilename(header.Filename) {
		http.Error(w, "file must be a CSV", http.StatusBadRequest)
		return
	}

	// Read file content
	csvData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	// Process CSV
	result, err := h.usecase.UploadCSV(userID, csvData)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to process CSV: %v", err), http.StatusBadRequest)
		return
	}

	// Convert result to response
	response := CSVUploadResponse{
		TotalRecords:      result.TotalRecords,
		SuccessfulRecords: result.SuccessfulRecords,
		FailedRecords:     result.FailedRecords,
		Errors:            make([]CSVRowErrorResponse, len(result.Errors)),
	}

	for i, err := range result.Errors {
		response.Errors[i] = CSVRowErrorResponse{
			RowNumber: err.RowNumber,
			Row:       err.Row,
			Error:     err.Error,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if result.FailedRecords > 0 {
		w.WriteHeader(http.StatusPartialContent) // 206
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("failed to encode response: %v\n", err)
	}
}

func isCsvFilename(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".csv"
}
