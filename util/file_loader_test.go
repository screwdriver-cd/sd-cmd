package util

import (
	"testing"
)

func TestLoadFile(t *testing.T) {
	inputPath := "./testdata/command_spec.yml"
	loadFile(inputPath)
}

func TestLoadYml(t *testing.T) {
	inputPath := "./testdata/command_spec.yml"
	LoadYml(inputPath)
}
