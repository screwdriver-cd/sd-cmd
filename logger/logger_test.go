package logger

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/screwdriver-cd/sd-cmd/config"
)

var tempDir string

type dummyLogFile struct {
	buffer *bytes.Buffer
}

func (d *dummyLogFile) Close() error { return nil }
func (d *dummyLogFile) Write(p []byte) (n int, err error) {
	return d.buffer.Write(p)
}

func setup() {
	tempDir, _ = ioutil.TempDir("", "sd-cmd_logger")
}

func teardown() {
	os.RemoveAll(config.SDArtifactsDir)
}

func TestNew(t *testing.T) {
	// success
	dir := filepath.Join(tempDir, "CreateLogFile")
	filename := fmt.Sprintf("logger_test_%v", time.Now().Unix())
	_, err := New(dir, filename, log.Ldate, false)
	defer os.Remove(filepath.Join(dir, filename))
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestCreateLogFile(t *testing.T) {
	dir := filepath.Join(tempDir, "CreateLogFile")
	filename := fmt.Sprintf("logger_test_%v", time.Now().Unix())
	defer os.RemoveAll(dir)
	CreateLogFile(dir, filename)

	_, err := os.Stat(dir)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// check there is a file
	fileInfos, _ := ioutil.ReadDir(dir)
	if len(fileInfos) > 0 && fileInfos[0].Name() != filename {
		t.Errorf("file name is %q want %q", fileInfos[0].Name(), filename)
	}
}

func TestSetInfos(t *testing.T) {
	lagger := new(Logger)
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}
	lagger.SetInfos(d, log.Ldate, false)

	if lagger.Debug.Flags() != log.Ldate {
		t.Errorf("lagger.Debug.Flags=%q, want %q", lagger.Debug.Flags(), log.Ldate)
	}
	if lagger.Error.Flags() != log.LstdFlags {
		t.Errorf("lagger.Error.Flags=%q, want %q", lagger.Debug.Flags(), log.LstdFlags)
	}
}

func TestCanLogFile(t *testing.T) {
	lagger := new(Logger)
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}

	// check Error debug = false
	lagger.SetInfos(d, 0, false)
	contents := "Hello this is Error debug false"
	lagger.Error.Println(contents)
	if !strings.Contains(d.buffer.String(), contents) {
		t.Errorf("error log=%q want=\"<somedate> %v\"", d.buffer.String(), contents)
	}
	d.buffer.Reset()

	// check Error debug = true
	lagger.SetInfos(d, 0, true)
	contents = "Hello this is Error debug true"
	lagger.Error.Println(contents)
	if !strings.Contains(d.buffer.String(), contents) {
		t.Errorf("error log=%q want=\"<somedate> %v\"", d.buffer.String(), contents)
	}
	d.buffer.Reset()

	// check Debug debug = false
	lagger.SetInfos(d, 0, false)
	contents = "Hello this is Debug debug false"
	lagger.Debug.Println(contents)
	if !strings.Contains(d.buffer.String(), contents) {
		t.Errorf("error log=%q want=\"<somedate> %v\"", d.buffer.String(), contents)
	}
	d.buffer.Reset()

	// check Debug debug = true
	lagger.SetInfos(d, 0, true)
	contents = "Hello this is Debug debug true"
	lagger.Debug.Println(contents)
	if !strings.Contains(d.buffer.String(), contents) {
		t.Errorf("error log=%q want=\"<somedate> %v\"", d.buffer.String(), contents)
	}
	d.buffer.Reset()
}

func Example_logStderr() {
	// Now can not test stderr, thus I set stderr as stdout.
	cacheFile := os.Stderr
	os.Stderr = os.Stdout
	defer func() { os.Stderr = cacheFile }()

	lagger := new(Logger)
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}
	lagger.SetInfos(d, 0, false)

	// check Debug debug = false
	contents := "Hello this is Debug debug false"
	lagger.Debug.Print(contents)

	// check Debug debug = true
	lagger.SetInfos(d, 0, true)
	contents = "Hello this is Debug debug true"
	lagger.Debug.Print(contents)

	// Output:
	// Hello this is Debug debug true
}
func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
