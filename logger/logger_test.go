package logger

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

func newLogger(file io.WriteCloser) (lgr Logger) {
	lgr = Logger{file: file, debugFlag: 0, errorFlag: 0}
	lgr.buildLoggers()
	return
}

func TestNew(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})
	dummyFile := &dummyLogFile{buffer: buffer}

	cases := []struct {
		name    string
		isDebug bool
	}{
		{
			name:    "debugFlag: log.Ldate, isOutputDebugLog: true",
			isDebug: true,
		},
		{
			name:    "debugFlag: 0, isOutputDebugLog: false",
			isDebug: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			lgr, err := New()
			if tt.isDebug {
				lgr, err = New(OptDebug(dummyFile))
				assert.Equal(t, dummyFile, lgr.file)
			}
			assert.Nil(t, err)
		})
	}

	t.Run("default value", func(t *testing.T) {
		lgr, err := New()
		assert.Nil(t, err)
		assert.Equal(t, log.LstdFlags, lgr.Error.Flags())
		assert.Equal(t, log.LstdFlags, lgr.Debug.Flags())
		assert.Nil(t, lgr.File())
	})
}

// TODO should check able to write file
func TestOutputToFileWithCreate(t *testing.T) {
	dir := filepath.Join(tempDir, "CreateLogFile")
	filename := fmt.Sprintf("logger_test_%v", time.Now().Unix())

	lgr, err := New(OptDebugWithCreate(dir, filename))
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	assert.NotNil(t, lgr.File())

	_, err = os.Stat(dir)
	assert.Nil(t, err)

	// check there is a file
	fileInfos, _ := ioutil.ReadDir(dir)
	assert.Equal(t, fileInfos[0].Name(), filename)
}

func TestCanLogFile(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}

	errorCases := []struct {
		name       string
		lgr        Logger
		logMessage string
		expect     string
	}{
		{
			name:       "debug: true",
			lgr:        newLogger(d),
			logMessage: "Hello",
			expect:     "Hello\n",
		},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			defer d.buffer.Reset()
			tt.lgr.Error.Println(tt.logMessage)
			assert.Equal(t, tt.expect, d.buffer.String())
		})
	}

	debugCases := []struct {
		name       string
		lgr        Logger
		logMessage string
		expect     string
	}{
		{
			name:       "debug true",
			lgr:        newLogger(d),
			logMessage: "Hello",
			expect:     "Hello\n",
		},
	}
	for _, tt := range debugCases {
		t.Run(tt.name, func(t *testing.T) {
			defer d.buffer.Reset()
			tt.lgr.Debug.Println(tt.logMessage)
			assert.Equal(t, tt.expect, d.buffer.String())
		})
	}
}

func Example_logStderr() {
	// Now can not test stderr, thus I set stderr as stdout.
	cacheFile := os.Stderr
	os.Stderr = os.Stdout
	defer func() { os.Stderr = cacheFile }()
	buffer := bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: buffer}

	// check Debug debug = false
	lgr := newLogger(d)
	contents := "Hello this is Debug debug false"
	lgr.Debug.Println(contents)
	contents = "Hello this is Error with file"
	lgr.Error.Println(contents)

	lgr = newLogger(nil)
	contents = "Hello this is Debug debug falase with no file"
	lgr.Debug.Println(contents)
	contents = "Hello this is Error with no file"
	lgr.Error.Println(contents)

	// Output:
	// Hello this is Error with file
	// Hello this is Error with no file
}
func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
