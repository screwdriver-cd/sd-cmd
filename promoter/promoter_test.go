package promoter

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
	"github.com/stretchr/testify/assert"
)

var (
	dummyNameSpace      = "foo-dummy"
	dummyName           = "name-dummy"
	dummyVersion        = "1.0.0"
	dummyTargetVersion  = "1.0.1"
	dummyCmdName        = dummyNameSpace + "/" + dummyName
	invalidDummyCmdName = "invalid/invalid/invalid"
	dummyTag            = "stable"
	invalidDummyTag     = "-invalid-"
)

type dummySDAPI struct{}

func (d *dummySDAPI) GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return &util.CommandSpec{
		Version: dummyVersion,
	}, nil
}

func (d *dummySDAPI) PostCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func (d *dummySDAPI) ValidateCommand(yamlString string) (*util.ValidateResponse, error) {
	return nil, nil
}

func (d *dummySDAPI) TagCommand(spec *util.CommandSpec, targetVersion, tag string) (*util.TagResponse, error) {
	return &util.TagResponse{
		Namespace: dummyNameSpace,
		Name:      dummyName,
		Tag:       dummyTag,
		Version:   dummyTargetVersion,
	}, nil
}

type dummyInvalidSDAPI struct{}

func (d *dummyInvalidSDAPI) GetCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return &util.CommandSpec{
		Version: dummyVersion,
	}, nil
}

func (d *dummyInvalidSDAPI) PostCommand(smallSpec *util.CommandSpec) (*util.CommandSpec, error) {
	return nil, nil
}

func (d *dummyInvalidSDAPI) ValidateCommand(yamlString string) (*util.ValidateResponse, error) {
	return nil, nil
}

func (d *dummyInvalidSDAPI) TagCommand(spec *util.CommandSpec, targetVersion, tag string) (*util.TagResponse, error) {
	return nil, errors.New("error")
}

func TestNew(t *testing.T) {
	sdapi := api.API(new(dummySDAPI))

	// success
	expected := &Promoter{
		smallSpec: &util.CommandSpec{
			Namespace: dummyNameSpace,
			Name:      dummyName,
			Version:   dummyTag,
		},
		sdAPI:         sdapi,
		targetVersion: dummyTargetVersion,
		tag:           dummyTag,
	}
	p, err := New(sdapi, []string{dummyCmdName, dummyTargetVersion, dummyTag})
	assert.Nil(t, err)
	assert.Equal(t, p, expected)

	// failure with no args
	_, err = New(sdapi, []string{})
	assert.EqualError(t, err, "parameters are not enough")

	// failure with invalid command name
	_, err = New(sdapi, []string{invalidDummyCmdName, dummyTargetVersion, dummyTag})
	assert.EqualError(t, err, invalidDummyCmdName+" is invalid command name")

	// failure with invalid tag name
	_, err = New(sdapi, []string{dummyCmdName, dummyTargetVersion, invalidDummyTag})
	assert.EqualError(t, err, invalidDummyTag+" is invalid tag name")
}

func TestRun(t *testing.T) {
	// success
	sdapi := api.API(new(dummySDAPI))
	p, err := New(sdapi, []string{dummyCmdName, dummyTargetVersion, dummyTag})
	if err != nil {
		assert.Fail(t, "err should be nil")
	}
	err = p.Run()
	assert.Nil(t, err)

	// success with already tagged version
	dummyTargetVersion = "1.0.0"
	p, err = New(sdapi, []string{dummyCmdName, dummyTargetVersion, dummyTag})
	if err != nil {
		assert.Fail(t, "err should be nil")
	}
	err = p.Run()
	assert.Nil(t, err)

	// failure with error response from TagCommand
	dummyTargetVersion = "1.0.1"
	sdapi = api.API(new(dummyInvalidSDAPI))
	p, err = New(sdapi, []string{dummyCmdName, dummyTargetVersion, dummyTag})
	if err != nil {
		assert.Fail(t, "err should be nil")
	}
	err = p.Run()
	assert.EqualError(t, err, "error")
}
