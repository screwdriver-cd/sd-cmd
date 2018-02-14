package publisher

import (
	"testing"
)

// func TestNew(t *testing.T) {
// 	pub := New([]string{"sd-cmd", "publish", "-f", "./testdata/command_spec.yml"})
// 	if err != nil {
// 		t.Errorf("err=%q, want nil", err)
// 	}
// }

func TestRun(t *testing.T) {
	pub := New([]string{"sd-cmd", "publish", "-f", "./testdata/command_spec.yml"})
	pub.Run()
}
