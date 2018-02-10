package publisher

import (
	"testing"
)

func TestNew(t *testing.T) {
	_, err := New([]string{"sd-cmd", "publish", "-f", "./testdata/command_spec.yml"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestRun(t *testing.T) {
	pub, _ := New([]string{"sd-cmd", "publish", "-f", "./testdata/command_spec.yml"})
	pub.Run()
}
