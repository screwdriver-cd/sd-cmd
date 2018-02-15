package util

import (
	"flag"
	"fmt"
)

func ParseCommand(command []string) (map[string]string, error) {
	fs := flag.NewFlagSet(command[0], flag.ExitOnError)
	var (
		ymlPath = fs.String("f", "default-value", "Path of yaml to publish")
	)

	subCommand := command[1]
	err := fs.Parse(command[2:])
	if err != nil {
		return nil, fmt.Errorf("Failed to parse input command:%q", err)
	}

	m := make(map[string]string)
	m["subCommand"] = subCommand
	m["ymlPath"] = *ymlPath

	return m, err
}
