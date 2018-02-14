package util

import "flag"

func ParseCommand(command []string) map[string]string {
	fs := flag.NewFlagSet(command[0], flag.ExitOnError)
	var (
		ymlPath = fs.String("f", "default-value", "Path of yaml to publish")
	)

	subCommand := command[1]
	fs.Parse(command[2:])
	// fmt.Println("subComamand:", subCommand)
	// fmt.Println("ymlPath:", *ymlPath)
	// fmt.Println("args:", fs.Args())

	m := make(map[string]string)
	m["subCommand"] = subCommand
	m["ymlPath"] = *ymlPath

	return m
}
