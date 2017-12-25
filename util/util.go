package util

import (
	"fmt"
	"regexp"
)

// full command has <COMMAND_NAMESPACE>/<COMMAND_NAME>@<VERSION>.
// COMMAND_NAMESPACE can only be named with A-Z,a-z,0-9,-,_
// COMMAND_NAME can only be named with A-Z,a-z,0-9,-,_
// VERSION can only be a-z0-9.~*^
// ex(cmd-namespace/cmd_name@1.0.0)
var fullCommandRegexp = regexp.MustCompile(`^([\w-]+)\/([\w-]+)@([a-z0-9-~\*\^\.]+)$`)

// xrangesRegexp check VERSION of X-Ranges.
// ex(1.2.* 1.2 1.x 1 *)
var xrangesRegexp = regexp.MustCompile(`^(?:(\d+)\.)?(?:(\d+)\.)?([\*x]|\d+)$`)

// tildeRangesRegexp check VERSION of Tilde Ranges
// ex(~1.2.3 ~1.2 ~1)
var tildeRangesRegexp = regexp.MustCompile(`^~\d(\.\d)?(\.\d)?$`)

// caretRangesAndPinningRegexp check VERSION of Caret Ranges and Explicit Pinning
// ex(^1.2.3 ^0.2.5 ^0.0.4 1.2.3 1.5.3)
var caretRangesAndPinningRegexp = regexp.MustCompile(`^(\^)?\d(\.\d){2}$`)

// tagRegexp check VERSION of Tags. Tags can only be named with A-Z,a-z,0-9,-
// ex(latest stable feature-abc)
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
