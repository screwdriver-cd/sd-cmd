package publisher

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/datafortest"
)

func TestRun(t *testing.T) {
	testDataPath := datafortest.TestDataRootPath + "/yaml/sd-command.yaml"
	pub, err := New([]string{"-f", testDataPath})
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
	pub.Run()
}
