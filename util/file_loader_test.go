package util

import (
	"testing"
)

const (
	commandSpecYamlPath        = "../testdata/yaml/sd-command.yaml"
	invalidCommandSpecYamlPath = "../testdata/yaml/invalid_sd-command.yaml"
)

func TestLoadFile(t *testing.T) {
	LoadByte(commandSpecYamlPath)
}

func TestLoadYaml(t *testing.T) {
	actual, _ := LoadYaml(commandSpecYamlPath)

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

	expect = "foo@bar.com"
	if actual.Maintainer != expect {
		t.Errorf("got %q\nwant %q", actual.Maintainer, expect)
	}

	expect = "1.0"
	if actual.Version != expect {
		t.Errorf("got %q\nwant %q", actual.Version, expect)
	}

	expect = "binary"
	if actual.Format != expect {
		t.Errorf("got %q\nwant %q", actual.Format, expect)
	}

	expect = "./testdata/binary/hello"
	if actual.Binary.File != expect {
		t.Errorf("got %q\nwant %q", actual.Binary.File, expect)
	}

	// failure. no yaml path
	actual, err := LoadYaml("./there/is/no/such/path")
	if err == nil {
		t.Errorf("err=nil, want error")
	}

	// failure. invalid yaml
	actual, err = LoadYaml(invalidCommandSpecYamlPath)
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
