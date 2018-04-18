package executor

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

var dummyArgs = []string{"arg1", "arg2"}

func fakeExecCommand(name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	// fmt.Printf("%v \n %v", os.Args, cs)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestNewHabitat(t *testing.T) {
	_, err := NewHabitat(dummyAPICommand(habitatFormat), dummyArgs)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestRunHabitat(t *testing.T) {
	logBuffer.Reset()
	command = fakeExecCommand
	defer func() { command = exec.Command }()

	spec := dummyAPICommand(habitatFormat)
	hab, _ := NewHabitat(spec, dummyArgs)
	err := hab.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "no command\n")
		os.Exit(2)
	}

	cmd, subcmd, subsubcmd, args := args[0], args[1], args[2], args[3:]

	if cmd != "/hab/bin/hab" {
		fmt.Fprintf(os.Stderr, "expected '/hab/bin/hab', but %v\n", cmd)
		os.Exit(1)
	}

	if subcmd != "pkg" {
		fmt.Fprintf(os.Stderr, "expected 'pkg', but %v\n", subcmd)
		os.Exit(1)
	}

	switch subsubcmd {
	case "exec":
		execDummyArgs := append([]string{dummyPackage, dummyCommand}, dummyArgs...)
		argsLen := len(args)
		dummyLen := len(execDummyArgs)
		if argsLen != dummyLen {
			fmt.Fprintf(os.Stderr, "length of exec args is expected %v, but %v\n", dummyLen, argsLen)
			os.Exit(1)
		}
		for i := range execDummyArgs {
			if args[i] != execDummyArgs[i] {
				fmt.Fprintf(os.Stderr, "exec cmd args is expected %v, but %v\n", dummyLen, argsLen)
				os.Exit(1)
			}
		}
	case "install":
		installDummyArgs := []string{dummyPackage}
		argsLen := len(args)
		dummyLen := len(installDummyArgs)
		if argsLen != dummyLen {
			fmt.Fprintf(os.Stderr, "length of install cmd args is expected %v, but %v\n", dummyLen, argsLen)
			os.Exit(1)
		}
		for i := range installDummyArgs {
			if args[i] != installDummyArgs[i] {
				fmt.Fprintf(os.Stderr, "install cmd args is expected %v, but %v\n", dummyLen, argsLen)
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "hab command is something wrong")
		os.Exit(1)
	}
}
