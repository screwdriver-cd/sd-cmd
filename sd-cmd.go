package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/executor"
	"github.com/screwdriver-cd/sd-cmd/logger"
	"github.com/screwdriver-cd/sd-cmd/publisher"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

const minArgLength = 2

func cleanExit() {
	logger.CloseAll()
}

func successExit() {
	cleanExit()
	os.Exit(0)
}

// failureExit exits process with 1
func failureExit(err error) {
	cleanExit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
	os.Exit(1)
}

// finalRecover makes one last attempt to recover from a panic.
// This should only happen if the previous recovery caused a panic.
func finalRecover() {
	if p := recover(); p != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Something terrible has happened. Please file a ticket with this info:")
		fmt.Fprintf(os.Stderr, "ERROR: %v\n%v\n", p, string(debug.Stack()))
		failureExit(nil)
	}
	successExit()
}

func init() {
	config.LoadConfig()
}

func runExecutor(sdAPI api.API, args []string) error {
	exec, err := executor.New(sdAPI, args)
	if err != nil {
		return err
	}
	err = exec.Run()
	if err != nil {
		return err
	}
	return nil
}

func runPublisher(inputCommand []string) {
	pub := publisher.New(inputCommand)
	pub.Run()
}

func runCommand(sdAPI api.API, args []string) error {
	if len(os.Args) < minArgLength {
		return fmt.Errorf("The number of arguments is not enough")
	}

	switch args[1] {
	case "exec":
		return runExecutor(sdAPI, args)
	case "publish":
		runPublisher(args)
		return nil
	case "promote":
		return fmt.Errorf("promote is not implemented yet")
	default:
		return runExecutor(sdAPI, args)
	}
}

func main() {
	defer finalRecover()

	sdAPI := api.New(config.SDAPIURL, config.SDToken)

	err := runCommand(sdAPI, os.Args)
	if err != nil {
		failureExit(err)
	}
}
