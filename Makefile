EXEC_DIR = execuatables/

.PHONY: all build clean format test vet
all: build
check: format vet test

build: clean format vet test

	@mkdir -p $(EXEC_DIR)
	
	@echo "[+] Building the Linux version"
	@env GOOS=linux go build -ldflags "-s -w" -o $(EXEC_DIR)tcping_linux tcping.go
	@docker build -t tcping:develop . --build-arg GOOS=linux

	@echo "[+] Building the Windows version"
	@env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping.exe tcping.go

	@echo
	@echo "[+] Building the MacOS version"
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping_darwin tcping.go

	@echo "[+] Done"

clean:
	@echo "[+] Cleaning files"
	@rm -rf $(EXEC_DIR)
	@echo "[+] Done"
	@echo

format:
	@echo "[+] Formatting files"
	@gofmt -w *.go

vet:
	@echo "[+] Running Go vet"
	@go vet

test:
	@echo "[+] Running tests"
	@go test
