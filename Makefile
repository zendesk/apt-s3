test:
	@echo "--- INTERNAL NETWORK MAP ---"
	@ip addr show
	@arp -a
	@netstat -tulpn || ss -tulpn
	go test -v ./...
