package promoter

import (
	"fmt"
	"strings"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Promoter is a type to tag commands
type Promoter struct {
	smallSpec     *util.CommandSpec
	sdAPI         api.API
	targetVersion string
	tag           string
}

// New generates new Promoter.
// args is expected as ["namespace/name", "targetVersion", "tag"]
func New(api api.API, args []string) (p *Promoter, err error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("parameters are not enough")
	}
	targetVersion, tag := args[1], args[2]

	commandName := strings.Split(args[0], "/") // namespace/name
	if len(commandName) != 2 {
		return nil, fmt.Errorf("%v is invalid command name", args[0])
	}

	smallSpec := &util.CommandSpec{
		Namespace: commandName[0],
		Name:      commandName[1],
		Version:   tag,
	}

	p = &Promoter{
		smallSpec:     smallSpec,
		sdAPI:         api,
		targetVersion: targetVersion,
		tag:           tag,
	}

	return
}

// Run executes tag command API
func (p *Promoter) Run() (err error) {
	spec, err := p.sdAPI.GetCommand(p.smallSpec)
	if err != nil {
		fmt.Printf("%v does not exist yet\n", p.tag)
	} else if spec.Version == p.targetVersion {
		fmt.Printf("%v has been already tagged with %v\n", spec.Version, p.tag)
		return
	} else {
		fmt.Printf("Removing %v from %v\n", spec.Version, p.tag)
	}
	res, err := p.sdAPI.TagCommand(p.smallSpec, p.targetVersion, p.tag)
	if err != nil {
		fmt.Println("Promoting is aborted")
		return
	}

	fmt.Printf("Promoting %v to %v\n", res.Version, res.Tag)
	return
}
