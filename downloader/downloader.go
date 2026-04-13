// Package downloader parses an s3 URI and downloads the specified file to the
// filesystem.
package downloader

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// Downloader tracks the region and AWS config and only recreates the config
// if the region has changed
type Downloader struct {
	region string
	cfg    aws.Config
	ctx    context.Context
	s3Cfg  S3Config

	credsFilePath string
}

type S3Config struct {
	EndpointURL    string
	Region         string
	ForcePathStyle bool
}

var (
	awsRegionalHostPattern = regexp.MustCompile(`^([^.]+)\.s3[.-]([a-z0-9-]+)\.amazonaws\.com$`)
	awsGlobalHostPattern   = regexp.MustCompile(`^([^.]+)\.s3\.amazonaws\.com$`)
)

func New(ctx context.Context) *Downloader {
	d := &Downloader{
		ctx:           ctx,
		credsFilePath: "/etc/apt/s3creds",
	}
	return d
}

func (d *Downloader) Configure(cfg S3Config) error {
	if cfg.EndpointURL != "" {
		endpointURL, err := url.Parse(cfg.EndpointURL)
		if err != nil {
			return fmt.Errorf("invalid Acquire::s3::Endpoint: %w", err)
		}
		if endpointURL.Scheme == "" || endpointURL.Host == "" {
			return fmt.Errorf("invalid Acquire::s3::Endpoint: expected full URL, got %q", cfg.EndpointURL)
		}
	}

	d.s3Cfg = cfg
	return nil
}

func parseCredentialValue(line string) string {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (d *Downloader) credentialsFromFile(fileName string) (string, string, string, error) {
	var accessKey, secretKey, token string

	file, err := os.Open(fileName)
	if err != nil {
		return "", "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "aws_access_key_id"):
			accessKey = parseCredentialValue(line)
		case strings.HasPrefix(line, "aws_secret_access_key"):
			secretKey = parseCredentialValue(line)
		case strings.HasPrefix(line, "aws_session_token"):
			token = parseCredentialValue(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", "", err
	}

	return accessKey, secretKey, token, nil
}

// loadCredentials prefers /etc/apt/s3creds when present, then falls back to
// the default AWS credential chain.
func (d *Downloader) loadCredentials(region string) (aws.Config, error) {
	if _, err := os.Stat(d.credsFilePath); err == nil {
		accessKey, secretKey, token, err := d.credentialsFromFile(d.credsFilePath)
		if err != nil {
			return aws.Config{}, err
		}
		if accessKey == "" || secretKey == "" {
			return aws.Config{}, fmt.Errorf("invalid credentials file %q: aws_access_key_id and aws_secret_access_key are required", d.credsFilePath)
		}

		cfg, err := config.LoadDefaultConfig(
			d.ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, token)),
		)
		return cfg, err
	} else if !os.IsNotExist(err) {
		return aws.Config{}, err
	}

	cfg, err := config.LoadDefaultConfig(d.ctx, config.WithRegion(region))
	return cfg, err
}

func (d *Downloader) defaultRegion() string {
	if d.s3Cfg.Region != "" {
		return d.s3Cfg.Region
	}

	return "us-east-1"
}

func (d *Downloader) parseBucketRegion(host string) (string, string, bool) {
	if matches := awsRegionalHostPattern.FindStringSubmatch(host); len(matches) == 3 {
		return matches[1], matches[2], true
	}
	if matches := awsGlobalHostPattern.FindStringSubmatch(host); len(matches) == 2 {
		return matches[1], "us-east-1", true
	}

	return "", "", false
}

func splitPathStyle(pathValue string) (string, string, error) {
	trimmed := strings.Trim(pathValue, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("path must be in format /<bucket>/<key>")
	}

	return parts[0], parts[1], nil
}

// parseUri takes an S3 URI and returns the bucket, region, key, and filename.
// Supported URI examples:
//   - s3://<bucket>.s3.<region>.amazonaws.com/key/file
//   - s3://<bucket>.s3.amazonaws.com/key/file
//   - s3://<bucket>/key/file (when using a custom endpoint)
//   - https://<custom-endpoint>/<bucket>/key/file (when using Acquire::s3::Endpoint)
func (d *Downloader) parseURI(uri string) (string, string, string, string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", "", "", "", err
	}

	switch parsedURL.Scheme {
	case "s3":
		if parsedURL.Host == "" {
			return "", "", "", "", fmt.Errorf("invalid S3 URI: missing bucket/host")
		}

		if bucket, region, ok := d.parseBucketRegion(parsedURL.Host); ok {
			key := strings.TrimPrefix(parsedURL.Path, "/")
			if key == "" {
				return "", "", "", "", fmt.Errorf("invalid S3 URI: missing object key")
			}
			return bucket, region, key, path.Base(key), nil
		}

		if d.s3Cfg.EndpointURL != "" {
			endpointURL, err := url.Parse(d.s3Cfg.EndpointURL)
			if err != nil {
				return "", "", "", "", fmt.Errorf("invalid Acquire::s3::Endpoint: %w", err)
			}
			if strings.EqualFold(parsedURL.Host, endpointURL.Host) {
				bucket, key, err := splitPathStyle(parsedURL.Path)
				if err != nil {
					return "", "", "", "", err
				}
				return bucket, d.defaultRegion(), key, path.Base(key), nil
			}
		}

		key := strings.TrimPrefix(parsedURL.Path, "/")
		if key == "" {
			return "", "", "", "", fmt.Errorf("invalid S3 URI: missing object key")
		}

		return parsedURL.Host, d.defaultRegion(), key, path.Base(key), nil
	case "https", "http":
		if d.s3Cfg.EndpointURL == "" {
			return "", "", "", "", fmt.Errorf("HTTP(S) URI requires Acquire::s3::Endpoint to be configured")
		}

		endpointURL, err := url.Parse(d.s3Cfg.EndpointURL)
		if err != nil {
			return "", "", "", "", fmt.Errorf("invalid Acquire::s3::Endpoint: %w", err)
		}
		if !strings.EqualFold(parsedURL.Host, endpointURL.Host) {
			return "", "", "", "", fmt.Errorf("URI host %q does not match Acquire::s3::Endpoint host %q", parsedURL.Host, endpointURL.Host)
		}

		bucket, key, err := splitPathStyle(parsedURL.Path)
		if err != nil {
			return "", "", "", "", err
		}
		return bucket, d.defaultRegion(), key, path.Base(key), nil
	default:
		return "", "", "", "", fmt.Errorf("unsupported URI scheme: %s", parsedURL.Scheme)
	}
}

func (d *Downloader) newS3Client() *s3.Client {
	return s3.NewFromConfig(d.cfg, func(o *s3.Options) {
		if d.s3Cfg.ForcePathStyle {
			o.UsePathStyle = true
		}
		if d.s3Cfg.EndpointURL != "" {
			o.BaseEndpoint = aws.String(d.s3Cfg.EndpointURL)
		}
	})
}

// GetFileAttributes queries the object in S3 and returns the timestamp and
// size in the format expected by apt
func (d *Downloader) GetFileAttributes(s3Uri string) (string, int64, error) {
	var err error
	bucket, region, key, _, err := d.parseURI(s3Uri)
	if err != nil {
		return "", -1, err
	}

	if d.region != region {
		d.region = region
		d.cfg, err = d.loadCredentials(region)
		if err != nil {
			return "", -1, err
		}
	}

	client := d.newS3Client()

	result, err := client.GetObject(d.ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			return "", -1, errors.New(strings.Join(strings.Split(ae.Error(), "\n"), " "))
		}
		return "", -1, err
	}

	return result.LastModified.Format("2006-01-02T15:04:05+00:00"), *result.ContentLength, nil
}

// DownloadFile pulls the file from an S3 bucket and writes it to the specified
// path
func (d *Downloader) DownloadFile(s3Uri string, path string) (string, error) {
	var err error
	bucket, region, key, filename, err := d.parseURI(s3Uri)
	if err != nil {
		return "", err
	}
	if path != "" {
		filename = path
	}

	if d.region != region {
		d.region = region
		d.cfg, err = d.loadCredentials(region)
		if err != nil {
			return "", err
		}
	}

	client := d.newS3Client()
	downloader := manager.NewDownloader(client)

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := downloader.Download(d.ctx, f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		os.Remove(filename)
		return "", err
	}
	return filename, nil
}
