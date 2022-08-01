package executor

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

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

// exec subcommand flags
var (
	isDebug   = false
	isVerbose = false
)

// Executor is a Executor endpoint
type Executor interface {
	Run() error
}

func prepareLog(smallSpec *util.CommandSpec, isDebug bool) (err error) {
	var file io.WriteCloser

	if isDebug || config.DEBUG {
		dirPath := filepath.Join(config.SDArtifactsDir, ".sd", "commands")
		filename := fmt.Sprintf("%v-%v-%v.log", time.Now().Unix(), smallSpec.Namespace, smallSpec.Name)
		file, err = createLogFile(dirPath, filename)
		if err != nil {
			return err
		}
	}
	lgr, err = logger.New(file)
	return
}

func parseExecSubCommands(args []string) ([]string, error) {
	f := flag.NewFlagSet("exec", flag.ContinueOnError)
	f.BoolVar(&isDebug, "debug", false, "output log to file")
	f.BoolVar(&isVerbose, "v", false, "output verbose log to console")
	err := f.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exec args: %w", err)
	}
	return f.Args(), nil
}

func createLogFile(dirPath, filename string) (io.WriteCloser, error) {
	if filename == "" {
		filename = fmt.Sprintf("%v.log", time.Now().Unix())
	}
	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return nil, fmt.Errorf("Failed to create logging directory: %v", err)
	}

	filePath := filepath.Join(dirPath, filename)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Failed to create logging file: %v", err)
	}
	return file, nil
}

// New returns each format type of Executor
func New(sdAPI api.API, args []string) (Executor, error) {
	args, err := parseExecSubCommands(args)
	if err != nil {
		return nil, err
	}

	smallSpec, pos, err := util.SplitCmdWithSearch(args)
	if err != nil {
		return nil, err
	}

	err = prepareLog(smallSpec, isDebug)
	if err != nil {
		return nil, err
	}

	sdAPI.SetVerbose(isVerbose)

	spec, err := sdAPI.GetCommand(smallSpec)
	if err != nil {
		return nil, err
	}

	switch spec.Format {
	case "binary":
		return NewBinary(spec, args[pos+1:], isVerbose)
	case "habitat":
		return NewHabitat(spec, args[pos+1:], isVerbose)
	case "docker":
		return nil, errors.New("the docker format is not yet implemented")
	default:
		return nil, errors.New("the format is not allowed")
	}
}

func execCommand(path string, args []string) (err error) {
	cmd := command(path, args...)
	if !terminal.IsTerminal(syscall.Stdin) {
		cmd.Stdin = os.Stdin
	}

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

// CleanUp close file you use.
func CleanUp() {
	lgr.Close()
}
