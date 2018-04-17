package executor

import (
	"github.com/screwdriver-cd/sd-cmd/util"
)

var habPath = "/hab/bin/hab"

// Habitat is the Habitat Executor struct
type Habitat struct {
	Args []string
	Spec *util.Habitat
}

// NewHabitat returns the Habitat struct
func NewHabitat(spec *util.CommandSpec, args []string) (habitat *Habitat, err error) {
	habitat = &Habitat{
		Args: args,
		Spec: &util.Habitat{
			Mode:    spec.Habitat.Mode,
			Package: spec.Habitat.Package,
			Command: spec.Habitat.Command,
		},
	}
	return habitat, nil
}

// install executes "hab install" with a package name
func (h *Habitat) install() (err error) {
	pkgInstallArgs := []string{"pkg", "install", h.Spec.Package}

	return execCommand(habPath, pkgInstallArgs)
}

// exec executes "hab exec" with a package name, command and args from a CLI
func (h *Habitat) exec() (err error) {
	execArgs := append([]string{"pkg", "exec", h.Spec.Package, h.Spec.Command}, h.Args...)

	return execCommand(habPath, execArgs)
}

// Run executes "hab install" and "hab exec"
func (h *Habitat) Run() (err error) {
	lgr.Debug.Println("start installing habitat command.")

	err = h.install()
	if err != nil {
		lgr.Debug.Println(err)
		return
	}

	lgr.Debug.Println("start executing habitat command.")
	err = h.exec()
	if err != nil {
		lgr.Debug.Println(err)
	} else {
		lgr.Debug.Println("execute habitat command succeeded.")
	}
	return
}
