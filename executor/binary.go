package executor

import (
	"os/exec"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
)

// Binary is a Binary Executor object
type Binary struct {
	Spec *api.Command
	Arg  []string
}

// NewBinary returns Binary object
func NewBinary(spec *api.Command, arg []string) (*Binary, error) {
	binary := &Binary{
		Spec: spec,
		Arg:  arg,
	}
	return binary, nil
}

// Run executes command and returns output
func (b *Binary) Run() ([]byte, error) {
	return exec.Command("ls", b.Arg...).Output()
}
