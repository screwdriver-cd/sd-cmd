package executor

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
)

var (
	validShell   string
	invalidShell string
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
		Meta: dummySDCommand(binaryFormat),
	}
}

func TestNewBinary(t *testing.T) {
	bin, err := NewBinary(dummySDCommand(binaryFormat), store.Store(new(dummyStore)), []string{"arg1", "arg2"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if !reflect.DeepEqual(bin.APICommand, dummySDCommand(binaryFormat)) {
		t.Errorf("bin.Cmd=%q, want %q", bin.APICommand, dummySDCommand(binaryFormat))
	}
}

func TestRun(t *testing.T) {
	// success with no arguments
	bin, _ := NewBinary(dummySDCommand(binaryFormat), store.Store(new(dummyStore)), []string{})
	data, err := bin.Run()
	defer os.RemoveAll("/opt/sd")
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if !reflect.DeepEqual(data, []byte("Hello World\n")) {
		t.Errorf("result=%q, want %q", data, []byte("Hello World\n"))
	}

	// success with arguments
	bin, _ = NewBinary(dummySDCommand(binaryFormat), store.Store(new(dummyStore)), []string{"arg1", "arg2"})
	data, err = bin.Run()
	defer os.RemoveAll("/opt/sd")
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if !reflect.DeepEqual(data, []byte("Hello World\narg1\narg2\n")) {
		t.Errorf("result=%q, want %q", data, []byte("Hello World\narg1\narg2\n"))
	}

	// failure. the command is broken
	bin, _ = NewBinary(dummySDCommand(binaryFormat), store.Store(new(dummyStoreBroken)), []string{})
	_, err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. the store api return error
	bin, _ = NewBinary(dummySDCommand(binaryFormat), store.Store(new(dummyStoreError)), []string{})
	_, err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
