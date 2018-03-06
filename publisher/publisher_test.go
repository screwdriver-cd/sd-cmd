package publisher

import (
	"testing"
)

func TestRun(t *testing.T) {
	testDataPath := "../testdata/yaml/sd-command.yaml"
	pub, err := New([]string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
	pub.Run()
}
