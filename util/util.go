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

// tagRegexp check VERSION of Tags. Tags can only be named with A-Z,a-z,0-9,-,.
// ex(latest stable feature-abc v1.0.0)
var tagRegexp = regexp.MustCompile(`^[a-zA-Z][\w-.]+$`)

// A Habitat represents a set of data for Habitat.
// All value will be omitted if it is not set.
// This will works as a part of CommandSpec.
type Habitat struct {
	Mode    string `json:"mode,omitempty" yaml:"mode,omitempty"`
	File    string `json:"file,omitempty" yaml:"file,omitempty"`
	Package string `json:"package,omitempty" yaml:"package,omitempty"`
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
}

// A Docker represents a set of data for Docker.
// All value will be omitted if it is not set.
// This will works as a part of CommandSpec.
type Docker struct {
	Image   string `json:"image,omitempty" yaml:"image,omitempty"`
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
}

// A Binary represents a set of data for Binary.
// All value will be omitted if it is not set.
// This will works as a part of CommandSpec.
type Binary struct {
	File string `json:"file,omitempty" yaml:"file,omitempty"`
}

// A CommandSpec represents a set of data for commands.
// Some value will be omitted if it is not set.
type CommandSpec struct {
	ID           int      `json:"id,omitempty" yaml:"id,omitempty"`
	Namespace    string   `json:"namespace" yaml:"namespace"`
	Name         string   `json:"name" yaml:"name"`
	Description  string   `json:"description" yaml:"description"`
	Usage        string   `json:"usage,omitempty" yaml:"usage,omitempty"`
	Maintainer   string   `json:"maintainer" yaml:"maintainer"`
	Version      string   `json:"version" yaml:"version"`
	Format       string   `json:"format" yaml:"format"`
	Habitat      *Habitat `json:"habitat,omitempty" yaml:"habitat,omitempty"`
	Docker       *Docker  `json:"docker,omitempty" yaml:"docker,omitempty"`
	Binary       *Binary  `json:"binary,omitempty" yaml:"binary,omitempty"`
	PipelineID   int      `json:"pipelineId,omitempty" yaml:"pipelineId,omitempty"`
	SpecYamlPath string   `json:"-" yaml:"-"`
}

// PayloadYaml represents a set of data for posting command.
// The key "yaml" and the value of json string is needed to post to api.
type PayloadYaml struct {
	Yaml string `json:"yaml"`
}

// ValidateResponse represents a response from API when validates command.
type ValidateResponse struct {
	Errors []ValidateError `json:"errors"`
}

// ValidateError represents an error message of a command validation.
type ValidateError struct {
	Message string `json:"message"`
}

// TagTargetVersion represents a body of a tagging request.
type TagTargetVersion struct {
	Version string `json:"version"`
}

// TagResponse represents a response from API when tags command
type TagResponse struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	Version   string `json:"version"`
}

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
func SplitCmd(cmd string) (smallSpec *CommandSpec, err error) {
	smallSpec = new(CommandSpec)
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

	smallSpec.Namespace = values[0][1]
	smallSpec.Name = values[0][2]
	smallSpec.Version = values[0][3]
	return
}

// SplitCmdWithSearch searches full command. If there is valid full command, return split full command name.
func SplitCmdWithSearch(cmds []string) (smallSpec *CommandSpec, pos int, err error) {
	for i, val := range cmds {
		smallSpec, err := SplitCmd(val)
		if err == nil {
			return smallSpec, i, err
		}
	}
	return nil, -1, fmt.Errorf("There is no valid command format")
}

// ValidateTagName validates tag name
func ValidateTagName(tag string) bool {
	if tagRegexp.Match([]byte(tag)) {
		return true
	}
	return false
}
