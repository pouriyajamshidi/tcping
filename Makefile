# ==================================================
# Constants
# ==================================================

# Meta
SHELL := /bin/bash
VERSION := 2.7.0
MAINTAINER := https://github.com/pouriyajamshidi
DESCRIPTION := Ping TCP ports using tcping. Inspired by Linux's ping utility. Written in Go

# IO directories
TARGET_DIR := target
OUTPUT_DIR := output
TAPES_DIR := Images/tapes
GIFS_DIR := Images/gifs

# File lists
RELEASE_ARTIFACTS := \
	$(OUTPUT_DIR)/tcping-freebsd-amd64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-freebsd-amd64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-freebsd-arm64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-freebsd-arm64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-amd64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-amd64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-arm64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-arm64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-amd64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-amd64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-arm64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-arm64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-windows-amd64-static.zip \
	$(OUTPUT_DIR)/tcping-windows-amd64-dynamic.zip \
	$(OUTPUT_DIR)/tcping-windows-arm64-static.zip \
	$(OUTPUT_DIR)/tcping-windows-arm64-dynamic.zip \
	$(OUTPUT_DIR)/tcping-amd64.deb \
	$(OUTPUT_DIR)/tcping-arm64.deb
GIF_ARTIFACTS := \
	$(GIFS_DIR)/tcping.gif \
	$(GIFS_DIR)/tcping_resolve.gif \
	$(GIFS_DIR)/tcping_json_pretty.gif

# Conditionals
ifeq ($(OS),Windows_NT)
BIN_NAME := tcping.exe
else
BIN_NAME := tcping
endif

# ==================================================
# Phony targets
# ==================================================

.PHONY: all build release clean update format vet test tidyup container gifs

all: build

# Build for current platform
build: $(TARGET_DIR)/$(BIN_NAME)

# Build all release artifacts
release: $(RELEASE_ARTIFACTS)

check: format vet test

# Remove all build artifacts
clean:
	rm -rf $(TARGET_DIR)/ $(OUTPUT_DIR)/

update:
	@echo "[+] Updating Go dependencies"
	@go get -u
	@echo "[+] Done"

format:
	@echo "[+] Formatting files"
	@gofmt -w *.go

vet:
	@echo "[+] Running Go vet"
	@go vet

test:
	@echo "[+] Running tests"
	@go test

tidyup:
	@echo "[+] Running go mod tidy"
	@go get -u ./...
	@go mod tidy

container:
	@echo "[+] Building container image"
	@docker build -t tcping:latest .

gifs: $(GIF_ARTIFACTS)

# ==================================================
# Raw binaries
# ==================================================

# Output directory
.PRECIOUS: $(TARGET_DIR)/
$(TARGET_DIR)/:
	@mkdir -p $@

# Binary for current platform
.PRECIOUS: $(TARGET_DIR)/$(BIN_NAME)
$(TARGET_DIR)/$(BIN_NAME): $(TARGET_DIR)/
	@echo "[+] Building binary for current platform: $@"
	@go build -ldflags "-s -w -X main.version=$(VERSION)" -o $@;

# Per-target output directory
.PRECIOUS: $(TARGET_DIR)/%/
$(TARGET_DIR)/%/:
	@mkdir -p $@

# Per-target tcping binary
.PRECIOUS: $(TARGET_DIR)/%/tcping
$(TARGET_DIR)/%/tcping: $(TARGET_DIR)/%/
	@echo "[+] Building binary: $@"
	@export GOOS=$(word 1, $(subst -, ,$*)); \
	export GOARCH=$(word 2, $(subst -, ,$*)); \
	[ $(word 3, $(subst -, ,$*)) = static ] && export CGO_ENABLED=0; \
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $@;

# Per-target tcping.exe binary (Windows)
.PRECIOUS: $(TARGET_DIR)/windows-%/tcping.exe
$(TARGET_DIR)/windows-%/tcping.exe: $(TARGET_DIR)/windows-%/
	@echo "[+] Building binary: $@"
	@export GOOS=windows; \
	export GOARCH=$(word 1, $(subst -, ,$*)); \
	[ $(word 2, $(subst -, ,$*)) = static ] && export CGO_ENABLED=0; \
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $@;

# ==================================================
# Release outputs
# ==================================================

# Output directory
$(OUTPUT_DIR)/:
	@mkdir -p $@

# .tar.gz archive
$(OUTPUT_DIR)/tcping-%.tar.gz: $(TARGET_DIR)/%/tcping $(OUTPUT_DIR)/
	@echo "[+] Compressing binary: $@"
	@tar -C $$(dirname $<) -czvf $@ tcping >/dev/null
	@sha256sum $@

# .zip archive (Windows)
$(OUTPUT_DIR)/tcping-windows-%.zip: $(TARGET_DIR)/windows-%/tcping.exe $(OUTPUT_DIR)/
	@echo "[+] Compressing binary: $@"
	@zip -j $@ $< >/dev/null
	@sha256sum $@

# .deb package (Linux)
$(OUTPUT_DIR)/tcping-%.deb: $(TARGET_DIR)/linux-%-static/tcping $(OUTPUT_DIR)/
	@echo "[+] Creating debian package: $@"
	@PKG_DIR=$$(mktemp -dt make-tcping.XXXXX); \
	\
	install -Dm 755 -t $$PKG_DIR/usr/bin/ $<; \
	\
	mkdir $$PKG_DIR/DEBIAN; pushd $$PKG_DIR/DEBIAN >/dev/null; \
	echo "Package: tcping" >>control; \
	echo "Version: $(VERSION)" >>control; \
	echo "Section: custom" >>control; \
	echo "Priority: optional" >>control; \
	echo "Architecture: $*" >>control; \
	echo "Essential: no" >>control; \
	echo "Maintainer: $(MAINTAINER)" >>control; \
	echo "Description: $(DESCRIPTION)" >>control; \
	popd >/dev/null; \
	\
	dpkg-deb --build $$PKG_DIR $@

# ==================================================
# Miscellaneous outputs
# ==================================================

# GIF generation
$(GIFS_DIR)/%.gif: $(TAPES_DIR)/%.tape FORCE
	@echo "[+] Generating GIF: $@"
	@vhs $< -o $@

# ==================================================
# Helpers
# ==================================================

# Force target
# See https://www.gnu.org/software/make/manual/html_node/Force-Targets.html
FORCE:
