package executor

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/screwdriver-cd/sd-cmd/logger"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
)

// Binary is a Binary Executor object
type Binary struct {
	APICommand *api.Command
	Arg        []string
	Store      store.Store
}

// NewBinary returns Binary object
func NewBinary(cmd *api.Command, storeapi store.Store, arg []string) (*Binary, error) {
	binary := &Binary{
		APICommand: cmd,
		Arg:        arg,
		Store:      storeapi,
	}
	return binary, nil
}

func (b *Binary) download() (*store.Command, error) {
	return b.Store.GetCommand()
}

func (b *Binary) commandPath() string {
	return fmt.Sprintf("/opt/sd/commands/%s/%s", b.APICommand.Namespace, b.APICommand.Command)
}

func (b *Binary) install(cmd *store.Command) (string, error) {
	dirPath := b.commandPath()

	if err := os.MkdirAll(dirPath, 0777); err != nil {
		return "", fmt.Errorf("Failed to create command directory: %v", err)
	}
	path := fmt.Sprintf("%s/%s", dirPath, cmd.Meta.Version)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("Failed to create command file: %v", err)
	}
	defer file.Close()
	_, err = file.Write(cmd.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to write command file: %v", err)
	}
	if err := os.Chmod(path, 0777); err != nil {
		return "", fmt.Errorf("Failed to change the access permissions of command file: %v", err)
	}
	return path, nil
}

// Run executes command and returns output
func (b *Binary) Run() ([]byte, error) {
	logger.Write("Download binary from store API")
	command, err := b.download()
	if err != nil {
		return nil, err
	}

	logger.Write("Install binary to this repository")
	path, err := b.install(command)
	if err != nil {
		return nil, err
	}

	logger.Write("Execute the binary")
	result, err := exec.Command(path, b.Arg...).Output()
	if err != nil {
		return nil, fmt.Errorf("Failed to execute Command: %v", err)
	}
	return result, nil
}
