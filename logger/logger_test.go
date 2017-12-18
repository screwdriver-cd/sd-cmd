package logger

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
)

type writeCloser struct {
	buf []byte
}

func (w *writeCloser) Write(v []byte) (int, error) {
	w.buf = v
	return len(v), nil
}

func (w *writeCloser) Close() error { return nil }

func setup() {
	config.SDArtifactsDir = "/opt/sd/"
}

func teardown() {
	logger = &Logger{
		Path:   "",
		file:   nil,
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}
func TestMakeLogToFile(t *testing.T) {
	logger = new(Logger)
	// success
	err := MakeLogToFile([]string{"log/test@1.0"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	defer os.RemoveAll("/opt/sd")
	defer Close()
	_, err = os.Stat(logger.Path)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// failure
	err = MakeLogToFile([]string{"sample"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestWrite(t *testing.T) {
	// success
	logger = new(Logger)
	wc := new(writeCloser)
	logger = &Logger{
		Path:   "/opt/sd",
		file:   wc,
		logger: log.New(wc, "", log.LstdFlags),
	}
	Write("log message")
	str := string(wc.buf)
	if !strings.Contains(str, "log message") {
		t.Errorf("log=%q, should include %q", str, "log message")
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
