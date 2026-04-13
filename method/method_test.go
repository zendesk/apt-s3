package method

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/zendesk/apt-s3/downloader"
)

func TestSendCapabilities(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	m := &Method{}

	m.sendCapabilities()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	if buf.String() != "100 Capabilities\nSend-Config: true\nPipeline: true\nSingle-Instance: yes\n\n" {
		t.Errorf("sendCapabilities() unexpected string %s", buf.String())
	}
}

func TestFindLine(t *testing.T) {
	lines := []string{
		"Not this line",
		"Foo: bar",
		"Not this line either",
	}

	m := &Method{}
	key := m.findLine("Foo", lines)
	if key != "bar" {
		t.Errorf("findLine() Foo = %s (expected bar)", key)
	}
}

func TestConfigFromLines(t *testing.T) {
	lines := []string{
		"Config-Item: Acquire::s3::Endpoint=https://storage.zigpos.com",
		"Config-Item: Acquire::s3::Region=eu-central-1",
		"Config-Item: Acquire::s3::ForcePathStyle=true",
	}

	m := New(context.Background())
	cfg, err := m.configFromLines(lines)
	if err != nil {
		t.Fatalf("configFromLines() unexpected error: %v", err)
	}

	if cfg.EndpointURL != "https://storage.zigpos.com" {
		t.Errorf("configFromLines() EndpointURL = %s (expected https://storage.zigpos.com)", cfg.EndpointURL)
	}
	if cfg.Region != "eu-central-1" {
		t.Errorf("configFromLines() Region = %s (expected eu-central-1)", cfg.Region)
	}
	if !cfg.ForcePathStyle {
		t.Error("configFromLines() ForcePathStyle = false (expected true)")
	}
}

func TestConfigFromLinesQuotedValues(t *testing.T) {
	lines := []string{
		"Config-Item: Acquire::s3::Endpoint=\"https://storage.zigpos.com\"",
		"Config-Item: Acquire::s3::Region=\"eu-central-1\"",
		"Config-Item: Acquire::s3::ForcePathStyle=\"true\"",
	}

	m := New(context.Background())
	cfg, err := m.configFromLines(lines)
	if err != nil {
		t.Fatalf("configFromLines() unexpected error: %v", err)
	}

	if cfg.EndpointURL != "https://storage.zigpos.com" {
		t.Errorf("configFromLines() EndpointURL = %s (expected https://storage.zigpos.com)", cfg.EndpointURL)
	}
	if cfg.Region != "eu-central-1" {
		t.Errorf("configFromLines() Region = %s (expected eu-central-1)", cfg.Region)
	}
	if !cfg.ForcePathStyle {
		t.Error("configFromLines() ForcePathStyle = false (expected true)")
	}
}

func TestConfigFromLinesInvalidBool(t *testing.T) {
	lines := []string{
		"Config-Item: Acquire::s3::ForcePathStyle=not-a-bool",
	}

	m := New(context.Background())
	if _, err := m.configFromLines(lines); err == nil {
		t.Error("configFromLines() expected error for invalid bool")
	}
}

func TestApplyConfigFrom601Lines(t *testing.T) {
	lines := []string{
		"601 Configuration",
		"Config-Item: Acquire::s3::Endpoint=https://storage.zigpos.com",
		"Config-Item: Acquire::s3::Region=eu-central-1",
		"Config-Item: Acquire::s3::ForcePathStyle=true",
	}

	m := New(context.Background())
	if err := m.applyConfig(lines); err != nil {
		t.Fatalf("applyConfig() unexpected error: %v", err)
	}

	if m.s3Config.EndpointURL != "https://storage.zigpos.com" {
		t.Errorf("s3Config.EndpointURL = %s (expected https://storage.zigpos.com)", m.s3Config.EndpointURL)
	}
	if m.s3Config.Region != "eu-central-1" {
		t.Errorf("s3Config.Region = %s (expected eu-central-1)", m.s3Config.Region)
	}
	if !m.s3Config.ForcePathStyle {
		t.Error("s3Config.ForcePathStyle = false (expected true)")
	}
}

func TestApplyConfigNoConfigItemsKeepsPrevious(t *testing.T) {
	lines := []string{
		"600 URI Acquire",
		"URI: s3://aptly-dev/ubuntu/dists/noble-updates/InRelease",
	}

	m := New(context.Background())
	m.s3Config = downloader.S3Config{
		EndpointURL:    "https://storage.zigpos.com",
		Region:         "eu-central-1",
		ForcePathStyle: true,
	}
	if err := m.Downloader.Configure(m.s3Config); err != nil {
		t.Fatalf("Configure() unexpected error: %v", err)
	}

	if err := m.applyConfig(lines); err != nil {
		t.Fatalf("applyConfig() unexpected error: %v", err)
	}

	if m.s3Config.EndpointURL != "https://storage.zigpos.com" {
		t.Errorf("s3Config.EndpointURL changed to %s", m.s3Config.EndpointURL)
	}
	if m.s3Config.Region != "eu-central-1" {
		t.Errorf("s3Config.Region changed to %s", m.s3Config.Region)
	}
	if !m.s3Config.ForcePathStyle {
		t.Error("s3Config.ForcePathStyle changed to false")
	}
}

func TestHandleError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	m := &Method{}
	m.handleError("s3://foobar.s3.amazonaws.com/foo", errors.New("Foobar error"))

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "400 URI Failure\nMessage: Foobar error\nURI: s3://foobar.s3.amazonaws.com/foo\n\n") {
		t.Errorf("handleError() unexpected error message %s", buf.String())
	}
}
