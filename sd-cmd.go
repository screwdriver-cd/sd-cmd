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

func runExecutor(sdAPI api.API, args []string) error {
	executor, err := executor.New(sdAPI, args)
	if err != nil {
		return err
	}
	output, err := executor.Run()
	if err != nil {
		fmt.Println(string(output))
		return err
	}
	fmt.Println(string(output))
	return nil
}

func runCommand(sdAPI api.API, args []string) error {
	switch args[1] {
	case "exec":
		return runExecutor(sdAPI, args)
	case "publish":
		return fmt.Errorf("publish is not implemented yet")
	case "promote":
		return fmt.Errorf("promote is not implemented yet")
	default:
		return runExecutor(sdAPI, args)
	}
}

func main() {
	defer finalRecover()

	if len(os.Args) < 2 {
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
		fmt.Println(err)
		os.Exit(0)
	}
}
