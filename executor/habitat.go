package executor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/store"
	"github.com/screwdriver-cd/sd-cmd/util"
)

const habPath = "/hab/bin/hab"

// Habitat is the Habitat Executor struct
type Habitat struct {
	Args  []string
	Spec  *util.CommandSpec
	Store store.Store
}

// NewHabitat returns the Habitat struct
func NewHabitat(spec *util.CommandSpec, args []string) (habitat *Habitat, err error) {
	storeapi := store.New(config.SDStoreURL, spec, config.SDToken)

	habitat = &Habitat{
		Args:  args,
		Spec:  spec,
		Store: storeapi,
	}
	return habitat, nil
}

func (h *Habitat) getPkgDirPath() string {
	return filepath.Join(config.BaseCommandPath, h.Spec.Namespace, h.Spec.Name, h.Spec.Version)
}

func (h *Habitat) getPkgFilePath() string {
	fileName := filepath.Base(h.Spec.Habitat.File)
	return filepath.Join(h.getPkgDirPath(), fileName)
}

func (h *Habitat) isDownloaded() bool {
	fInfo, err := os.Stat(h.getPkgFilePath())
	if err != nil && !os.IsExist(err) {
		return false
	}
	if fInfo.Size() == 0 {
		return false
	}
	return true
}

func (h *Habitat) download() error {
	cmd, err := h.Store.GetCommand()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(h.getPkgDirPath(), 0777); err != nil {
		return fmt.Errorf("Failed to create command directory: %v", err)
	}

	path := h.getPkgFilePath()
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Failed to create command file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(cmd.Body)
	if err != nil {
		return fmt.Errorf("Failed to write command file: %v", err)
	}
	if err := os.Chmod(path, 0777); err != nil {
		return fmt.Errorf("Failed to change the access permissions of command file: %v", err)
	}
	return nil
}

// install executes "hab install" with a package name
func (h *Habitat) install() (err error) {
	var installPkg string
	if h.Spec.Habitat.Mode == "local" {
		installPkg = h.Spec.Habitat.File
	} else {
		installPkg = h.Spec.Habitat.Package
	}
	pkgInstallArgs := []string{"pkg", "install", installPkg}

	return execCommand(habPath, pkgInstallArgs)
}

// exec executes "hab exec" with a package name, command and args from a CLI
func (h *Habitat) exec() (err error) {
	execArgs := append([]string{"pkg", "exec", h.Spec.Habitat.Package, h.Spec.Habitat.Command}, h.Args...)

	return execCommand(habPath, execArgs)
}

// Run executes "hab install" and "hab exec"
func (h *Habitat) Run() (err error) {
	lgr.Debug.Println("start installing habitat command.")

	if h.Spec.Habitat.Mode == "local" && h.isDownloaded() == false {
		lgr.Debug.Println("start downloading local mode habitat package.")

		err = h.download()
		if err != nil {
			lgr.Debug.Println(err)
			return
		}
	}

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
