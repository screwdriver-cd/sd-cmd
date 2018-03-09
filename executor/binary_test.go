package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
	"github.com/screwdriver-cd/sd-cmd/util"
)

type dummyStore struct {
	cmdType string
	body    []byte
	spec    *util.CommandSpec
	err     error
}

func newDummyStore(cmdType string, body string, spec *util.CommandSpec, err error) store.Store {
	ds := &dummyStore{
		cmdType: cmdType,
		body:    []byte(body),
		spec:    spec,
		err:     err,
	}
	return store.Store(ds)
}

func (d *dummyStore) GetCommand() (*store.Command, error) {
	storeCmd := &store.Command{
		Type: d.cmdType,
		Body: d.body,
		Spec: d.spec,
	}
	return storeCmd, d.err
}

func TestNewBinary(t *testing.T) {
	_, err := NewBinary(dummySpec(binaryFormat), []string{"arg1", "arg2"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestRun(t *testing.T) {
	logBuffer.Reset()

	// success with no arguments
	spec := dummySpec(binaryFormat)
	bin, _ := NewBinary(spec, []string{})
	bin.Store = newDummyStore(binaryFormat, validShell, spec, nil)
	err := bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// check file directory
	binPath := filepath.Join(config.BaseCommandPath, spec.Namespace, spec.Name, spec.Version, dummyBinaryFileName)
	fInfo, err := os.Stat(binPath)
	if os.IsNotExist(err) {
		t.Errorf("err=%q, file should exist at %q", binPath, err)
	}
	if fInfo.IsDir() {
		t.Errorf("%q is directory, must be file", binPath)
	}

	// success with arguments
	spec = dummySpec(binaryFormat)
	bin, _ = NewBinary(spec, []string{"arg1", "arg2"})
	bin.Store = newDummyStore(binaryFormat, validShell, spec, nil)
	err = bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// success binary.file is relative path
	spec = dummySpec(binaryFormat)
	spec.Binary.File = "./sample/relative_path"
	bin, _ = NewBinary(spec, []string{})
	bin.Store = newDummyStore(binaryFormat, validShell, spec, nil)
	err = bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	binPath = filepath.Join(config.BaseCommandPath, spec.Namespace, spec.Name, spec.Version, "relative_path")
	fInfo, err = os.Stat(binPath)
	if os.IsNotExist(err) {
		t.Errorf("err=%q, file should exist at %q", binPath, err)
	}
	if fInfo.IsDir() {
		t.Errorf("%q is directory, must be file", binPath)
	}

	// failure. the command is broken
	spec = dummySpec(binaryFormat)
	bin, _ = NewBinary(spec, []string{})
	bin.Store = newDummyStore(binaryFormat, invalidShell, spec, nil)
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. the store api return error
	spec = dummySpec(binaryFormat)
	bin, _ = NewBinary(spec, []string{})
	bin.Store = newDummyStore(binaryFormat, validShell, spec, fmt.Errorf("store cause error"))
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
