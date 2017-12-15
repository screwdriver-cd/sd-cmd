package executor

import (
	"fmt"
	"strings"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
)

// Executor is a Executor endpoint
type Executor interface {
	Run() ([]byte, error)
}

func splitCmd(cmd string) (namespace, command, version string, err error) {
	splitNamespce := strings.Split(cmd, "/")
	if len(splitNamespce) < 2 {
		err = fmt.Errorf("Exec command format is not valid")
		return
	}
	splitCmdAndVer := strings.Split(splitNamespce[1], "@")
	if len(splitCmdAndVer) < 2 {
		err = fmt.Errorf("Exec command format is not valid")
		return
	}
	namespace = splitNamespce[0]
	command = splitCmdAndVer[0]
	version = splitCmdAndVer[1]
	return
}

// New returns each format type of Executor
func New(sdAPI api.API, args []string) (Executor, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Args are not enough")
	}

	ns, cmd, ver, err := splitCmd(args[0])
	if err != nil {
		return nil, err
	}
	sdCmd, err := sdAPI.GetCommand(ns, cmd, ver)
	if err != nil {
		return nil, err
	}
	switch sdCmd.Format {
	case "binary":
		store, err := store.New(sdCmd)
		if err != nil {
			return nil, err
		}
		return NewBinary(sdCmd, store, args[1:])
	case "habitat":
		return nil, fmt.Errorf("habitat is not implemented yet")
	case "docker":
		return nil, fmt.Errorf("docker is not implemented yet")
	default:
		return nil, fmt.Errorf("Can not execute such type: %v", sdCmd.Format)
	}
}
