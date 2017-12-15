package executor

import (
	"fmt"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

const (
	binaryFormat  = "binary"
	dockerFormat  = "docker"
	habitatFormat = "habitat"
)

const (
	dummyNameSpace   = "foo-dummy"
	dummyCommand     = "cmd-dummy"
	dummyVersion     = "1.0.1"
	dummyFile        = "sd-step"
	dummyDescription = "dummy description"
)

type dummySDAPIBinary struct{}

func (d *dummySDAPIBinary) SetJWT() error { return nil }

func (d *dummySDAPIBinary) GetCommand(namespace, command, version string) (*api.Command, error) {
	return dummySDCommand(binaryFormat), nil
}

type dummySDAPIBroken struct{}

func (d *dummySDAPIBroken) SetJWT() error { return nil }
func (d *dummySDAPIBroken) GetCommand(namespace, command, version string) (*api.Command, error) {
	return nil, fmt.Errorf("Something error happen")
}

func dummySDCommand(format string) (cmd *api.Command) {
	cmd = &api.Command{
		Namespace:   dummyNameSpace,
		Command:     dummyCommand,
		Description: dummyDescription,
		Version:     dummyVersion,
		Format:      format,
	}
	cmd.Binary.File = dummyFile
	return cmd
}

func TestNew(t *testing.T) {
	// success
	sdapi := api.API(new(dummySDAPIBinary))
	executor, err := New(sdapi, []string{"ns/cmd@ver"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if _, ok := executor.(Executor); !ok {
		t.Errorf("New does not fulfill API interface")
	}

	// failure. no command
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. invalid command
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{"ns@cmd/ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. invalid command
	sdapi = api.API(new(dummySDAPIBinary))
	_, err = New(sdapi, []string{"ns-cmd-ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. Screwdriver API error
	sdapi = api.API(new(dummySDAPIBroken))
	_, err = New(sdapi, []string{"ns/cmd@ver"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
