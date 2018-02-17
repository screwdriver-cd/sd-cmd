package publisher

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/testdata"
)

func TestRun(t *testing.T) {
	testDataPath := testdata.TestDataRootPath + "/command_spec.yml"
	pub, err := New([]string{"sd-cmd", "publish", "-f", testDataPath})
	if err != nil {
		t.Errorf("err=%v, want nil", err)
	}
	pub.Run()
}
