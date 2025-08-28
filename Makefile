# Go parameters
TARGET := sdate

# Directories
BIN_DIR := ./bin

# Get the version from the latest git tag. Default to v0.0.0-dev if no tags.
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev")
# Get the commit hash
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Get the build date
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# ldflags to inject version information. Note the quotes to handle spaces.
LDFLAGS := -ldflags="-X 'main.version=$(VERSION) (commit: $(COMMIT), build_date: $(BUILD_DATE))'"

# Cross-compilation targets
GOOS_LIST := linux windows darwin
GOARCH_LIST := amd64

.PHONY: all build clean install test lint release package universal-mac

all: build

build: 
	@echo "Building $(TARGET) for local development..."
	@go build $(LDFLAGS) -o $(TARGET) .

test: 
	@echo "Running tests..."
	@go test -v ./...

lint: 
	@echo "Running linter..."
	@gofmt -l -w . 
	@go vet ./...

clean: 
	@echo "Cleaning up..."
	@rm -f $(TARGET)
	@rm -rf $(BIN_DIR)

release: clean
	@echo "Building release binaries..."
	@mkdir -p $(BIN_DIR)
	@for goos in $(GOOS_LIST); do \
		for goarch in $(GOARCH_LIST); do \
			PLATFORM="$${goos}-$${goarch}"; \
			OUTPUT_DIR="$(BIN_DIR)/$${PLATFORM}"; \
			EXE_NAME="$(TARGET)"; \
			if [ "$$goos" = "windows" ]; then EXE_NAME="$(TARGET).exe"; fi; \
			echo "--> Building for $${PLATFORM}"; \
			mkdir -p "$${OUTPUT_DIR}"; \
			GOOS=$$goos GOARCH=$$goarch go build $(LDFLAGS) -o "$${OUTPUT_DIR}/$${EXE_NAME}" .; \
		done; \
	done

package: release
	@echo "Creating release packages..."
	@for platform_dir in $(BIN_DIR)/*; do \
		if [ -d "$$platform_dir" ]; then \
			PLATFORM=$$(basename "$$platform_dir"); \
			PKG_NAME="$(TARGET)-$(VERSION)-$${PLATFORM}"; \
			PKG_DIR="$(BIN_DIR)/$${PKG_NAME}"; \
			echo "--> Packaging for $${PLATFORM}"; \
			mkdir -p "$${PKG_DIR}"; \
			cp "$${platform_dir}"/* "$${PKG_DIR}/"; \
			cp README.md README.ja.md LICENSE "$${PKG_DIR}/"; \
			( \
				cd $(BIN_DIR) && \
				( \
					if [[ "$$PLATFORM" == "windows"* ]]; then \
						zip -r "$${PKG_NAME}.zip" "$${PKG_NAME}" > /dev/null; \
					else \
						tar -czf "$${PKG_NAME}.tar.gz" "$${PKG_NAME}"; \
					fi \
				) && \
				rm -r "$${PKG_NAME}"; \
			); \
		fi; \
	done
	@echo "Packaging complete. Archives are in $(BIN_DIR)"

universal-mac: clean
	@echo "Building macOS universal binary..."
	@mkdir -p $(BIN_DIR)/darwin-amd64 $(BIN_DIR)/darwin-arm64
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/darwin-amd64/$(TARGET) .
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/darwin-arm64/$(TARGET) .
	@lipo -create -output $(BIN_DIR)/$(TARGET)-darwin-universal $(BIN_DIR)/darwin-amd64/$(TARGET) $(BIN_DIR)/darwin-arm64/$(TARGET)
	@echo "Universal binary created at $(BIN_DIR)/$(TARGET)-darwin-universal"

install: build
	@echo "Installing $(TARGET) to $(shell go env GOPATH)/bin..."
	@go install $(LDFLAGS)