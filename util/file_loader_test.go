package util

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/testdata"
)

var commandSpecYmlPath = testdata.TestDataRootPath + "/sd-command.yaml"

func TestLoadFile(t *testing.T) {
	loadFile(commandSpecYmlPath)
}

func TestLoadYml(t *testing.T) {
	actual, _ := LoadYml(commandSpecYmlPath)

	expect := "foo"
	if actual.Namespace != expect {
		t.Errorf("got %q\nwant %q", actual.Namespace, expect)
	}

	expect = "bar"
	if actual.Name != expect {
		t.Errorf("got %q\nwant %q", actual.Name, expect)
	}

	expect = "Lorem ipsum dolor sit amet.\n"
	if actual.Description != expect {
		t.Errorf("got %q\nwant %q", actual.Description, expect)
	}

	expect = "1.0"
	if actual.Version != expect {
		t.Errorf("got %q\nwant %q", actual.Version, expect)
	}

	expect = "habitat"
	if actual.Format != expect {
		t.Errorf("got %q\nwant %q", actual.Format, expect)
	}

	expect = "remote"
	if actual.Habitat.Mode != expect {
		t.Errorf("got %q\nwant %q", actual.Habitat.Mode, expect)
	}

	expect = "core/git/2.14.1"
	if actual.Habitat.Package != expect {
		t.Errorf("got %q\nwant %q", actual.Habitat.Package, expect)
	}

	expect = "git"
	if actual.Habitat.Command != expect {
		t.Errorf("got %q\nwant %q", actual.Habitat.Command, expect)
	}
}
