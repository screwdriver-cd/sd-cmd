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
	buffer *bytes.Buffer
}

func (d *dummyLogFile) Close() error { return nil }
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
	l := new(logger.Logger)
	logBuffer = bytes.NewBuffer([]byte{})
	d := &dummyLogFile{buffer: logBuffer}
	l.SetInfos(d, 0, true)
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

func (d *dummySDAPI) TagCommand(spec *util.CommandSpec, targetVersion, tag string) error {
	return nil
}

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
	// case binary format
	spec := dummyCommandSpec(binaryFormat)
	sdapi := newDummySDAPI(spec, nil)
	// success
	executor, err := New(sdapi, []string{"ns/cmd@ver"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if _, ok := executor.(Executor); !ok {
		t.Errorf("New does not fulfill API interface")
	}

	// failure. no command
	_, err = New(sdapi, []string{})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. invalid command
	_, err = New(sdapi, []string{"sd-cmd", "ns@cmd/ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// case habitat format
	spec = dummyCommandSpec(habitatFormat)
	sdapi = newDummySDAPI(spec, nil)
	// success
	executor, err = New(sdapi, []string{"exec", "ns/cmd@ver"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if _, ok := executor.(Executor); !ok {
		t.Errorf("New does not fulfill API interface")
	}

	// failure. Screwdriver API error
	spec = dummyCommandSpec("Unknown")
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{"sd-cmd", "ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
