EXEC_DIR = executables/
SOURCE_FILES = $(tcping.go statsprinter.go)

.PHONY: all build update clean format test vet gitHubActions container
all: build
check: format vet test

build: clean update format vet test

	@mkdir -p $(EXEC_DIR)
	
	@echo "[+] Building the Linux version"
	@go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)

	@echo "[+] Packaging the Linux version"
	@tar -czvf $(EXEC_DIR)tcping_Linux.tar.gz -C $(EXEC_DIR) tcping > /dev/null

	@echo "[+] Removing the Linux binary"
	@rm $(EXEC_DIR)tcping

	@echo "[+] Building the static Linux version"
	@env GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)

	@echo "[+] Packaging the static Linux version"
	@tar -czvf $(EXEC_DIR)tcping_Linux_static.tar.gz -C $(EXEC_DIR) tcping > /dev/null

	@echo "[+] Removing the static Linux binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the Windows version"
	@env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping.exe $(SOURCE_FILES)

	@echo "[+] Packaging the Windows version"
	@zip -j $(EXEC_DIR)tcping_Windows.zip $(EXEC_DIR)tcping.exe > /dev/null

	@echo "[+] Removing the Windows binary"
	@rm $(EXEC_DIR)tcping.exe

	@echo
	@echo "[+] Building the MacOS version"
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)

	@echo "[+] Packaging the MacOS version"
	@tar -czvf $(EXEC_DIR)tcping_MacOS.tar.gz -C $(EXEC_DIR) tcping > /dev/null

	@echo "[+] Removing the MacOS binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the MacOS ARM version"
	@env GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping $(SOURCE_FILES)
	
	@echo "[+] Packaging the MacOS ARM version"
	@tar -czvf $(EXEC_DIR)tcping_MacOS_ARM.tar.gz -C $(EXEC_DIR) tcping > /dev/null

	@echo "[+] Removing the MacOS ARM binary"
	@rm $(EXEC_DIR)tcping

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
	@echo "[+] Tidying up"
	@go get -u
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
