package executor

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/logger"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

var (
	command = exec.Command
	lgr     *logger.Logger
)

// Executor is a Executor endpoint
type Executor interface {
	Run() error
}

func prepareLog(smallSpec *util.CommandSpec) (err error) {
	dirPath := filepath.Join(config.SDArtifactsDir, ".sd", "commands")
	filename := fmt.Sprintf("%v-%v-%v.log", time.Now().Unix(), smallSpec.Namespace, smallSpec.Name)
	lgr, err = logger.New(dirPath, filename, log.LstdFlags, false)
	return
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
		return NewHabitat(spec, args[pos+1:])
	case "docker":
		return nil, errors.New("the docker format is not yet implemented")
	default:
		return nil, errors.New("the format is not allowed")
	}
}

func execCommand(path string, args []string) (err error) {
	cmd := command(path, args...)
	lgr.Debug.Println("mmmmmm START COMMAND OUTPUT mmmmmm")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	lgr.Debug.Println("mmmmmm FINISH COMMAND OUTPUT mmmmmm")
	if err != nil {
		lgr.Debug.Printf("failed to exec command: %v", err)
		return
	}

	state := cmd.ProcessState
	lgr.Debug.Printf("System Time: %v, User Time: %v\n", state.SystemTime(), state.UserTime())
	return
}
