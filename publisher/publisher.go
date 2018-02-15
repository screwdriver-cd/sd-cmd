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
	commandSpec  []byte
}

func (p *Publisher) Run() {
	sdAPI := api.New(config.SDAPIURL, config.SDToken)
	sdAPI.PostCommand(p.commandSpec)
}

func New(inputCommand []string) (*Publisher, error) {
	var p Publisher
	var err error
	p.inputCommand, err = util.ParseCommand(inputCommand)
	if err != nil {
		return nil, fmt.Errorf("Command parse fail:%q", err)
	}
	cs, err := util.LoadYml(p.inputCommand["ymlPath"])
	if err != nil {
		return nil, fmt.Errorf("Yaml load failed:%q", err)
	}
	p.commandSpec = util.CommandSpecToJsonBytes(*cs)
	return &p, nil
}
