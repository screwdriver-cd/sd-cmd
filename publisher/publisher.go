package publisher

import (
	"flag"
	"fmt"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Publisher is a type to publish sdapi and sdstore.
// It receives strings which input by a user.
// If -f option is valid, yaml file will be loaded to commandSpec struct.
type Publisher struct {
	specPath    string
	commandSpec *util.CommandSpec
}

// Run is a method to publish sdapi and sdstore.
func (p *Publisher) Run() error {
	sdAPI := api.New(config.SDAPIURL, config.SDToken)
	specResponse, err := sdAPI.PostCommand(p.specPath, p.commandSpec)
	if err != nil {
		return fmt.Errorf("Post failed:%v", err)
	}

	// Published successfully
	// Show version number of command published by sd-cmd
	fmt.Println(specResponse.Version)

	return nil
}

// New is a method to Generate new Publisher.
// Publisher variable will be returned if input command and yaml file is valid.
func New(inputCommand []string) (p *Publisher, err error) {
	p = new(Publisher)

	p.specPath, err = parsePublishCommand(inputCommand)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse command:%v", err)
	}

	p.commandSpec, err = util.LoadYaml(p.specPath)
	if err != nil {
		return nil, fmt.Errorf("Yaml load failed:%v", err)
	}

	return
}

func parsePublishCommand(inputCommand []string) (string, error) {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	yamlPath := fs.String("f", "sd-command.yaml", "Path of yaml to publish")

	err := fs.Parse(inputCommand)
	if err != nil {
		return "", fmt.Errorf("Failed to parse input args:%v", err)
	}

	return *yamlPath, nil
}
