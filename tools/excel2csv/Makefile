.PHONY: build install clean test help

# Binary names
EXCEL2CSV=excel2csv
CSVTXN=csvtxn

# Build both binaries
build: build-excel2csv build-csvtxn

# Build excel2csv
build-excel2csv:
	@echo "Building $(EXCEL2CSV)..."
	GOWORK=off go build -o bin/$(EXCEL2CSV) excel2csv.go
	@echo "✓ Build complete: bin/$(EXCEL2CSV)"

# Build csvtxn
build-csvtxn:
	@echo "Building $(CSVTXN)..."
	GOWORK=off go build -o bin/$(CSVTXN) csvtxn.go
	@echo "✓ Build complete: bin/$(CSVTXN)"

# Install the binaries to $GOPATH/bin
install:
	@echo "Installing $(EXCEL2CSV) and $(CSVTXN)..."
	GOWORK=off go build -o bin/$(EXCEL2CSV) excel2csv.go
	GOWORK=off go build -o bin/$(CSVTXN) csvtxn.go
	@echo "✓ Installed to bin/"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f bin/$(EXCEL2CSV) bin/$(CSVTXN)
	rm -f ioFiles/*.csv
	@echo "✓ Clean complete"

# Run tests (if any)
test:
	@echo "Running tests..."
	go test -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "✓ Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "✓ Dependencies tidied"

# Run the tool with a sample file (if exists)
run-excel2csv:
	@if [ -f "ioFiles/AllTradesReport.xlsx" ]; then \
		./bin/${EXCEL2CSV} --sheets "Combined" ioFiles/AllTradesReport.xlsx; \
	else \
		echo "No AllTradesReport.xlsx file found"; \
	fi

run-csvtxn:
	@if [ -f "ioFiles/AllTradesReport_Combined.csv" ]; then \
		./bin/${CSVTXN} ioFiles/AllTradesReport_Combined.csv; \
	else \
		echo "No AllTradesReport_Combined.csv file found"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  build            - Build both tools (excel2csv and csvtxn)"
	@echo "  build-excel2csv  - Build excel2csv only"
	@echo "  build-csvtxn     - Build csvtxn only"
	@echo "  install          - Install both tools to GOPATH/bin"
	@echo "  clean            - Remove build artifacts"
	@echo "  test             - Run tests"
	@echo "  deps             - Download dependencies"
	@echo "  tidy             - Tidy dependencies"
	@echo "  run              - Run excel2csv with AllTradesReport.xlsx"
	@echo "  help             - Show this help message"
