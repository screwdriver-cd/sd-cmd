package publisher

import (
	"github.com/sonic-screwdriver-cd/sd-cmd/config"
	"github.com/sonic-screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/sonic-screwdriver-cd/sd-cmd/util"
)

func init() {
	config.LoadConfig()
}

type Publisher struct {
	inputCommand map[string]string
	commandSpec  []byte
}

func (p *Publisher) Run() {
	sdAPI, _ := api.New(config.SDAPIURL, config.SDToken)
	sdAPI.PostCommand(p.commandSpec)
}

func New(inputCommand []string) Publisher {
	var p Publisher
	p.inputCommand = util.ParseCommand(inputCommand)
	cs := util.LoadYml(p.inputCommand["ymlPath"])
	p.commandSpec = util.CommandSpecToJsonBytes(cs)
	return p
}
