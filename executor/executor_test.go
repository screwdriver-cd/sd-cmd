package executor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/logger"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

const (
	binaryFormat  = "binary"
	dockerFormat  = "docker"
	habitatFormat = "habitat"
)

const (
	dummyNameSpace      = "foo-dummy"
	dummyName           = "name-dummy"
	dummyVersion        = "1.0.1"
	dummyDescription    = "dummy description"
	dummyBinaryFileName = "sd-step"
	dummyBinaryFile     = "/dummy/" + dummyBinaryFileName
	dummyDockerImage    = "chefdk:1.2.3"
	dummyDockerCmd      = "knife"
	dummyHabitatMode    = "remote"
	dummyHabitatPkg     = "core/git/2.14.1"
	dummyHabitatCmd     = "git"
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

func (d *dummySDAPI) PostCommand(specPath string, smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func newDummySDAPI(spec *util.CommandSpec, err error) api.API {
	d := &dummySDAPI{
		spec: spec,
		err:  err,
	}
	return api.API(d)
}

func dummySpec(format string) (cmd *util.CommandSpec) {
	cmd = &util.CommandSpec{
		Namespace:   dummyNameSpace,
		Name:        dummyName,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      format,
	}
	switch format {
	case binaryFormat:
		cmd.Binary = new(util.Binary)
		cmd.Binary.File = dummyBinaryFile
	case dockerFormat:
		cmd.Docker = new(util.Docker)
		cmd.Docker.Command = dummyDockerCmd
		cmd.Docker.Image = dummyDockerImage
	case habitatFormat:
		cmd.Habitat = new(util.Habitat)
		cmd.Habitat.Command = dummyHabitatCmd
		cmd.Habitat.Mode = dummyHabitatMode
		cmd.Habitat.Package = dummyHabitatPkg
	}
	return cmd
}

func TestNew(t *testing.T) {
	// success
	spec := dummySpec(binaryFormat)
	sdapi := newDummySDAPI(spec, nil)
	executor, err := New(sdapi, []string{"ns/cmd@ver"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if _, ok := executor.(Executor); !ok {
		t.Errorf("New does not fulfill API interface")
	}

	// success binary mode
	spec = dummySpec(binaryFormat)
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{"exec", "ns/cmd@ver"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// failure. no command
	spec = dummySpec(binaryFormat)
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. invalid command
	spec = dummySpec(binaryFormat)
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{"sd-cmd", "ns@cmd/ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. Screwdriver API error
	spec = dummySpec(binaryFormat)
	sdapi = newDummySDAPI(spec, fmt.Errorf("Something error happen"))
	_, err = New(sdapi, []string{"sd-cmd", "ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. habitat type(not implemented yet)
	spec = dummySpec(habitatFormat)
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{"sd-cmd", "ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. docker type(not implemented yet)
	spec = dummySpec(dockerFormat)
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{"sd-cmd", "ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. Unknown type
	spec = dummySpec("Unknown")
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
