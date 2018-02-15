package util

import (
	"testing"

	"github.com/screwdriver-cd/sd-cmd/testdata"
)

var commandSpecYmlPath = testdata.TestDataRootPath + "/command_spec.yml"

func TestLoadFile(t *testing.T) {
	loadFile(commandSpecYmlPath)
}

func TestLoadYml(t *testing.T) {
	actual, _ := LoadYml(commandSpecYmlPath)

	expect := CommandSpec{}
	expect.Namespace = "foo"
	expect.Name = "bar"
	expect.Description = "Lorem ipsum dolor sit amet.\n"
	expect.Version = "1.0"
	expect.Format = "habitat"
	expect.Habitat.Mode = "remote"
	expect.Habitat.Package = "core/git/2.14.1"
	expect.Habitat.Command = "git"
	expect.Docker.Image = "chefdk:1.2.3"
	expect.Docker.Command = "knife"
	expect.Binary.File = "./foobar.sh"

	if *actual != expect {
		t.Errorf("Result should be %v, is %v", expect, actual)
	}
}
