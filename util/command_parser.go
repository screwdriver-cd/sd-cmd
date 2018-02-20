package util

import (
	"flag"
	"fmt"
)

// ParseCommand parses and check user input.
// It returns a map if input is valid.
func ParseCommand(command []string) (map[string]string, error) {
	fs := flag.NewFlagSet(command[0], flag.ContinueOnError)
	var (
		ymlPath = fs.String("f", "./sd-command.yaml", "Path of yaml to publish")
	)

	subCommand := command[1]
	err := fs.Parse(command[2:])
	if err != nil {
		return nil, fmt.Errorf("Failed to parse input command:%v", err)
	}

	m := make(map[string]string)
	m["subCommand"] = subCommand
	m["ymlPath"] = *ymlPath

	return m, err
}
