package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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

func TestGetBinDirPath(t *testing.T) {
	spec := dummyAPICommand(binaryFormat)
	bin, _ := NewBinary(spec, []string{})
	bin.Store = store.Store(new(dummyStore))
	assert.Equal(t, bin.getBinDirPath(), "/tmp/sd/foo-dummy/name-dummy/1.0.1")
}

func TestGetBinFilePath(t *testing.T) {
	spec := dummyAPICommand(binaryFormat)
	bin, _ := NewBinary(spec, []string{})
	bin.Store = store.Store(new(dummyStore))
	assert.Equal(t, bin.getBinFilePath(), "/tmp/sd/foo-dummy/name-dummy/1.0.1/sd-step")
}

func TestIsInstalled(t *testing.T) {
	spec := dummyAPICommand(binaryFormat)
	bin, _ := NewBinary(spec, []string{})
	bin.Store = store.Store(new(dummyStore))
	// Not exists
	assert.False(t, bin.isInstalled())

	// 0 size file
	os.MkdirAll(bin.getBinDirPath(), 0777)
	file, _ := os.Create(bin.getBinFilePath())
	assert.False(t, bin.isInstalled())

	// non 0 size file
	file.Write(([]byte)("dummy script."))
	assert.True(t, bin.isInstalled())

	defer file.Close()
	defer os.RemoveAll(bin.getBinDirPath())
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

	// check file directory
	binPath := filepath.Join(config.BaseCommandPath, spec.Namespace, spec.Name, spec.Version, dummyFileName)
	fInfo, err := os.Stat(binPath)
	if os.IsNotExist(err) {
		t.Errorf("err=%q, file should exist at %q", binPath, err)
	}
	if fInfo.IsDir() {
		t.Errorf("%q is directory, must be file", binPath)
	}
	assert.True(t, bin.isInstalled())
	os.Remove(binPath)

	// success with arguments
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{"arg1", "arg2"})
	bin.Store = store.Store(new(dummyStore))
	err = bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	assert.True(t, bin.isInstalled())
	os.Remove(binPath)

	// failure. the command is broken
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{})
	bin.Store = store.Store(new(dummyStoreBroken))
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
	os.Remove(binPath)

	// failure. the store api return error
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{})
	bin.Store = store.Store(new(dummyStoreError))
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
	os.Remove(binPath)
}
