# SQL Parser Go Makefile

# Variables
BINARY_NAME=sqlparser
MAIN_PATH=./cmd/sqlparser
BUILD_DIR=./bin
PACKAGES=./...

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golint

.PHONY: all build clean test deps fmt lint run help install

# Default target
all: deps fmt lint test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v $(PACKAGES)

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Lint code
lint:
	@echo "Linting code..."
	@command -v golint >/dev/null 2>&1 || { echo "golint not installed. Installing..."; $(GOGET) -u golang.org/x/lint/golint; }
	golint $(PACKAGES)

# Run the application
run:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)

# Development commands
dev-query:
	@echo "Running development query analysis..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME) -query ./examples/queries/complex_query.sql -output table -verbose

dev-simple:
	@echo "Running simple query analysis..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME) -query ./examples/queries/simple_query.sql -output json

dev-log:
	@echo "Running log analysis..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	./$(BUILD_DIR)/$(BINARY_NAME) -log ./examples/logs/sample_profiler.log -output table

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all $(PACKAGES)

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@command -v gosec >/dev/null 2>&1 || { echo "gosec not installed. Installing..."; $(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	gosec $(PACKAGES)

# Run benchmarks
bench:
	@echo "Running performance benchmarks..."
	$(GOTEST) -run=none -bench=. -benchmem $(PACKAGES)

# Run benchmarks with CPU profiling
bench-cpu:
	@echo "Running benchmarks with CPU profiling..."
	$(GOTEST) -run=none -bench=. -benchmem -cpuprofile=cpu.prof $(PACKAGES)

# Run benchmarks with memory profiling  
bench-mem:
	@echo "Running benchmarks with memory profiling..."
	$(GOTEST) -run=none -bench=. -benchmem -memprofile=mem.prof $(PACKAGES)

# Performance comparison
perf-compare:
	@echo "Running performance comparison..."
	$(GOTEST) -run=none -bench=BenchmarkParser -count=5 -benchmem $(PACKAGES)

# Build optimized release version
build-release:
	@echo "Building optimized release version..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-s -w" -gcflags="-m" -o $(BUILD_DIR)/$(BINARY_NAME)-optimized $(MAIN_PATH)

# Build with performance profiling enabled
build-profile:
	@echo "Building with profiling enabled..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -tags profile -o $(BUILD_DIR)/$(BINARY_NAME)-profile $(MAIN_PATH)

# Optimize build
build-optimized:
	@echo "Building optimized $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="-s -w" -gcflags="-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-optimized $(MAIN_PATH)

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Run deps, fmt, lint, test, and build"
	@echo "  build        - Build the application"
	@echo "  build-optimized - Build with optimization flags"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  test         - Run tests"
	@echo "  bench        - Run benchmarks with memory stats"
	@echo "  perf         - Run performance tests"
	@echo "  profile      - Run with CPU profiling"
	@echo "  memcheck     - Check memory usage (requires valgrind)"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  run          - Build and run the application"
	@echo "  install      - Install binary to GOPATH/bin"
	@echo "  dev-query    - Run development query analysis"
	@echo "  dev-simple   - Run simple query analysis"
	@echo "  dev-log      - Run log analysis"
	@echo "  build-all    - Build for multiple platforms"
	@echo "  docs         - Generate documentation"
	@echo "  security     - Check for security vulnerabilities"
	@echo "  help         - Show this help"
