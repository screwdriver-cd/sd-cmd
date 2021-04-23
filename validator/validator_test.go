package validator

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

type dummySDAPIValidator struct{}

func (d *dummySDAPIValidator) GetCommand(spec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func (d *dummySDAPIValidator) PostCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func dummyAPIValidateCommand() (res *util.ValidateResponse) {
	res = &util.ValidateResponse{
		Errors: []util.ValidateError{},
	}
	return res
}

func (d *dummySDAPIValidator) ValidateCommand(yamlString string) (*util.ValidateResponse, error) {
	return dummyAPIValidateCommand(), nil
}

func (d *dummySDAPIValidator) TagCommand(spec *util.CommandSpec, targetVersion, tag string) (*util.TagResponse, error) {
	return nil, nil
}

func (d *dummySDAPIValidator) RemoveTagCommand(spec *util.CommandSpec, tag string) (*util.TagResponse, error) {
	return nil, nil
}

func TestNew(t *testing.T) {
	// success
	testDataPath := "../testdata/yaml/sd-command.yaml"
	sdapi := api.API(new(dummySDAPIValidator))
	_, err := New(sdapi, []string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestRun(t *testing.T) {
	testDataPath := "../testdata/yaml/sd-command.yaml"
	sdapi := api.API(new(dummySDAPIValidator))
	pub, err := New(sdapi, []string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
	err = pub.Run()
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
}
