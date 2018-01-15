package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/executor"
	"github.com/screwdriver-cd/sd-cmd/logger"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

const minArgLength = 2

var cleanExit = func() {
	logger.CloseAll()
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

func runCommand(sdAPI api.API, args []string) error {
	if len(os.Args) < minArgLength {
		return fmt.Errorf("The argument num is not enough")
	}

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

	sdAPI, err := api.New(config.SDAPIURL, config.SDToken)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}

	err = runCommand(sdAPI, os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
}
