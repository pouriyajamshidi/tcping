EXEC_DIR = execuatables/
DATE = $(shell date +%d_%m_%y)

.PHONY: all build clean format
all: build

build: clean
	@mkdir -p $(EXEC_DIR)
	
	@echo "[+] Building the Linux version"
	@go build -ldflags "-s -w" -o $(EXEC_DIR)tcping tcping.go

	@echo "[+] Packaging the Linux version"
	@zip -j $(EXEC_DIR)tcping_Linux_$(DATE).zip $(EXEC_DIR)tcping > /dev/null

	@echo "[+] Removing the Linux binary"
	@rm $(EXEC_DIR)tcping

	@echo
	@echo "[+] Building the Windows version"
	@env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping.exe tcping.go

	@echo "[+] Packaging the Windows version"
	@zip -j $(EXEC_DIR)tcping_Windows_$(DATE).zip $(EXEC_DIR)tcping.exe > /dev/null

	@echo "[+] Removing the Windows binary"
	@rm $(EXEC_DIR)tcping.exe

	@echo
	@echo "[+] Building the MacOS version"
	@env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o $(EXEC_DIR)tcping tcping.go

	@echo "[+] Packaging the MacOS version"
	@zip -j $(EXEC_DIR)tcping_MacOS_$(DATE).zip $(EXEC_DIR)tcping > /dev/null

	@echo "[+] Removing the MacOS binary"
	@rm $(EXEC_DIR)tcping

	@echo "[+] Done"

clean:
	@echo "[+] Cleaning files"
	@rm -rf $(EXEC_DIR)
	@echo "[+] Done"
	@echo

format:
	gofmt -w tcping.go
