# ==================================================
# Constants
# ==================================================

# Meta
SHELL := /bin/bash
MAINTAINER := https://github.com/pouriyajamshidi
DESCRIPTION := Ping TCP ports using tcping. Inspired by Linux's ping utility. Written in Go

# Derived from git so a release only needs a tag pushed - no more
# remembering to also hand-edit this file (this is exactly the failure
# mode behind the "fix: version typo" entry in the changelog). A commit
# exactly on a tag gets the clean tag version; anything else gets
# <branch>-<short-sha>, with a -dirty suffix for uncommitted changes.
GIT_EXACT_TAG := $(shell git describe --tags --exact-match 2>/dev/null)
ifneq ($(GIT_EXACT_TAG),)
VERSION := $(patsubst v%,%,$(GIT_EXACT_TAG))
else
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_DIRTY := $(shell git diff --quiet HEAD 2>/dev/null || echo -dirty)
ifeq ($(GIT_BRANCH),)
VERSION := unknown
else
VERSION := $(GIT_BRANCH)-$(GIT_SHA)$(GIT_DIRTY)
endif
endif

VERSION_PACKAGE := github.com/pouriyajamshidi/tcping/v3/internal/version
GO_LDFLAGS := -ldflags "-s -w -X $(VERSION_PACKAGE).Current=$(VERSION)"
GO_MAIN_PATH := ./cmd/tcping

# IO directories
TARGET_DIR := target
OUTPUT_DIR := output
TAPES_DIR := docs/Images/tapes
GIFS_DIR := docs/Images/gifs

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
	$(GIFS_DIR)/tcping_json_pretty.gif \
	$(GIFS_DIR)/tcping_dns_timing.gif \
	$(GIFS_DIR)/tcping_interface.gif \
	$(GIFS_DIR)/tcping_http.gif \
	$(GIFS_DIR)/tcping_http_verbose.gif \
	$(GIFS_DIR)/tcping_skip_tls.gif

# Conditionals
ifeq ($(OS),Windows_NT)
BIN_NAME := tcping.exe
else
BIN_NAME := tcping
endif

# ==================================================
# Phony targets
# ==================================================

.PHONY: all build release check clean update format vet test container gifs

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
	@go get -u -v ./...
	@go mod tidy
	@echo "[+] Done"

format:
	@echo "[+] Formatting files"
	@gofmt -l -w .

vet:
	@echo "[+] Running Go vet"
	@go vet ./...

test:
	@echo "[+] Running tests"
	@go test ./...

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
	@go build $(GO_LDFLAGS) -o $@ $(GO_MAIN_PATH);

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
	go build $(GO_LDFLAGS) -o $@ $(GO_MAIN_PATH);

# Per-target tcping.exe binary (Windows)
.PRECIOUS: $(TARGET_DIR)/windows-%/tcping.exe
$(TARGET_DIR)/windows-%/tcping.exe: $(TARGET_DIR)/windows-%/
	@echo "[+] Building binary: $@"
	@export GOOS=windows; \
	export GOARCH=$(word 1, $(subst -, ,$*)); \
	[ $(word 2, $(subst -, ,$*)) = static ] && export CGO_ENABLED=0; \
	go build $(GO_LDFLAGS) -o $@ $(GO_MAIN_PATH);

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
	@sha256sum $@ | awk '{print $$2 ": " $$1}'

# .zip archive (Windows)
$(OUTPUT_DIR)/tcping-windows-%.zip: $(TARGET_DIR)/windows-%/tcping.exe $(OUTPUT_DIR)/
	@echo "[+] Compressing binary: $@"
	@zip -j $@ $< >/dev/null
	@sha256sum $@ | awk '{print $$2 ": " $$1}'

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
#
# Built fresh from the current commit on every run (rather than relying on
# whatever "tcping" happens to already be on PATH, which could be an older
# system-installed build). The binary is still named "tcping" and its
# directory is only prepended to PATH for the vhs invocation, so the
# recorded command and its output are unaffected.
GIF_BIN_DIR := $(TARGET_DIR)/gif
GIF_BIN := $(GIF_BIN_DIR)/tcping

.PHONY: gif-binary
gif-binary:
	@echo "[+] Building tcping for GIF generation (version: $(VERSION))"
	@mkdir -p $(GIF_BIN_DIR)
	@go build $(GO_LDFLAGS) -o $(GIF_BIN) $(GO_MAIN_PATH)

$(GIFS_DIR)/%.gif: $(TAPES_DIR)/%.tape gif-binary FORCE
	@echo "[+] Generating GIF: $@"
	@PATH="$(abspath $(GIF_BIN_DIR)):$$PATH" vhs $< -o $@

# ==================================================
# Helpers
# ==================================================

# Force target
# See https://www.gnu.org/software/make/manual/html_node/Force-Targets.html
FORCE:
