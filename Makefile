# PODX Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)

.PHONY: all build clean install release

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o podx .

install: build
	sudo cp podx /usr/local/bin/

clean:
	rm -f podx podx-*

# Cross-compilation targets
linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o podx-linux-amd64 .

linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o podx-linux-arm64 .

darwin-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o podx-darwin-amd64 .

darwin-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o podx-darwin-arm64 .

windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o podx-windows-amd64.exe .

# Build all platforms
release: clean linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64
	sha256sum podx-* > checksums.txt
	@echo "✓ Built all platforms"
	@ls -la podx-*
