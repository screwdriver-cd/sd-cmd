package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
)

type dummyStore struct{}

func (d *dummyStore) GetCommand() (*store.Command, error) {
	return dummyStoreCommand(validShell), nil
}

type dummyStoreBroken struct{}

func (d *dummyStoreBroken) GetCommand() (*store.Command, error) {
	return dummyStoreCommand(invalidShell), nil
}

type dummyStoreError struct{}

func (d *dummyStoreError) GetCommand() (*store.Command, error) {
	return dummyStoreCommand(validShell), fmt.Errorf("store cause error")
}

func dummyStoreCommand(body string) (cmd *store.Command) {
	return &store.Command{
		Type: "binary",
		Body: []byte(body),
		Spec: dummyAPICommand(binaryFormat),
	}
}

func TestNewBinary(t *testing.T) {
	_, err := NewBinary(dummyAPICommand(binaryFormat), []string{"arg1", "arg2"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestRun(t *testing.T) {
	logBuffer.Reset()

	// success with no arguments
	spec := dummyAPICommand(binaryFormat)
	bin, _ := NewBinary(spec, []string{})
	bin.Store = store.Store(new(dummyStore))
	err := bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	// if !strings.Contains(logBuffer.String(), "Hello World\n") {
	// 	t.Errorf("log is %q, should include %q", logBuffer.String(), "Hello World\n")
	// }
	// check file directory
	binPath := filepath.Join(config.BaseCommandPath, spec.Namespace, spec.Name, spec.Version, spec.Binary.File)
	fInfo, err := os.Stat(binPath)
	if os.IsNotExist(err) {
		t.Errorf("err=%q, file should exist at %q", binPath, err)
	}
	if fInfo.IsDir() {
		t.Errorf("%q is directory, must be file", binPath)
	}

	logBuffer.Reset()

	// success with arguments
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{"arg1", "arg2"})
	bin.Store = store.Store(new(dummyStore))
	err = bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	// if !strings.Contains(logBuffer.String(), "Hello World\n") {
	// 	t.Errorf("log is %q, should include %q", logBuffer.String(), "Hello World")
	// }
	// if !strings.Contains(logBuffer.String(), "arg1\n") {
	// 	t.Errorf("log is %q, should include %q", logBuffer.String(), "arg1\n")
	// }
	// if !strings.Contains(logBuffer.String(), "arg2\n") {
	// 	t.Errorf("log is %q, should include %q", logBuffer.String(), "arg2\n")
	// }
	logBuffer.Reset()

	// failure. the command is broken
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{})
	bin.Store = store.Store(new(dummyStoreBroken))
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. the store api return error
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{})
	bin.Store = store.Store(new(dummyStoreError))
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
