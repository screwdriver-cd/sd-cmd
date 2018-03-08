package executor

import (
	"fmt"
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

var lgr *logger.Logger

// Executor is a Executor endpoint
type Executor interface {
	Run() error
}

func prepareLog(smallSpec *util.CommandSpec) (err error) {
	dirPath := filepath.Join(config.SDArtifactsDir, ".sd", "commands", smallSpec.Namespace, smallSpec.Name, smallSpec.Version)
	filename := fmt.Sprintf("%v.log", time.Now().Unix())
	lgr, err = logger.New(dirPath, filename, log.LstdFlags, false)
	if err != nil {
		return err
	}
	return nil
}

// New returns each format type of Executor
func New(sdAPI api.API, args []string) (Executor, error) {
	smallSpec, pos, err := util.SplitCmdWithSearch(args)
	if err != nil {
		return nil, err
	}

	err = prepareLog(smallSpec)
	if err != nil {
		return nil, err
	}

	spec, err := sdAPI.GetCommand(smallSpec)
	if err != nil {
		return nil, err
	}
	switch spec.Format {
	case "binary":
		return NewBinary(spec, args[pos+1:])
	case "habitat":
		return nil, nil
	case "docker":
		return nil, nil
	}
	return nil, nil
}

func execCommand(path string, args []string) error {
	cmd := exec.Command(path, args...)
	lgr.Debug.Println("mmmmmm START COMMAND OUTPUT mmmmmm")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	lgr.Debug.Println("mmmmmm FINISH COMMAND OUTPUT mmmmmm")
	state := cmd.ProcessState
	lgr.Debug.Printf("System Time: %v, User Time: %v\n", state.SystemTime(), state.UserTime())
	if err != nil {
		return fmt.Errorf("failed to exec command: %v", err)
	}
	return nil
}
