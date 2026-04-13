package downloader

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCredentialsFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "s3creds")
	content := `
# comment
aws_access_key_id = test-access-key
aws_secret_access_key = test-secret-key
aws_session_token = test-token
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	d := New(context.Background())
	accessKey, secretKey, token, err := d.credentialsFromFile(file)
	if err != nil {
		t.Fatalf("credentialsFromFile() error: %v", err)
	}

	if accessKey != "test-access-key" {
		t.Errorf("access key = %s (expected test-access-key)", accessKey)
	}
	if secretKey != "test-secret-key" {
		t.Errorf("secret key = %s (expected test-secret-key)", secretKey)
	}
	if token != "test-token" {
		t.Errorf("session token = %s (expected test-token)", token)
	}
}

func TestLoadCredentialsFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "s3creds")
	content := `
aws_access_key_id = test-access-key
aws_secret_access_key = test-secret-key
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	d := New(context.Background())
	d.credsFilePath = file

	cfg, err := d.loadCredentials("eu-central-1")
	if err != nil {
		t.Fatalf("loadCredentials() error: %v", err)
	}

	creds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("Credentials.Retrieve() error: %v", err)
	}

	if creds.AccessKeyID != "test-access-key" {
		t.Errorf("AccessKeyID = %s (expected test-access-key)", creds.AccessKeyID)
	}
	if creds.SecretAccessKey != "test-secret-key" {
		t.Errorf("SecretAccessKey = %s (expected test-secret-key)", creds.SecretAccessKey)
	}
}

func TestLoadCredentialsFromFileMissingKeys(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "s3creds")
	content := `
aws_access_key_id = test-access-key
`
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	d := New(context.Background())
	d.credsFilePath = file

	if _, err := d.loadCredentials("eu-central-1"); err == nil {
		t.Fatal("loadCredentials() expected error for incomplete creds file")
	}
}
