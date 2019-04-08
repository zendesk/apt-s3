// Package method interacts with apt via stdin and stdout
package method

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/zendesk/apt-s3/downloader"
)

type Method struct {
	Downloader *downloader.Downloader
}

func New() *Method {
	m := &Method{
		Downloader: downloader.New(),
	}
	return m
}

// calculateHash calculates and returns a single hash. Used by calculateHashes
func (m *Method) calculateHash(h hash.Hash, f []byte) (string, error) {
	if _, err := io.Copy(h, bytes.NewReader(f)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// calculateHashes returns md5, sha1, sha256, and sha512 hashes of the downloaded file
func (m *Method) calculateHashes(filename string) (string, string, string, string, error) {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
	  return "", "", "", "", err
	}

	md5h, err := m.calculateHash(md5.New(), f)
	if err != nil {
	  return "", "", "", "", err
	}
	sha1h, err := m.calculateHash(sha1.New(), f)
	if err != nil {
	  return "", "", "", "", err
	}
	sha256h, err := m.calculateHash(sha256.New(), f)
	if err != nil {
	  return "", "", "", "", err
	}
	sha512h, err := m.calculateHash(sha512.New(), f)
	if err != nil {
	  return "", "", "", "", err
	}

	return md5h, sha1h, sha256h, sha512h, nil
}

// sendCapabilities tells apt what this method is capable of
func (m *Method) sendCapabilities() {
	fmt.Printf("100 Capabilities\nSend-Config: true\nPipeline: true\nSingle-Instance: yes\n\n")
}

// findLine finds the line that starts with key: and returns the value
func (m *Method) findLine(key string, lines []string) string {
	for i := 0; i < len(lines); i++ {
		linesSs := strings.Split(lines[i], ": ")
		if linesSs[0] == key {
			return linesSs[1]
		}
	}

	return ""
}

// UriStart downloads a file from S3 and tells apt about when the download
// starts and is finished
func (m *Method) UriStart(lines []string) error {
	uri := m.findLine("URI", lines)
	path := m.findLine("Filename", lines)

	lastModified, size, err := m.Downloader.GetFileAttributes(uri)
	if err != nil {
		m.handleError(uri, err)
	}

	fmt.Printf("200 URI Start\nLast-Modified: %s\nSize: %d\nURI: %s\n\n", lastModified, size, uri)

	filename, err := m.Downloader.DownloadFile(uri, path)
	if err != nil {
		m.handleError(uri, err)
	}
	md5Hash, sha1Hash, sha256Hash, sha512Hash, err := m.calculateHashes(filename)
	if err != nil {
		return err
	}
	fmt.Printf("201 URI Done\nFilename: %s\nLast-Modified: %s\n", filename, lastModified)
	fmt.Printf("MD5-Hash: %s\nMD5Sum-Hash: %s\nSHA1-Hash: %s\n", md5Hash, md5Hash, sha1Hash)
	fmt.Printf("SHA256-Hash: %s\nSHA512-Hash: %s\n", sha256Hash, sha512Hash)
	fmt.Printf("Size: %d\nURI: %s\n\n", size, uri)

	return nil
}

// handleError sends an error message to os.Stdout in a format which apt
// understands
func (m *Method) handleError(uri string, err error) {
	fmt.Printf("400 URI Failure\nMessage: %s\nURI: %s\n", strings.TrimRight(fmt.Sprintln(err), "\n"), uri)
	os.Exit(1)
}

// Start watches os.Stdin for a "600 URI Acquire" message from apt which
// triggers UriStart
func (m *Method) Start() {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	m.sendCapabilities()

	for scanner.Scan() {
		t := scanner.Text()
		if t != "" {
			lines = append(lines, t)
		} else {
			if len(lines) > 0 {
				if lines[0] == "600 URI Acquire" {
					if err := m.UriStart(lines); err != nil {
						m.handleError(strings.Split(lines[1], ": ")[1], err)
					}
				}
				lines = make([]string, 0)
			}
		}
	}
}
