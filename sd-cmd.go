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

var cleanExit = func() {
	logger.Close()
	os.Exit(1)
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
	case "publish":
		return fmt.Errorf("publish is not implemented yet")
	case "promote":
		return fmt.Errorf("promote is not implemented yet")
	default:
		exec, err := executor.New(sdAPI, args)
		if err != nil {
			return err
		}
		output, err := exec.Run()
		if err != nil {
			return err
		}
		logger.Write(fmt.Sprintf("command output: %v", string(output)))
		return nil
	}
}

func main() {
	defer finalRecover()

	if len(os.Args) < 2 {
		log.Println("The argument num is not enough")
		os.Exit(1)
	}

	err := logger.MakeLogToFile(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer logger.Close()

	sdAPI, err := api.New()
	if err != nil {
		logger.Write(err)
		os.Exit(1)
	}

	err = runCommand(sdAPI, os.Args)
	if err != nil {
		logger.Write(err)
		os.Exit(1)
	}
}
