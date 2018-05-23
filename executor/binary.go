package executor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Binary is a Binary Executor object
type Binary struct {
	Args []string
	// from SD API
	Spec  *util.CommandSpec
	Store store.Store
	// Note: this property is set after downloaing a binary via Store API
	Command *store.Command
}

// NewBinary returns Binary object
func NewBinary(spec *util.CommandSpec, arg []string) (*Binary, error) {
	storeapi := store.New(config.SDStoreURL, spec, config.SDToken)

	binary := &Binary{
		Args:  arg,
		Spec:  spec,
		Store: storeapi,
	}
	return binary, nil
}

func (b *Binary) getBinDirPath() string {
	return filepath.Join(config.BaseCommandPath, b.Spec.Namespace, b.Spec.Name, b.Spec.Version)
}

func (b *Binary) getBinFilePath() string {
	fileName := filepath.Base(b.Spec.Binary.File)
	return filepath.Join(b.getBinDirPath(), fileName)
}

func (b *Binary) isInstalled() bool {
	fInfo, err := os.Stat(b.getBinFilePath())
	if err != nil {
		return false
	}
	if fInfo.Size() == 0 {
		return false
	}
	return true
}

func (b *Binary) download() error {
	cmd, err := b.Store.GetCommand()
	if err != nil {
		return err
	}
	b.Command = cmd
	return nil
}

func (b *Binary) install() error {
	if err := os.MkdirAll(b.getBinDirPath(), 0777); err != nil {
		return fmt.Errorf("Failed to create command directory: %v", err)
	}

	path := b.getBinFilePath()
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Failed to create command file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(b.Command.Body)
	if err != nil {
		return fmt.Errorf("Failed to write command file: %v", err)
	}
	if err := os.Chmod(path, 0777); err != nil {
		return fmt.Errorf("Failed to change the access permissions of command file: %v", err)
	}
	return nil
}

// Run executes command and returns output
func (b *Binary) Run() error {
	if b.isInstalled() {
		lgr.Debug.Println("binary command already installed, skip installation.")
	} else {
		lgr.Debug.Println("start downloading binary command.")

		err := b.download()
		if err != nil {
			lgr.Debug.Println(err)
			return err
		}

		lgr.Debug.Println("start installing binary command.")
		err = b.install()
		if err != nil {
			lgr.Debug.Println(err)
			return err
		}
	}

	lgr.Debug.Println("start executing binary command.")
	err := execCommand(b.getBinFilePath(), b.Args)
	if err != nil {
		lgr.Debug.Println(err)
	} else {
		lgr.Debug.Println("execute binary command succeeded.")
	}
	return err
}
