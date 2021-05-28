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
	cases := []struct {
		name             string
		debugFlag        int
		isOutputDebugLog bool
	}{
		{
			name:             "debugFlag: log.Ldate, isOutputDebugLog: true",
			debugFlag:        log.Ldate,
			isOutputDebugLog: true,
		},
		{
			name:             "debugFlag: 0, isOutputDebugLog: false",
			debugFlag:        0,
			isOutputDebugLog: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			buffer := bytes.NewBuffer([]byte{})
			d := &dummyLogFile{buffer: buffer}
			lgr, err := New(OutputToFile(d), DebugFlag(tt.debugFlag), OutputDebugLog(tt.isOutputDebugLog))

			assert.Nil(t, err)
			assert.Equal(t, tt.debugFlag, lgr.debugFlag)
			assert.Equal(t, tt.debugFlag, lgr.Debug.Flags())
			assert.Equal(t, tt.isOutputDebugLog, lgr.isOutputDebugLog)
			assert.Equal(t, log.LstdFlags, lgr.Error.Flags())
		})
	}
}

func TestOutputToFileWithCreate(t *testing.T) {
	dir := filepath.Join(tempDir, "CreateLogFile")
	filename := fmt.Sprintf("logger_test_%v", time.Now().Unix())

	lgr, err := New(OutputToFileWithCreate(dir, filename), DebugFlag(log.Ldate))
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	assert.Equal(t, log.Ldate, lgr.debugFlag)
	assert.Equal(t, false, lgr.isOutputDebugLog)

	_, err = os.Stat(dir)
	assert.Nil(t, err)

	// check there is a file
	fileInfos, _ := ioutil.ReadDir(dir)
	assert.Equal(t, fileInfos[0].Name(), filename)
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
