package executor

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/logger"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
	"github.com/screwdriver-cd/sd-cmd/util"
	"github.com/stretchr/testify/assert"
)

const (
	binaryFormat  = "binary"
	dockerFormat  = "docker"
	habitatFormat = "habitat"
)

const (
	dummyNameSpace   = "foo-dummy"
	dummyName        = "name-dummy"
	dummyVersion     = "1.0.1"
	dummyFileName    = "sd-step"
	dummyFile        = "/dummy/" + dummyFileName
	dummyEmptyFile   = "empty_file_path"
	dummyDescription = "dummy description"
	dummyMode        = "dummy_mode"
	dummyPackage     = "dummy_org/dummy"
	dummyHartName    = "dummy.hart"
	dummyHart        = "/dummy/" + dummyHartName
	dummyCommand     = "dummy_get"
	dummyImage       = "dummy:latest"
)

var (
	validShell   string
	invalidShell string
)

var (
	// logBuffer has log information
	logBuffer *bytes.Buffer
)

type dummyLogFile struct {
	buffer   *bytes.Buffer
	isClosed bool
}

func (d *dummyLogFile) Close() error {
	d.isClosed = true
	return nil
}

func (d *dummyLogFile) Write(p []byte) (n int, err error) {
	return d.buffer.Write(p)
}

func setup() {
	config.SDAPIURL = "http://fake.com/v4/"
	config.SDStoreURL = "http://fake.store/v1/"
	config.BaseCommandPath = filepath.Join(os.TempDir(), "sd")
	config.SDArtifactsDir = filepath.Join(os.TempDir(), "artifact")
	b, _ := ioutil.ReadFile("testdata/validShell.sh")
	validShell = string(b)
	b, _ = ioutil.ReadFile("testdata/invalidShell.sh")
	invalidShell = string(b)

	// setting lgr for logging
	logBuffer = bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: logBuffer}
	l, _ := logger.New(d)
	lgr = l
}

func teardown() {
	os.RemoveAll(config.BaseCommandPath)
	os.RemoveAll(config.SDArtifactsDir)
}

type dummySDAPI struct {
	spec *util.CommandSpec
	err  error
}

func (d *dummySDAPI) GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return d.spec, d.err
}

func (d *dummySDAPI) PostCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func (d *dummySDAPI) ValidateCommand(yamlString string) (*util.ValidateResponse, error) {
	return nil, nil
}

func (d *dummySDAPI) TagCommand(spec *util.CommandSpec, targetVersion, tag string) (*util.TagResponse, error) {
	return nil, nil
}

func (d *dummySDAPI) RemoveTagCommand(spec *util.CommandSpec, tag string) (*util.TagResponse, error) {
	return nil, nil
}

func (d *dummySDAPI) SetDebug(isDebug bool) {}

func newDummySDAPI(spec *util.CommandSpec, err error) api.API {
	d := &dummySDAPI{
		spec: spec,
		err:  err,
	}
	return api.API(d)
}

type dummyStore struct {
	body []byte
	spec *util.CommandSpec
	err  error
}

func newDummyStore(body string, spec *util.CommandSpec, err error) store.Store {
	ds := &dummyStore{
		body: []byte(body),
		spec: spec,
		err:  err,
	}
	return store.Store(ds)
}

func (d *dummyStore) GetCommand() (*store.Command, error) {
	storeCmd := &store.Command{
		Body: d.body,
		Spec: d.spec,
	}
	return storeCmd, d.err
}

func dummyCommandSpec(format string) (spec *util.CommandSpec) {
	spec = &util.CommandSpec{
		Namespace:   dummyNameSpace,
		Name:        dummyName,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      format,
	}

	switch format {
	case binaryFormat:
		spec.Binary = &util.Binary{
			File: dummyFile,
		}
	case habitatFormat:
		spec.Habitat = &util.Habitat{
			Mode:    dummyMode,
			File:    dummyHart,
			Package: dummyPackage,
			Command: dummyCommand,
		}
	case dockerFormat:
		spec.Docker = &util.Docker{
			Image:   dummyImage,
			Command: dummyCommand,
		}
	}
	return spec
}

func TestNew(t *testing.T) {
	successCases := []struct {
		name         string
		spec         *util.CommandSpec
		args         []string
		debugFromEnv bool
		isLogFile    bool
	}{
		{
			name:         "binary format succes with no logging with file",
			spec:         dummyCommandSpec(binaryFormat),
			args:         []string{"ns/cmd@ver"},
			debugFromEnv: false,
			isLogFile:    false,
		},
		{
			name:         "binary format success with no logging with file",
			spec:         dummyCommandSpec(habitatFormat),
			args:         []string{"exec", "ns/cmd@ver"},
			debugFromEnv: false,
			isLogFile:    false,
		},
		{
			name:         "should output log file by option",
			spec:         dummyCommandSpec(binaryFormat),
			args:         []string{"--debug", "ns/cmd@ver"},
			debugFromEnv: false,
			isLogFile:    true,
		},
		{
			name:         "should output log file by env",
			spec:         dummyCommandSpec(binaryFormat),
			args:         []string{"ns/cmd@ver"},
			debugFromEnv: true,
			isLogFile:    true,
		},
		{
			name:         "should output log file by option and env",
			spec:         dummyCommandSpec(binaryFormat),
			args:         []string{"--debug", "ns/cmd@ver"},
			debugFromEnv: true,
			isLogFile:    true,
		},
		{
			name:         "should not output log file",
			spec:         dummyCommandSpec(binaryFormat),
			args:         []string{"ns/cmd@ver", "--debug"},
			debugFromEnv: false,
			isLogFile:    false,
		},
	}
	for _, tt := range successCases {
		t.Run(tt.name, func(t *testing.T) {
			l := lgr
			d := config.DEBUG
			defer func() {
				lgr = l
				config.DEBUG = d
			}()
			config.DEBUG = tt.debugFromEnv
			sdapi := newDummySDAPI(tt.spec, nil)
			executor, err := New(sdapi, tt.args)
			assert.Nil(t, err)
			_, ok := executor.(Executor)
			assert.True(t, ok)
			if tt.isLogFile {
				assert.NotNil(t, lgr.File())
			} else {
				assert.Nil(t, lgr.File())
			}
		})
	}

	failureCases := []struct {
		name string
		spec *util.CommandSpec
		args []string
	}{
		{
			name: "failure with no command",
			spec: dummyCommandSpec(binaryFormat),
			args: []string{},
		},
		{
			name: "failure with invalid command",
			spec: dummyCommandSpec(binaryFormat),
			args: []string{"sd-cmd", "ns@cmd/ver"},
		},
		{
			name: "failure with screwdriver API error",
			spec: dummyCommandSpec("Unknown"),
			args: []string{"sd-cmd", "ns/cmd@ver"},
		},
	}

	for _, tt := range failureCases {
		t.Run(tt.name, func(t *testing.T) {
			sdapi := newDummySDAPI(tt.spec, nil)
			_, err := New(sdapi, tt.args)
			assert.NotNil(t, err)
		})
	}
}

func TestCleanUp(t *testing.T) {
	l := lgr
	defer func() {
		lgr = l
	}()

	file := &dummyLogFile{buffer: bytes.NewBuffer([]byte{})}
	lgr, _ = logger.New(file)
	CleanUp()
	assert.Equal(t, true, file.isClosed)
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
