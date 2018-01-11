package executor

import (
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Executor is a Executor endpoint
type Executor interface {
	Run() ([]byte, error)
}

// New returns each format type of Executor
func New(sdAPI api.API, args []string) (Executor, error) {
	ns, name, ver, itr, err := util.SplitCmdWithSearch(args)
	if err != nil {
		return nil, err
	}
	spec, err := sdAPI.GetCommand(ns, name, ver)
	if err != nil {
		return nil, err
	}
	switch spec.Format {
	case "binary":
		return NewBinary(spec, args[itr+1:])
	case "habitat":
		return nil, nil
	case "docker":
		return nil, nil
	}
	return nil, nil
}
