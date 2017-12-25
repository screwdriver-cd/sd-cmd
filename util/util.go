package util

import (
	"fmt"
	"regexp"
)

var fullCommandRegexp = regexp.MustCompile(`^([\w-]+)\/([\w-]+)@([a-z0-9-~\*\^\.]+)$`)
var xrangesRegexp = regexp.MustCompile(`^(?:(\d+)\.)?(?:(\d+)\.)?([\*x]|\d+)$`)
var tildeRangesRegexp = regexp.MustCompile(`^~\d(\.\d)?(\.\d)?$`)
var caretRangesAndPinningRegexp = regexp.MustCompile(`^(\^)?\d(\.\d){2}$`)
var tagRegexp = regexp.MustCompile(`^[a-z][a-z0-9-]+$`)

func checkVersion(ver string) bool {
	if caretRangesAndPinningRegexp.Match([]byte(ver)) {
		return true
	}
	if tildeRangesRegexp.Match([]byte(ver)) {
		return true
	}
	if xrangesRegexp.Match([]byte(ver)) {
		return true
	}
	if tagRegexp.Match([]byte(ver)) {
		return true
	}
	return false
}

// SplitCmd splits full command to namespace, command, version.
// ex(ns/cmd/1.0.1 => ns cmd 1.0.1)
func SplitCmd(cmd string) (namespace, command, version string, err error) {
	values := fullCommandRegexp.FindAllStringSubmatch(cmd, -1)
	if len(values) < 1 {
		err = fmt.Errorf("There is no full command")
		return
	}

	if len(values[0]) != 4 {
		err = fmt.Errorf("There is something wrong with the full command")
		return
	}

	if !checkVersion(values[0][3]) {
		err = fmt.Errorf("The version part is invalid")
	}

	namespace = values[0][1]
	command = values[0][2]
	version = values[0][3]
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
