// Package downloader parses an s3 URI and downloads the specified file to the
// filesystem.
package downloader

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Downloader tracks the region and Session and only recreates the Session
// if the region has changed
type Downloader struct {
	region string
	sess   *session.Session
}

func New() *Downloader {
	d := &Downloader{}
	return d
}

// getValue parses a string and returns the value assigned to a key
func (d *Downloader) getValue(line string) string {
	splitLine := strings.Split(line, " = ")
	return (splitLine[len(splitLine)-1])
}

// credentialsFromFile loads AWS credentials from a non-standard path
func (d *Downloader) credentialsFromFile(fileName string) (string, string, string, error) {
	var accessKey, secretKey, token string

	file, err := os.Open(fileName)
	if err != nil {
		return "", "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "aws_access_key_id") {
			accessKey = d.getValue(scanner.Text())
		} else if strings.Contains(scanner.Text(), "aws_secret_access_key") {
			secretKey = d.getValue(scanner.Text())
		} else if strings.Contains(scanner.Text(), "aws_session_token") {
			token = d.getValue(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", "", err
	}

	return accessKey, secretKey, token, nil
}

// loadCredentials sets up a Session using credentials found in /etc/apt/s3creds
// or using the default configuration supported by AWS if /etc/apt/s3creds does
// not exist
func (d *Downloader) loadCredentials(region string) (*session.Session, error) {
	var config aws.Config
	var sess *session.Session

	if _, err := os.Stat("/etc/apt/s3creds"); err == nil {
		accessKey, secretKey, token, err := d.credentialsFromFile("/etc/apt/s3creds")
		if err != nil {
			return nil, err
		}
		config = aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(accessKey, secretKey, token),
		}
	} else if os.IsNotExist(err) {
		config = aws.Config{Region: aws.String(region)}
	}
	sess, err := session.NewSession(&config)

	return sess, err
}

// parseUri takes an S3 URI s3://<bucket>.s3-<region>.amazonaws.com/key/file
// and returns the bucket, region, key, and filename
func (d *Downloader) parseUri(keyString string) (string, string, string, string) {
	var region string
	ss := strings.Split(keyString, "/")
	bucketSs := strings.Split(ss[2], ".")
	bucket := bucketSs[0]
	regionSs := strings.Split(bucketSs[1], "-")
	// Default to us-east-1 if just <bucket>.s3.amazonaws.com is passed
	if len(regionSs) == 1 {
		region = "us-east-1"
	} else {
		region = strings.Join(regionSs[1:], "-")
	}
	key := strings.Join(ss[3:], "/")
	filename := ss[len(ss)-1]
	return bucket, region, key, filename
}

// GetFileAttributes queries the object in S3 and returns the timestamp and
// size in the format expected by apt
func (d *Downloader) GetFileAttributes(s3Uri string) (string, int64, error) {
	var err error
	bucket, region, key, _ := d.parseUri(s3Uri)

	if d.region != region {
		d.region = region
		d.sess, err = d.loadCredentials(region)
		if err != nil {
			return "", -1, err
		}
	}

	svc := s3.New(d.sess)

	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return "", -1, errors.New(strings.Join(strings.Split(aerr.Error(), "\n"), " "))
		}
	}

	return result.LastModified.Format("2006-01-02T15:04:05+00:00"), *result.ContentLength, nil
}

// DownloadFile pulls the file from an S3 bucket and writes it to the specified
// path
func (d *Downloader) DownloadFile(s3Uri string, path string) (string, error) {
	var err error
	bucket, region, key, filename := d.parseUri(s3Uri)
	if path != "" {
		filename = path
	}

	if d.region != region {
		d.region = region
		d.sess, err = d.loadCredentials(region)
		if err != nil {
			return "", err
		}
	}
	downloader := s3manager.NewDownloader(d.sess)

	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}

	if _, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}); err != nil {
		os.Remove(filename)
		return "", err
	}
	return filename, nil
}
