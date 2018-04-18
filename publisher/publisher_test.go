package publisher

import (
	"fmt"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

type dummySDAPIBinary struct{}

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
	validSpecYamlPath   = "../testdata/yaml/binary-sd-command.yaml"
	invalidSpecYamlPath = "../testdata/yaml/invalid_sd-command.yaml"
)

type dummySDAPI struct {
	spec *util.CommandSpec
	err  error
}

func (d *dummySDAPI) GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return d.spec, d.err
}

func (d *dummySDAPI) PostCommand(specPath string, smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return d.spec, d.err
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
	_, err := New(sdapi, []string{"-f", validSpecYamlPath})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// failure. invalid flag
	spec = dummySpec(binaryFormat)
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{"-x", "invalid_flag"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. invalid yaml file
	spec = dummySpec(binaryFormat)
	sdapi = newDummySDAPI(spec, nil)
	_, err = New(sdapi, []string{"-f", invalidSpecYamlPath})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}

func TestRun(t *testing.T) {
	spec := dummySpec(binaryFormat)
	sdapi := newDummySDAPI(spec, nil)
	pub, err := New(sdapi, []string{"-f", validSpecYamlPath})
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
	err = pub.Run()
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}

	// failure. failed to post command
	spec = dummySpec(binaryFormat)
	sdapi = newDummySDAPI(spec, fmt.Errorf("failed to post command"))
	pub, _ = New(sdapi, []string{"-f", validSpecYamlPath})
	err = pub.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
