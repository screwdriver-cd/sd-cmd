package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/executor"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

var cleanExit = func() {
	os.Exit(0)
}

// finalRecover makes one last attempt to recover from a panic.
// This should only happen if the previous recovery caused a panic.
func finalRecover() {
	if p := recover(); p != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Something terrible has happened. Please file a ticket with this info:")
		fmt.Fprintf(os.Stderr, "ERROR: %v\n%v\n", p, debug.Stack())
	}
	cleanExit()
}

func init() {
	config.LoadConfig()
}

func runCommand(sdAPI api.API, args []string) error {
	switch args[1] {
	case "exec":
		executor, err := executor.New(sdAPI, args[2:])
		if err != nil {
			return fmt.Errorf("Failed to create executor: %v", err)
		}
		output, err := executor.Run()
		if err != nil {
			fmt.Println(string(output))
			return fmt.Errorf("Failed to run exec command: %v", err)
		}
		fmt.Println(string(output))
		return nil
	case "publish":
		return fmt.Errorf("publish is not implemented yet")
	case "promote":
		return fmt.Errorf("promote is not implemented yet")
	default:
		return fmt.Errorf("No such type of command")
	}
}

func main() {
	defer finalRecover()

	if len(os.Args) < 3 {
		fmt.Printf("The argument num is not enough\n")
		os.Exit(0)
	}

	sdAPI, err := api.New(config.SDAPIURL, config.SDToken)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = runCommand(sdAPI, os.Args)
	if err != nil {
		fmt.Printf("error happen: %v\n", err)
		os.Exit(0)
	}
}
