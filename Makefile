# ──────────────────────────────────────────────────────────────────────────────
# Symphony CLI — Makefile
# Gunakan 'make help' untuk melihat semua target yang tersedia.
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: all check build test lint vet clean install tidy snapshot help

# Build variables — di-inject ke binary via ldflags
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE      := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MODULE    := github.com/Reivhell/symphony
LDFLAGS   := -ldflags="-s -w \
               -X '$(MODULE)/cmd.Version=$(VERSION)' \
               -X '$(MODULE)/cmd.Commit=$(COMMIT)' \
               -X '$(MODULE)/cmd.BuildDate=$(DATE)'"
BINARY    := ./bin/symphony

## all: Jalankan check lengkap (default target)
all: check

## check: Jalankan vet + build + test — wajib dijalankan sebelum push
check: vet build test
	@echo ""
	@echo "  ✔ All checks passed. Ready to push."
	@echo ""

## build: Kompilasi binary ke ./bin/symphony
build:
	@echo "Building $(VERSION)..."
	@if not exist ./bin mkdir ./bin
	@go build $(LDFLAGS) -o ./bin/symphony ./main.go
	@echo "  ✔ Built: ./bin/symphony"

## test: Jalankan seluruh test suite dengan race detector
test:
	@echo "Running tests..."
	@go test ./... -race -count=1

## test-cover: Jalankan test dan hasilkan coverage report
test-cover:
	@go test ./... -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "  Coverage report: coverage.html"

## vet: Jalankan go vet
vet:
	@go vet ./...

## lint: Jalankan golangci-lint
lint:
	@golangci-lint run ./...

## clean: Hapus semua build artifacts dan log files
clean:
	@rm -rf ./bin ./dist coverage.out coverage.html
	@find . -name "*.log" -not -path "./.git/*" -delete
	@echo "  ✔ Clean complete."

## install: Install binary ke GOPATH/bin
install:
	@go install $(LDFLAGS) ./main.go

## tidy: Jalankan go mod tidy
tidy:
	@go mod tidy

## snapshot: Build snapshot lokal via goreleaser (tanpa publish)
snapshot:
	@goreleaser release --snapshot --clean

## check-placeholders: Verifikasi tidak ada placeholder yang tertinggal
check-placeholders:
	@if grep -rn "username/symphony" \
	     --include="*.go" --include="*.yaml" --include="*.sh" \
	     --exclude-dir=".git" . 2>/dev/null; then \
	  echo "  ✖ ERROR: Placeholder strings found!" && exit 1; \
	fi
	@echo "  ✔ No placeholder strings found."

## help: Tampilkan daftar semua target yang tersedia
help:
	@echo "Symphony CLI — Available Makefile targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
