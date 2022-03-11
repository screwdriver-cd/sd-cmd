package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwdriver-cd/sd-cmd/config"
)

func TestNewBinary(t *testing.T) {
	_, err := NewBinary(dummyCommandSpec(binaryFormat), []string{"arg1", "arg2"}, false)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestGetBinDirPath(t *testing.T) {
	spec := dummyCommandSpec(binaryFormat)
	bin, _ := NewBinary(spec, []string{}, false)
	bin.Store = newDummyStore(validShell, spec, nil)
	// Note: config.BaseCommandPath is customized for test.
	// see executor/executor_test.go
	assert.Equal(t, bin.getBinDirPath(), filepath.Join(config.BaseCommandPath, "foo-dummy/name-dummy/1.0.1"))
}

func TestGetBinFilePath(t *testing.T) {
	spec := dummyCommandSpec(binaryFormat)
	bin, _ := NewBinary(spec, []string{}, false)
	bin.Store = newDummyStore(validShell, spec, nil)
	assert.Equal(t, bin.getBinFilePath(), filepath.Join(config.BaseCommandPath, "foo-dummy/name-dummy/1.0.1/sd-step"))
}

func TestIsInstalled(t *testing.T) {
	spec := dummyCommandSpec(binaryFormat)
	bin, _ := NewBinary(spec, []string{}, false)
	bin.Store = newDummyStore(validShell, spec, nil)
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

	spec := dummyCommandSpec(binaryFormat)
	// success with no arguments
	bin, _ := NewBinary(spec, []string{}, false)
	bin.Store = newDummyStore(validShell, spec, nil)
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

	// success binary.file is relative path
	spec.Binary.File = "./sample/relative_path"
	bin, _ = NewBinary(spec, []string{}, false)
	bin.Store = newDummyStore(validShell, spec, nil)
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
	assert.True(t, bin.isInstalled())
	os.Remove(binPath)

	// success with arguments
	bin, _ = NewBinary(spec, []string{"arg1", "arg2"}, false)
	bin.Store = newDummyStore(validShell, spec, nil)
	err = bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	assert.True(t, bin.isInstalled())
	os.Remove(binPath)

	// failure. the command is broken
	bin, _ = NewBinary(dummyCommandSpec(binaryFormat), []string{}, false)
	bin.Store = newDummyStore(invalidShell, spec, nil)
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
	os.Remove(binPath)

	// failure. the store api return error
	bin, _ = NewBinary(spec, []string{}, false)
	bin.Store = newDummyStore(validShell, spec, fmt.Errorf("store cause error"))
	err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
	os.Remove(binPath)
}
