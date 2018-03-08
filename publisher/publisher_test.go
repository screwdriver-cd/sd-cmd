package publisher

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

type dummySDAPIBinary struct{}

const (
	dummyNameSpace   = "foo-dummy"
	dummyName        = "name-dummy"
	dummyVersion     = "1.0.1"
	dummyFile        = "sd-step"
	dummyDescription = "dummy description"
)

func dummyAPICommand(format string) (cmd *util.CommandSpec) {
	cmd = &util.CommandSpec{
		Namespace:   dummyNameSpace,
		Name:        dummyName,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      format,
	}
	cmd.Binary = new(util.Binary)
	cmd.Binary.File = dummyFile
	return cmd
}

func (d *dummySDAPIBinary) GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return dummyAPICommand("binary"), nil
}

func (d *dummySDAPIBinary) PostCommand(specPath string, smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return dummyAPICommand("binary"), nil
}

func TestNew(t *testing.T) {
	// success
	testDataPath := "../testdata/yaml/sd-command.yaml"
	sdapi := api.API(new(dummySDAPIBinary))
	_, err := New(sdapi, []string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestRun(t *testing.T) {
	testDataPath := "../testdata/yaml/sd-command.yaml"
	sdapi := api.API(new(dummySDAPIBinary))
	pub, err := New(sdapi, []string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
	err = pub.Run()
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
}
