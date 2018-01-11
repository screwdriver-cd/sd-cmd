package executor

import (
	"fmt"
	"reflect"
	"testing"

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
	// success with no arguments
	bin, _ := NewBinary(dummyAPICommand(binaryFormat), []string{})
	bin.Store = store.Store(new(dummyStore))
	data, err := bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if !reflect.DeepEqual(data, []byte("Hello World\n")) {
		t.Errorf("result=%q, want %q", data, []byte("Hello World\n"))
	}

	// success with arguments
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{"arg1", "arg2"})
	bin.Store = store.Store(new(dummyStore))
	data, err = bin.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	if !reflect.DeepEqual(data, []byte("Hello World\narg1\narg2\n")) {
		t.Errorf("result=%q, want %q", data, []byte("Hello World\narg1\narg2\n"))
	}

	// failure. the command is broken
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{})
	bin.Store = store.Store(new(dummyStoreBroken))
	_, err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. the store api return error
	bin, _ = NewBinary(dummyAPICommand(binaryFormat), []string{})
	bin.Store = store.Store(new(dummyStoreError))
	_, err = bin.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
