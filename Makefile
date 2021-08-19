EXEC_DIR = execuatables/

.PHONY: all
all: build

.PHONY: build
build:
	@mkdir -p $(EXEC_DIR)
	
	@echo "[+] Building the Linux version"
	@go build -ldflags "-s -w" -o $(EXEC_DIR)tcping main.go

	@echo "[+] Packaging the Linux version"
	@zip -j $(EXEC_DIR)tcping_Linux.zip $(EXEC_DIR)tcping > /dev/null

	@echo "[+] Removing the Linux binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the Windows version"
	@env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping.exe main.go

	@echo "[+] Packaging the Windows version"
	@zip -j $(EXEC_DIR)tcping_Windows.zip $(EXEC_DIR)tcping.exe > /dev/null

	@echo "[+] Removing the Windows binary"
	@rm $(EXEC_DIR)tcping.exe

	@echo
	@echo "[+] Building the MacOS version"
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping main.go

	@echo "[+] Packaging the MacOS version"
	@zip -j $(EXEC_DIR)tcping_MacOS.zip $(EXEC_DIR)tcping > /dev/null

	@echo "[+] Removing the MacOS binary"
	@rm $(EXEC_DIR)tcping

	@echo "[+] Done"

.PHONY: clean
clean:
	@rm -rf $(EXEC_DIR)
	@echo "[+] Done"

.PHONY: format
format:
	gofmt -w main.go
