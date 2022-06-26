EXEC_DIR = executables/

.PHONY: all build clean format test vet
all: build
check: format vet test

build: clean format vet test

	@mkdir -p $(EXEC_DIR)
	
	@echo "[+] Building the Linux version"
	@go build -ldflags "-s -w" -o $(EXEC_DIR)tcping tcping.go

	@echo "[+] Packaging the Linux version"
	@zip -j $(EXEC_DIR)tcping_Linux.zip $(EXEC_DIR)tcping > /dev/null

	@echo "[+] Removing the Linux binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the Windows version"
	@env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping.exe tcping.go

	@echo "[+] Packaging the Windows version"
	@zip -j $(EXEC_DIR)tcping_Windows.zip $(EXEC_DIR)tcping.exe > /dev/null

	@echo "[+] Removing the Windows binary"
	@rm $(EXEC_DIR)tcping.exe

	@echo
	@echo "[+] Building the MacOS version"
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping tcping.go

	@echo "[+] Packaging the MacOS version"
	@zip -j $(EXEC_DIR)tcping_MacOS.zip $(EXEC_DIR)tcping > /dev/null

	@echo "[+] Removing the MacOS binary"
	@rm $(EXEC_DIR)tcping

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

container:
	@echo "[+] Building container image"
	@env GOOS=linux CGO_ENABLED=0 go build --ldflags '-s -w -extldflags "-static"' -o $(EXEC_DIR)tcping tcping.go statsprinter.go && chmod +x $(EXEC_DIR)tcping
	@docker build -t tcping:latest .
	@rm $(EXEC_DIR)tcping
	@echo "[+] Done"
