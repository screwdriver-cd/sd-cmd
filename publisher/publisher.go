package publisher

import (
	"flag"
	"fmt"
	"path"

	"github.com/screwdriver-cd/sd-cmd/promoter"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Publisher is a type to publish sdapi and sdstore.
// It receives strings which input by a user.
// If -f option is valid, yaml file will be loaded to commandSpec struct.
type Publisher struct {
	specPath    string
	commandSpec *util.CommandSpec
	sdAPI       api.API
	tag         string
}

func (p *Publisher) tagCommand(specResponse *util.CommandSpec) error {
	commandFullName := path.Join(specResponse.Namespace, specResponse.Name)
	promoter, err := promoter.New(p.sdAPI, []string{commandFullName, specResponse.Version, p.tag})
	if err != nil {
		return err
	}

	return promoter.Run()
}

// Run is a method to publish sdapi and sdstore.
func (p *Publisher) Run() error {
	specResponse, err := p.sdAPI.PostCommand(p.commandSpec)
	if err != nil {
		return fmt.Errorf("Post failed: %v", err)
	}

	err = p.tagCommand(specResponse)
	if err != nil {
		return fmt.Errorf("Tag failed: %v", err)
	}

	// Published successfully
	// Show version number of command published by sd-cmd
	fmt.Println(specResponse.Version)

	return nil
}

// New is a method to Generate new Publisher.
// Publisher variable will be returned if input command and yaml file is valid.
func New(api api.API, inputCommand []string) (p *Publisher, err error) {
	p = new(Publisher)

	p.sdAPI = api
	p.specPath, p.tag, err = parsePublishCommand(inputCommand)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse command:%v", err)
	}

	p.commandSpec, err = util.LoadYaml(p.specPath)
	if err != nil {
		return nil, fmt.Errorf("Yaml load failed:%v", err)
	}
	p.commandSpec.SpecPath = p.specPath

	return
}

func parsePublishCommand(inputCommand []string) (yamlPath, tag string, err error) {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	yamlPathAddr := fs.String("f", "sd-command.yaml", "Path of yaml to publish")
	tagAddr := fs.String("t", "latest", "Tag name for your command")

	err = fs.Parse(inputCommand)
	if err != nil {
		return "", "", fmt.Errorf("Failed to parse input args:%v", err)
	}

	return *yamlPathAddr, *tagAddr, err
}
