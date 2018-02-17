package publisher

import (
	"fmt"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

func init() {
	config.LoadConfig()
}

type Publisher struct {
	inputCommand map[string]string
	commandSpec  *util.CommandSpec
}

func (p *Publisher) Run() error {
	sdAPI := api.New(config.SDAPIURL, config.SDToken)
	err := sdAPI.PostCommand(p.commandSpec)
	if err != nil {
		return fmt.Errorf("Post failed:%v", err)
	}

	// Published successfully
	println(p.commandSpec.Version)
	return nil
}

func New(inputCommand []string) (*Publisher, error) {
	var p Publisher
	var err error
	p.inputCommand, err = util.ParseCommand(inputCommand)
	if err != nil {
		return nil, fmt.Errorf("Command parse fail:%v1", err)
	}

	cs, err := util.LoadYml(p.inputCommand["ymlPath"])
	if err != nil {
		return nil, fmt.Errorf("Yaml load failed:%v", err)
	}

	p.commandSpec = cs
	return &p, nil
}
