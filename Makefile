EXEC_DIR = executables/
TAPE_DIR = Images/tapes/
GIFS_DIR = Images/gifs/
SOURCE_FILES = $(tcping.go statsprinter.go)
PACKAGE_NAME = tcping
VERSION = 2.0.0
ARCHITECTURE = amd64
DEB_PACKAGE_DIR = $(EXEC_DIR)/debian
DEBIAN_DIR = $(DEB_PACKAGE_DIR)/DEBIAN
CONTROL_FILE = $(DEBIAN_DIR)/control
EXECUTABLE_PATH = $(EXEC_DIR)/tcping
TARGET_EXECUTABLE_PATH = $(DEB_PACKAGE_DIR)/usr/bin/
PACKAGE = $(PACKAGE_NAME)_$(ARCHITECTURE).deb
MAINTAINER = https://github.com/pouriyajamshidi
DESCRIPTION = Ping TCP ports using tcping. Inspired by Linux's ping utility. Written in Go

.PHONY: all build update clean format test vet gitHubActions container
all: build
check: format vet test

build: clean update tidyup format vet test
	@echo
	@echo "[+] Version: $(VERSION)"
	@echo

	@mkdir -p $(EXEC_DIR)
	@mkdir -p $(DEB_PACKAGE_DIR)
	@mkdir -p $(DEBIAN_DIR)
	@mkdir -p $(TARGET_EXECUTABLE_PATH)
	
	@echo "[+] Building the Linux version"
	@go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)

	@echo "[+] Packaging the Linux version"
	@tar -czvf $(EXEC_DIR)tcping_Linux.tar.gz -C $(EXEC_DIR) tcping > /dev/null
	@sha256sum $(EXEC_DIR)tcping_Linux.tar.gz

	@echo "[+] Removing the Linux binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the static Linux version"
	@env GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)

	@echo "[+] Packaging the static Linux version"
	@tar -czvf $(EXEC_DIR)tcping_Linux_static.tar.gz -C $(EXEC_DIR) tcping > /dev/null
	@sha256sum $(EXEC_DIR)tcping_Linux_static.tar.gz

	@echo
	@echo "[+] Building the Debian package"
	@cp $(EXECUTABLE_PATH) $(TARGET_EXECUTABLE_PATH)

	@echo "[+] Creating control file"
	@echo "Package: $(PACKAGE_NAME)" > $(CONTROL_FILE)
	@echo "Version: $(VERSION)" >> $(CONTROL_FILE)
	@echo "Section: custom" >> $(CONTROL_FILE)
	@echo "Priority: optional" >> $(CONTROL_FILE)
	@echo "Architecture: amd64" >> $(CONTROL_FILE)
	@echo "Essential: no" >> $(CONTROL_FILE)
	@echo "Installed-Size: 2048" >> $(CONTROL_FILE)
	@echo "Maintainer: $(MAINTAINER)" >> $(CONTROL_FILE)
	@echo "Description: $(DESCRIPTION)" >> $(CONTROL_FILE)

	@echo "[+] Building package"
	@dpkg-deb --build $(DEB_PACKAGE_DIR)

	@echo "[+] Renaming package"
	@mv $(DEB_PACKAGE_DIR).deb $(EXEC_DIR)/$(PACKAGE)

	@echo "[+] Removing the static Linux binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the Windows version"
	@env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping.exe $(SOURCE_FILES)

	@echo "[+] Packaging the Windows version"
	@zip -j $(EXEC_DIR)tcping_Windows.zip $(EXEC_DIR)tcping.exe > /dev/null
	@sha256sum  $(EXEC_DIR)tcping_Windows.zip

	@echo "[+] Removing the Windows binary"
	@rm $(EXEC_DIR)tcping.exe

	@echo
	@echo "[+] Building the MacOS version"
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)

	@echo "[+] Packaging the MacOS version"
	@tar -czvf $(EXEC_DIR)tcping_MacOS.tar.gz -C $(EXEC_DIR) tcping > /dev/null
	@sha256sum $(EXEC_DIR)tcping_MacOS.tar.gz

	@echo "[+] Removing the MacOS binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the MacOS ARM version"
	@env GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)
	
	@echo "[+] Packaging the MacOS ARM version"
	@tar -czvf $(EXEC_DIR)tcping_MacOS_ARM.tar.gz -C $(EXEC_DIR) tcping > /dev/null
	@sha256sum $(EXEC_DIR)tcping_MacOS_ARM.tar.gz

	@echo "[+] Removing the MacOS ARM binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the FreeBSD version"
	@env GOOS=freebsd GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)

	@echo "[+] Packaging the FreeBSD AMD64 version"
	@tar -czvf $(EXEC_DIR)tcping_freebsd.tar.gz -C $(EXEC_DIR) tcping > /dev/null
	@sha256sum $(EXEC_DIR)tcping_freebsd.tar.gz

	@echo "[+] Removing the FreeBSD binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Done"

update:
	@echo "[+] Updating Go dependencies"
	@go get -u
	@echo "[+] Done"

clean:
	@echo "[+] Cleaning files"
	@rm -rf $(EXEC_DIR)
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
	@env GOOS=linux CGO_ENABLED=0 go build --ldflags '-s -w -extldflags "-static"' -o $(EXEC_DIR)tcpingDocker $(SOURCE_FILES)
	@docker build -t tcping:latest .
	@rm $(EXEC_DIR)tcpingDocker
	@echo "[+] Done"

gitHubActions:
	@echo "[+] Building container image - GitHub Actions"
	@env GOOS=linux CGO_ENABLED=0 go build --ldflags '-s -w -extldflags "-static"' -o tcpingDocker $(SOURCE_FILES)
	@echo "[+] Done"

gifs:
	@echo "[+] Making tcping.gif"
	@vhs $(TAPE_DIR)tcping.tape -o $(GIFS_DIR)tcping.gif
	@echo "[+] Done"

	@echo "[+] Making tcping_resolve.gif"
	@vhs $(TAPE_DIR)tcping_resolve.tape -o $(GIFS_DIR)tcping_resolve.gif
	@echo "[+] Done"

	@echo "[+] Making tcping_json_pretty.gif"
	@vhs $(TAPE_DIR)tcping_json_pretty.tape -o $(GIFS_DIR)tcping_json_pretty.gif
	@echo "[+] Done"
