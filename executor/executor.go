package executor

import (
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

// Executor is a Executor endpoint
type Executor interface {
	Run() ([]byte, error)
}

// New return each format type of Executor
func New(args []string) (Executor, error) {
	sdAPI, err := api.New()
	if err != nil {
		return nil, err
	}
	sdCmd, err := sdAPI.GetCommand("namespace", "command", "version")
	if err != nil {
		return nil, err
	}
	switch sdCmd.Format {
	case "binary":
		return NewBinary(sdCmd, args[1:])
	case "habitat":
		return nil, nil
	case "docker":
		return nil, nil
	}
	return nil, nil
}
