package executor

import (
	"fmt"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Executor is a Executor endpoint
type Executor interface {
	Run() ([]byte, error)
}

// New returns each format type of Executor
func New(sdAPI api.API, args []string) (Executor, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("Args are not enough")
	}
	ns, cmd, ver, itr, err := util.SplitCmdWithSearch(args)
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
		return NewBinary(sdCmd, store, args[itr+1:])
	case "habitat":
		return nil, fmt.Errorf("habitat is not implemented yet")
	case "docker":
		return nil, fmt.Errorf("docker is not implemented yet")
	default:
		return nil, fmt.Errorf("Can not execute such type: %v", sdCmd.Format)
	}
}
