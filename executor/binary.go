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
	Args    []string
	Store   store.Store
	Command *store.Command
}

// NewBinary returns Binary object
func NewBinary(spec *util.CommandSpec, arg []string) (*Binary, error) {
	storeapi, err := store.New(config.SDStoreURL, spec)
	if err != nil {
		return nil, err
	}
	binary := &Binary{
		Args:  arg,
		Store: storeapi,
	}
	return binary, nil
}

func (b *Binary) download() error {
	cmd, err := b.Store.GetCommand()
	if err != nil {
		return err
	}
	b.Command = cmd
	return nil
}

func (b *Binary) install() (string, error) {
	dirPath := filepath.Join(config.BaseCommandPath, b.Command.Spec.Namespace, b.Command.Spec.Name, b.Command.Spec.Version)

	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return "", fmt.Errorf("Failed to create command directory: %v", err)
	}

	path := filepath.Join(dirPath, b.Command.Spec.Binary.File)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("Failed to create command file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(b.Command.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to write command file: %v", err)
	}
	if err := os.Chmod(path, 0777); err != nil {
		return "", fmt.Errorf("Failed to change the access permissions of command file: %v", err)
	}
	return path, nil
}

// Run executes command and returns output
func (b *Binary) Run() error {
	lgr.Debug.Println("start downloading binary command.")

	err := b.download()
	if err != nil {
		lgr.Debug.Println(err)
		return err
	}

	lgr.Debug.Println("start installing binary command.")
	path, err := b.install()
	if err != nil {
		lgr.Debug.Println(err)
		return err
	}

	lgr.Debug.Println("start executing binary command.")
	err = execCommand(path, b.Args)
	if err != nil {
		lgr.Debug.Println(err)
	} else {
		lgr.Debug.Println("execute binary command succeeded.")
	}
	return err
}
