package publisher

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/testdata"
)

// func TestNew(t *testing.T) {
// 	pub := New([]string{"sd-cmd", "publish", "-f", "./testdata/command_spec.yml"})
// 	if err != nil {
// 		t.Errorf("err=%q, want nil", err)
// 	}
// }

func TestRun(t *testing.T) {
	testDataPath := testdata.TestDataRootPath + "/command_spec.yml"
	pub, err := New([]string{"sd-cmd", "publish", "-f", testDataPath})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	pub.Run()
}
