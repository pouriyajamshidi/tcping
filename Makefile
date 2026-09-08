# ==================================================
# Constants
# ==================================================

# Meta
SHELL := /bin/bash
MAINTAINER := https://github.com/pouriyajamshidi
DESCRIPTION := Ping TCP ports using tcping. Inspired by Linux's ping utility. Written in Go

# Read from the version package, which is the one place the version is
# written down. Only the Linux packages and the build messages need it here,
# the binary picks it up from the package itself.
VERSION_FILE := internal/version/version.go
VERSION := $(shell sed -n 's/^var Current = "\(.*\)"/\1/p' $(VERSION_FILE))

GO_LDFLAGS := -ldflags "-s -w"
GO_MAIN_PATH := ./cmd/tcping

# tcping has no cgo code, and it resolves names with Go's own resolver
# rather than the one in libc, so a cgo build buys nothing and only ties the
# binary to the glibc it was built against. Every build is static, including
# the local one, so what you test is what gets released.
export CGO_ENABLED := 0

# Linters. Pinned so a local run and a CI run report the same thing.
# Bumping revive is what pulls in newly added revive rules, see revive.toml.
REVIVE_VERSION := v1.15.0
STATICCHECK_VERSION := 2026.2.1

# Builds the Linux packages from nfpm.yaml. Pinned for the same reason.
NFPM_VERSION := v2.47.0

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
	$(OUTPUT_DIR)/tcping-freebsd-amd64.tar.gz \
	$(OUTPUT_DIR)/tcping-freebsd-arm64.tar.gz
LINUX_ARTIFACTS := \
	$(OUTPUT_DIR)/tcping-linux-amd64.tar.gz \
	$(OUTPUT_DIR)/tcping-linux-arm64.tar.gz \
	$(OUTPUT_DIR)/tcping-amd64.deb \
	$(OUTPUT_DIR)/tcping-arm64.deb \
	$(OUTPUT_DIR)/tcping-amd64.rpm \
	$(OUTPUT_DIR)/tcping-arm64.rpm \
	$(OUTPUT_DIR)/tcping-amd64.apk \
	$(OUTPUT_DIR)/tcping-arm64.apk \
	$(OUTPUT_DIR)/tcping-amd64.pkg.tar.zst \
	$(OUTPUT_DIR)/tcping-arm64.pkg.tar.zst
DARWIN_ARTIFACTS := \
	$(OUTPUT_DIR)/tcping-darwin-amd64.tar.gz \
	$(OUTPUT_DIR)/tcping-darwin-arm64.tar.gz
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

.PHONY: all build release freebsd linux darwin windows check check-format check-fix clean update format fix vet lint staticcheck test container gifs

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

# The one gate to run before pushing. The CI workflows run the same steps,
# so a clean "make check" means a green pull request.
#
# Nothing here rewrites your files. When check-format or check-fix complains,
# "make format" and "make fix" are what apply the changes it asked for.
check: check-format check-fix vet lint staticcheck test

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

fix:
	@echo "[+] Applying Go fixes"
	@go fix ./...

# The two report-only halves of the above, which is what the gate runs so a
# check never edits your files behind your back.
check-format:
	@echo "[+] Checking formatting"
	@files=$$(gofmt -l .); \
	if [ -n "$$files" ]; then \
		echo "[-] These files need formatting, run 'make format':"; \
		echo "$$files"; \
		exit 1; \
	fi

check-fix:
	@echo "[+] Checking for outdated constructs"
	@go fix -diff ./... || \
		(echo "[-] Outdated constructs found, run 'make fix'"; exit 1)

vet:
	@echo "[+] Running Go vet"
	@go vet ./...

lint:
	@echo "[+] Running Revive"
	@go run github.com/mgechev/revive@$(REVIVE_VERSION) -config revive.toml -set_exit_status ./...

# "all" turns on the stylecheck and quickfix checks that are off by default.
staticcheck:
	@echo "[+] Running Staticcheck"
	@go run honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION) -checks=all ./...

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
	go build $(GO_LDFLAGS) -o $@ $(GO_MAIN_PATH);

# Per-target tcping.exe binary (Windows)
.PRECIOUS: $(TARGET_DIR)/windows-%/tcping.exe
$(TARGET_DIR)/windows-%/tcping.exe: $(TARGET_DIR)/windows-%/
	@echo "[+] Building binary: $@"
	@export GOOS=windows; \
	export GOARCH=$*; \
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

# Linux packages
#
# All four formats are built from nfpm.yaml, which is where the file list and
# the package metadata live. The rules only differ in the format they ask for.
$(OUTPUT_DIR)/tcping-%.deb: $(TARGET_DIR)/linux-%/tcping $(UNIX_COMPLETIONS) LICENSE nfpm.yaml $(OUTPUT_DIR)/
	@echo "[+] Creating Debian package: $@"
	@TCPING_ARCH=$* TCPING_VERSION=$(VERSION) TCPING_BIN=$< \
		go run github.com/goreleaser/nfpm/v2/cmd/nfpm@$(NFPM_VERSION) package --packager deb --target $@ >/dev/null
	@sha256sum $@ | awk '{print "    sha256: " $$1}'
	@echo

$(OUTPUT_DIR)/tcping-%.rpm: $(TARGET_DIR)/linux-%/tcping $(UNIX_COMPLETIONS) LICENSE nfpm.yaml $(OUTPUT_DIR)/
	@echo "[+] Creating RPM package: $@"
	@TCPING_ARCH=$* TCPING_VERSION=$(VERSION) TCPING_BIN=$< \
		go run github.com/goreleaser/nfpm/v2/cmd/nfpm@$(NFPM_VERSION) package --packager rpm --target $@ >/dev/null
	@sha256sum $@ | awk '{print "    sha256: " $$1}'
	@echo

$(OUTPUT_DIR)/tcping-%.apk: $(TARGET_DIR)/linux-%/tcping $(UNIX_COMPLETIONS) LICENSE nfpm.yaml $(OUTPUT_DIR)/
	@echo "[+] Creating Alpine package: $@"
	@TCPING_ARCH=$* TCPING_VERSION=$(VERSION) TCPING_BIN=$< \
		go run github.com/goreleaser/nfpm/v2/cmd/nfpm@$(NFPM_VERSION) package --packager apk --target $@ >/dev/null
	@sha256sum $@ | awk '{print "    sha256: " $$1}'
	@echo

# Arch reserves the hyphen for the package release number, so a version like
# 3.0.0-rc1 has to be spelled 3.0.0_rc1 here. nfpm drops the "rc1" instead of
# converting it, which would leave a release candidate claiming to be 3.0.0.
$(OUTPUT_DIR)/tcping-%.pkg.tar.zst: $(TARGET_DIR)/linux-%/tcping $(UNIX_COMPLETIONS) LICENSE nfpm.yaml $(OUTPUT_DIR)/
	@echo "[+] Creating Arch Linux package: $@"
	@TCPING_ARCH=$* TCPING_VERSION=$(subst -,_,$(VERSION)) TCPING_BIN=$< \
		go run github.com/goreleaser/nfpm/v2/cmd/nfpm@$(NFPM_VERSION) package --packager archlinux --target $@ >/dev/null
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
