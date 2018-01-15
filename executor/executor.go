package executor

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/logger"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

var lager *logger.Logger

// Executor is a Executor endpoint
type Executor interface {
	Run() error
}

func prepareLog(namespace, name, version string) (err error) {
	dirPath := filepath.Join(config.SDArtifactsDir, ".sd", "commands", namespace, name, version)
	filename := fmt.Sprintf("%v.log", time.Now().Unix())
	lager, err = logger.New(dirPath, filename, log.LstdFlags, false)
	if err != nil {
		return err
	}
	return nil
}

// New returns each format type of Executor
func New(sdAPI api.API, args []string) (Executor, error) {
	ns, name, ver, itr, err := util.SplitCmdWithSearch(args)
	if err != nil {
		return nil, err
	}

	err = prepareLog(ns, name, ver)
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

func execCommand(path string, args []string) error {
	cmd := exec.Command(path, args...)
	lager.Debug.Println("mmmmmm START COMMAND OUTPUT mmmmmm")

	cmd.Stdout = io.MultiWriter(lager.File, os.Stderr)
	cmd.Stderr = cmd.Stdout

	err := cmd.Run()

	lager.Debug.Println("mmmmmm FINISH COMMAND OUTPUT mmmmmm")
	state := cmd.ProcessState
	lager.Debug.Printf("System Time: %v, User Time: %v\n", state.SystemTime(), state.UserTime())
	if err != nil {
		return fmt.Errorf("failed to exec command: %v", err)
	}
	return nil
}
