package validator

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

type dummySDAPIValidateCommand struct{}

func (d *dummySDAPIValidateCommand) GetCommand(spec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func (d *dummySDAPIValidateCommand) PostCommand(specPath string, smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func dummyAPIValidateCommand() (res *util.ValidateResponse) {
	res = &util.ValidateResponse{
		Errors: []util.ValidateError{},
	}
	return res
}

func (d *dummySDAPIValidateCommand) ValidateCommand(yamlString string) (*util.ValidateResponse, error) {
	return dummyAPIValidateCommand(), nil
}

func TestNew(t *testing.T) {
	// success
	testDataPath := "../testdata/yaml/sd-command.yaml"
	sdapi := api.API(new(dummySDAPIValidateCommand))
	_, err := New(sdapi, []string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestRun(t *testing.T) {
	testDataPath := "../testdata/yaml/sd-command.yaml"
	sdapi := api.API(new(dummySDAPIValidateCommand))
	pub, err := New(sdapi, []string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
	err = pub.Run()
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
}
