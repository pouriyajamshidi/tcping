# ==================================================
# Constants
# ==================================================

# Meta
SHELL := /bin/bash
MAINTAINER := https://github.com/pouriyajamshidi
DESCRIPTION := Ping TCP ports using tcping. Inspired by Linux's ping utility. Written in Go

# Read from the version package, which is the one place the version is
# written down. Only the .deb package and the build messages need it here,
# the binary picks it up from the package itself.
VERSION_FILE := internal/version/version.go
VERSION := $(shell sed -n 's/^var Current = "\(.*\)"/\1/p' $(VERSION_FILE))

GO_LDFLAGS := -ldflags "-s -w"
GO_MAIN_PATH := ./cmd/tcping

# IO directories
TARGET_DIR := target
OUTPUT_DIR := output
TAPES_DIR := docs/Images/tapes
GIFS_DIR := docs/Images/gifs
COMPLETIONS_DIR := completions

# Shipped inside the release archives so they can be installed without
# cloning the repository. PowerShell only goes in the Windows zip.
UNIX_COMPLETIONS := \
	$(COMPLETIONS_DIR)/tcping.bash \
	$(COMPLETIONS_DIR)/_tcping \
	$(COMPLETIONS_DIR)/tcping.fish
WINDOWS_COMPLETIONS := $(COMPLETIONS_DIR)/tcping.ps1

# File lists
# One list per platform so a single platform can be built on its own,
# for example "make windows" instead of the full "make release".
FREEBSD_ARTIFACTS := \
	$(OUTPUT_DIR)/tcping-freebsd-amd64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-freebsd-amd64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-freebsd-arm64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-freebsd-arm64-dynamic.tar.gz
LINUX_ARTIFACTS := \
	$(OUTPUT_DIR)/tcping-linux-amd64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-amd64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-arm64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-arm64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-amd64.deb \
	$(OUTPUT_DIR)/tcping-arm64.deb
DARWIN_ARTIFACTS := \
	$(OUTPUT_DIR)/tcping-darwin-amd64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-amd64-dynamic.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-arm64-static.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-arm64-dynamic.tar.gz
WINDOWS_ARTIFACTS := \
	$(OUTPUT_DIR)/tcping-windows-amd64.zip \
	$(OUTPUT_DIR)/tcping-windows-arm64.zip

RELEASE_ARTIFACTS := \
	$(FREEBSD_ARTIFACTS) \
	$(LINUX_ARTIFACTS) \
	$(DARWIN_ARTIFACTS) \
	$(WINDOWS_ARTIFACTS)
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

.PHONY: all build release freebsd linux darwin windows check clean update format vet test container gifs

all: build

# Build for current platform
build: $(TARGET_DIR)/$(BIN_NAME)

# Build all release artifacts
release: $(RELEASE_ARTIFACTS)
	@echo "[+] Checksums for the release page"
	@echo
	@sha256sum $(RELEASE_ARTIFACTS) | awk '{sub(".*/", "", $$2); print $$2 ": " $$1}'

# Build the release artifacts of a single platform
freebsd: $(FREEBSD_ARTIFACTS)
	@echo "[+] Checksums for the release page"
	@echo
	@sha256sum $(FREEBSD_ARTIFACTS) | awk '{sub(".*/", "", $$2); print $$2 ": " $$1}'

linux: $(LINUX_ARTIFACTS)
	@echo "[+] Checksums for the release page"
	@echo
	@sha256sum $(LINUX_ARTIFACTS) | awk '{sub(".*/", "", $$2); print $$2 ": " $$1}'

darwin: $(DARWIN_ARTIFACTS)
	@echo "[+] Checksums for the release page"
	@echo
	@sha256sum $(DARWIN_ARTIFACTS) | awk '{sub(".*/", "", $$2); print $$2 ": " $$1}'

windows: $(WINDOWS_ARTIFACTS)
	@echo "[+] Checksums for the release page"
	@echo
	@sha256sum $(WINDOWS_ARTIFACTS) | awk '{sub(".*/", "", $$2); print $$2 ": " $$1}'

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
#
# There is no static/dynamic split here like there is for the other
# platforms: tcping does not use cgo on Windows, so both flavors came out
# byte-identical. We just always build static.
.PRECIOUS: $(TARGET_DIR)/windows-%/tcping.exe
$(TARGET_DIR)/windows-%/tcping.exe: $(TARGET_DIR)/windows-%/
	@echo "[+] Building binary: $@"
	@export GOOS=windows; \
	export GOARCH=$*; \
	export CGO_ENABLED=0; \
	go build $(GO_LDFLAGS) -o $@ $(GO_MAIN_PATH);

# ==================================================
# Release outputs
# ==================================================

# Output directory
$(OUTPUT_DIR)/:
	@mkdir -p $@

# .tar.gz archive
$(OUTPUT_DIR)/tcping-%.tar.gz: $(TARGET_DIR)/%/tcping $(UNIX_COMPLETIONS) $(OUTPUT_DIR)/
	@echo "[+] Compressing binary: $@"
	@tar -C $$(dirname $<) -czvf $@ tcping -C "$(CURDIR)" $(UNIX_COMPLETIONS) >/dev/null
	@sha256sum $@ | awk '{print "    sha256: " $$1}'
	@echo

# .zip archive (Windows)
$(OUTPUT_DIR)/tcping-windows-%.zip: $(TARGET_DIR)/windows-%/tcping.exe $(WINDOWS_COMPLETIONS) $(OUTPUT_DIR)/
	@echo "[+] Compressing binary: $@"
	@zip -j $@ $< $(WINDOWS_COMPLETIONS) >/dev/null
	@sha256sum $@ | awk '{print "    sha256: " $$1}'
	@echo

# .deb package (Linux)
$(OUTPUT_DIR)/tcping-%.deb: $(TARGET_DIR)/linux-%-static/tcping $(UNIX_COMPLETIONS) $(OUTPUT_DIR)/
	@echo "[+] Creating debian package: $@"
	@PKG_DIR=$$(mktemp -dt make-tcping.XXXXX); \
	\
	chmod 755 $$PKG_DIR; \
	\
	install -Dm 755 -t $$PKG_DIR/usr/bin/ $<; \
	\
	install -Dm 644 $(COMPLETIONS_DIR)/tcping.bash $$PKG_DIR/usr/share/bash-completion/completions/tcping; \
	install -Dm 644 $(COMPLETIONS_DIR)/_tcping $$PKG_DIR/usr/share/zsh/vendor-completions/_tcping; \
	install -Dm 644 $(COMPLETIONS_DIR)/tcping.fish $$PKG_DIR/usr/share/fish/vendor_completions.d/tcping.fish; \
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
	dpkg-deb --build $$PKG_DIR $@ >/dev/null
	@sha256sum $@ | awk '{print "    sha256: " $$1}'
	@echo

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
