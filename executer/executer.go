package executer

import (
	"github.com/screwdriver-cd/sd-cmd/api/screwdriver"
)

// Executer is a Executer endpoint
type Executer interface {
	Run() ([]byte, error)
}

// New return each format type of Executer
func New(args []string) (Executer, error) {
	sdAPI, err := screwdriver.New()
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
