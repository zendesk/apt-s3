package method

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
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

func TestHandleError(t *testing.T) {
	if os.Getenv("TEST_EXIT") == "1" {
		m := &Method{}
		m.handleError("s3://foobar.s3.amazonaws.com/foo", errors.New("Foobar error"))
		return
	}
	r, w, _ := os.Pipe()
	cmd := exec.Command("go", "test", "github.com/abrener5735/apt-s3/method", "-test.run=TestHandleError")
	cmd.Env = append(os.Environ(), "TEST_EXIT=1")
	cmd.Stdout = w
	cmd.Run()
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "400 URI Failure\nMessage: Foobar error\nURI: s3://foobar.s3.amazonaws.com/foo\n\n") {
		t.Errorf("handleError() unexpected error message %s", buf.String())
	}
}
