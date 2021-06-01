package logger

import (
	"bytes"
	"fmt"
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

func TestNew(t *testing.T) {
	buffer := bytes.NewBuffer([]byte{})
	dummyFile := &dummyLogFile{buffer: buffer}

	cases := []struct {
		name      string
		debugFlag int
		isDebug   bool
	}{
		{
			name:      "debugFlag: log.Ldate, isOutputDebugLog: true",
			debugFlag: log.Ldate,
			isDebug:   true,
		},
		{
			name:      "debugFlag: 0, isOutputDebugLog: false",
			debugFlag: 0,
			isDebug:   false,
		},
		{
			name: "file is nil",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			lgr, err := New(OptDebugFlag(tt.debugFlag))
			if tt.isDebug {
				lgr, err = New(OptDebug(dummyFile), OptDebugFlag(tt.debugFlag))
				assert.Equal(t, dummyFile, lgr.file)
			}

			assert.Nil(t, err)
			assert.Equal(t, tt.debugFlag, lgr.debugFlag)
			assert.Equal(t, tt.debugFlag, lgr.Debug.Flags())
			assert.Equal(t, tt.isDebug, lgr.isDebug)
		})
	}

	t.Run("default value", func(t *testing.T) {
		lgr, err := New()
		assert.Nil(t, err)
		assert.Equal(t, log.LstdFlags, lgr.Error.Flags())
		assert.Equal(t, log.LstdFlags, lgr.Debug.Flags())
		assert.Nil(t, lgr.File())
		assert.False(t, lgr.isDebug)
	})
}

func TestOutputToFileWithCreate(t *testing.T) {
	dir := filepath.Join(tempDir, "CreateLogFile")
	filename := fmt.Sprintf("logger_test_%v", time.Now().Unix())

	lgr, err := New(OptDebugWithCreate(dir, filename), OptDebugFlag(log.Ldate))
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	assert.Equal(t, log.Ldate, lgr.debugFlag)
	assert.Equal(t, false, lgr.isDebug)

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
		options    []LogOption
		logMessage string
		expect     string
	}{
		{
			name:       "OutputToFile: true, OutputDebugLog: false",
			options:    []LogOption{OptDebug(d), OptErrorFlag(0)},
			logMessage: "Hello",
			expect:     "Hello\n",
		},
		{
			name:       "OutputToFile: true, OutputDebugLog: true",
			options:    []LogOption{OptDebug(d), OptErrorFlag(0)},
			logMessage: "Hello",
			expect:     "Hello\n",
		},
		{
			name:       "OutputToFile: false, OutputDebugLog: false",
			options:    []LogOption{OptErrorFlag(0)},
			logMessage: "Hello",
			expect:     "",
		},
		{
			name:       "OutputToFile: false, OutputDebugLog: true",
			options:    []LogOption{OptErrorFlag(0)},
			logMessage: "Hello",
			expect:     "",
		},
	}
	for _, tt := range errorCases {
		t.Run(tt.name, func(t *testing.T) {
			defer d.buffer.Reset()
			l, _ := New(tt.options...)
			l.Error.Println(tt.logMessage)
			assert.Equal(t, tt.expect, d.buffer.String())
		})
	}

	debugCases := []struct {
		name       string
		options    []LogOption
		logMessage string
		expect     string
	}{
		{
			name:       "OptDebug: true",
			options:    []LogOption{OptDebug(d), OptDebugFlag(0)},
			logMessage: "Hello",
			expect:     "Hello\n",
		},
		{
			name:       "OptDebug: false",
			options:    []LogOption{OptDebugFlag(0)},
			logMessage: "Hello",
			expect:     "",
		},
	}
	for _, tt := range debugCases {
		t.Run(tt.name, func(t *testing.T) {
			defer d.buffer.Reset()
			l, _ := New(tt.options...)
			l.Debug.Println(tt.logMessage)
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
	lgr, _ := New(OptDebug(d), OptDebugFlag(0))
	contents := "Hello this is Debug debug false"
	lgr.Debug.Println(contents)

	// check Debug debug = true
	lgr, _ = New(OptDebug(d), OptDebugFlag(0))
	contents = "Hello this is Debug debug true"
	lgr.Debug.Println(contents)

	lgr, _ = New(OptDebugFlag(0))
	contents = "Hello this is Debug debug falase with no file"
	lgr.Debug.Println(contents)

	lgr, _ = New(OptDebugFlag(0))
	contents = "Hello this is Debug debug falase with file"
	lgr.Debug.Println(contents)

	lgr, _ = New(OptDebug(d), OptErrorFlag(0))
	contents = "Hello this is Error with file"
	lgr.Error.Println(contents)

	lgr, _ = New(OptDebug(d), OptErrorFlag(0))
	contents = "Hello this is Error with debug false"
	lgr.Error.Println(contents)

	lgr, _ = New(OptErrorFlag(0))
	contents = "Hello this is Error with no file"
	lgr.Error.Println(contents)

	// Output:
	// Hello this is Error with file
	// Hello this is Error with debug false
	// Hello this is Error with no file
}
func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
