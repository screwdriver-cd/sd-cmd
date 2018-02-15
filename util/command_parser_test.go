package util

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	// Parse success
	command := []string{"sd-cmd", "publish", "-f", "command_spec.yml"}

	actual, err := ParseCommand(command)
	if err != nil {
		t.Errorf("err=%q", err)
	}

	expected := map[string]string{
		"subCommand": "publish",
		"ymlPath":    "command_spec.yml",
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("got %v\nwant %v", actual, expected)
	}
}
