package executor

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Executor is a Executor endpoint
type Executor interface {
	Run() error
}

// New returns each format type of Executor
func New(sdAPI api.API, args []string) (Executor, error) {
	ns, name, ver, itr, err := util.SplitCmdWithSearch(args)
	if err != nil {
		return nil, err
	}

	if itr > 2 {
		return nil, fmt.Errorf("exec command argument is not right")
	}
	if itr == 2 && args[1] != "exec" {
		return nil, fmt.Errorf("no such type of command")
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
	m := new(sync.Mutex)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe for exec command: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe for exec command: %v", err)
	}

	log.Println("mmmmmm START COMMAND OUTPUT mmmmmm")
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m.Lock()
			log.Println(scanner.Text())
			m.Unlock()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m.Lock()
			log.Println(scanner.Text())
			m.Unlock()
		}
	}()

	err = cmd.Run()
	log.Println("mmmmmm FINISH COMMAND OUTPUT mmmmmm")
	state := cmd.ProcessState
	log.Printf("System Time: %v, User Time: %v\n", state.SystemTime(), state.UserTime())
	if err != nil {
		return fmt.Errorf("failed to exec command: %v", err)
	}
	return nil
}
