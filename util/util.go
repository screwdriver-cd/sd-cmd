package util

import (
	"fmt"
	"strings"
)

// SplitCmd splits full command to namespace, command, version.
// ex(ns/cmd/1.0.1 => ns cmd 1.0.1)
func SplitCmd(cmd string) (namespace, command, version string, err error) {
	splitNamespce := strings.Split(cmd, "/")
	if len(splitNamespce) < 2 {
		err = fmt.Errorf("Command format is not valid")
		return
	}
	splitCmdAndVer := strings.Split(splitNamespce[1], "@")
	if len(splitCmdAndVer) < 2 {
		err = fmt.Errorf("Command format is not valid")
		return
	}
	namespace = splitNamespce[0]
	command = splitCmdAndVer[0]
	version = splitCmdAndVer[1]
	return
}

// SplitCmdWithSearch searches full command. If there is valid full command, return split full command name.
func SplitCmdWithSearch(cmds []string) (namespace, command, version string, itr int, err error) {
	for i, val := range cmds {
		ns, cmd, ver, err := SplitCmd(val)
		if err == nil {
			return ns, cmd, ver, i, err
		}
	}
	return "", "", "", -1, fmt.Errorf("There is no valid command format")
}
