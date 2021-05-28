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
	"github.com/stretchr/testify/assert"
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

// TODO CreateLogFile,SetInfo should be private
// TODO executor_test should use New
func TestNew(t *testing.T) {
	// success
	dir := filepath.Join(tempDir, "CreateLogFile")
	filename := fmt.Sprintf("logger_test_%v", time.Now().Unix())

	_, err := New(OutputToFileWithCreate(dir, filename), DebugFlag(log.Ldate), OutputDebugLog())
	defer os.RemoveAll(dir)
	assert.Nil(t, err)

	_, err = os.Stat(dir)
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
	lgr := new(Logger)
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}
	lgr.setInfos(d, log.Ldate, false)

	if lgr.Debug.Flags() != log.Ldate {
		t.Errorf("lgr.Debug.Flags=%q, want %q", lgr.Debug.Flags(), log.Ldate)
	}
	if lgr.Error.Flags() != log.LstdFlags {
		t.Errorf("lgr.Error.Flags=%q, want %q", lgr.Debug.Flags(), log.LstdFlags)
	}
}

func TestCanLogFile(t *testing.T) {
	lgr := new(Logger)
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}

	// check Error debug = false
	lgr.setInfos(d, 0, false)
	contents := "Hello this is Error debug false"
	lgr.Error.Println(contents)
	if !strings.Contains(d.buffer.String(), contents) {
		t.Errorf("error log=%q want=\"<somedate> %v\"", d.buffer.String(), contents)
	}
	d.buffer.Reset()

	// check Error debug = true
	lgr.setInfos(d, 0, true)
	contents = "Hello this is Error debug true"
	lgr.Error.Println(contents)
	if !strings.Contains(d.buffer.String(), contents) {
		t.Errorf("error log=%q want=\"<somedate> %v\"", d.buffer.String(), contents)
	}
	d.buffer.Reset()

	// check Debug debug = false
	lgr.setInfos(d, 0, false)
	contents = "Hello this is Debug debug false"
	lgr.Debug.Println(contents)
	if !strings.Contains(d.buffer.String(), contents) {
		t.Errorf("error log=%q want=\"<somedate> %v\"", d.buffer.String(), contents)
	}
	d.buffer.Reset()

	// check Debug debug = true
	lgr.setInfos(d, 0, true)
	contents = "Hello this is Debug debug true"
	lgr.Debug.Println(contents)
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

	lgr := new(Logger)
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}
	lgr.setInfos(d, 0, false)

	// check Debug debug = false
	contents := "Hello this is Debug debug false"
	lgr.Debug.Print(contents)

	// check Debug debug = true
	lgr.setInfos(d, 0, true)
	contents = "Hello this is Debug debug true"
	lgr.Debug.Print(contents)

	// Output:
	// Hello this is Debug debug true
}
func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
