package executor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/screwdriver-cd/sd-cmd/config"
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
	_, err := NewHabitat(dummyCommandSpec(habitatFormat), dummyArgs)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
}

func TestGetPkgDirPath(t *testing.T) {
	spec := dummyCommandSpec(habitatFormat)
	hab, _ := NewHabitat(spec, []string{})
	hab.Store = newDummyStore(validShell, spec, nil)
	// Note: config.BaseCommandPath is customized for test.
	// see executor/executor_test.go
	assert.Equal(t, hab.getPkgDirPath(), filepath.Join(config.BaseCommandPath, "foo-dummy/name-dummy/1.0.1"))
}

func TestGetPkgFilePath(t *testing.T) {
	spec := dummyCommandSpec(habitatFormat)
	hab, _ := NewHabitat(spec, []string{})
	hab.Store = newDummyStore(validShell, spec, nil)
	assert.Equal(t, hab.getPkgFilePath(), filepath.Join(config.BaseCommandPath, "foo-dummy/name-dummy/1.0.1/dummy.hart"))
}

func TestIsDownloaded(t *testing.T) {
	spec := dummyCommandSpec(habitatFormat)
	hab, _ := NewHabitat(spec, []string{})
	hab.Store = newDummyStore(validShell, spec, nil)
	// Not exists
	assert.False(t, hab.isDownloaded())

	// 0 size file
	os.MkdirAll(hab.getPkgDirPath(), 0777)
	file, _ := os.Create(hab.getPkgFilePath())
	assert.False(t, hab.isDownloaded())

	// non 0 size file
	file.Write(([]byte)("dummy script."))
	assert.True(t, hab.isDownloaded())

	defer file.Close()
	defer os.RemoveAll(hab.getPkgDirPath())
}

func TestRunHabitat(t *testing.T) {
	logBuffer.Reset()
	command = fakeExecCommand
	defer func() { command = exec.Command }()

	spec := dummyCommandSpec(habitatFormat)

	// case remote mode
	hab, _ := NewHabitat(spec, dummyArgs)
	err := hab.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case local mode
	spec.Habitat.Mode = "local"
	os.Setenv("HABITAT_MODE", "local")
	hab, _ = NewHabitat(spec, dummyArgs)
	hab.Store = newDummyStore(validShell, spec, nil)
	err = hab.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// case hart file is not downloaded
	spec.Habitat.File = dummyEmptyFile
	assert.False(t, hab.isDownloaded())
	err = hab.Run()
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}
	hartPath := filepath.Join(config.BaseCommandPath, spec.Namespace, spec.Name, spec.Version, dummyEmptyFile)
	os.Remove(hartPath)

	// store returns error
	hab.Store = newDummyStore(validShell, spec, fmt.Errorf("store cause error"))
	err = hab.Run()
	if err == nil {
		t.Errorf("err=nil, want error")
	}
	os.Remove(hartPath)
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

	if cmd != "/opt/sd/bin/hab" {
		fmt.Fprintf(os.Stderr, "expected '/opt/sd/bin/hab', but %v\n", cmd)
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
				fmt.Fprintf(os.Stderr, "exec cmd args is expected %v, but %v\n", execDummyArgs[i], args[i])
				os.Exit(1)
			}
		}
	case "install":
		var installDummyArgs []string
		if os.Getenv("HABITAT_MODE") == "local" {
			installDummyArgs[0] = filepath.Join(config.BaseCommandPath, "foo-dummy/name-dummy/1.0.1/dummy.hart")
		} else {
			installDummyArgs[0] = dummyPackage
		}
		argsLen := len(args)
		dummyLen := len(installDummyArgs)
		if argsLen != dummyLen {
			fmt.Fprintf(os.Stderr, "length of install cmd args is expected %v, but %v\n", dummyLen, argsLen)
			os.Exit(1)
		}
		for i := range installDummyArgs {
			if args[i] != installDummyArgs[i] {
				fmt.Fprintf(os.Stderr, "install cmd args is expected %v, but %v\n", installDummyArgs[i], args[i])
				os.Exit(1)
			}
		}
	default:
		fmt.Fprintln(os.Stderr, "hab command is something wrong")
		os.Exit(1)
	}
}
