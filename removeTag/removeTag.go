package removeTag

import (
	"fmt"
	"strings"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// RemoveTag is a type to remove tags from commands
type RemoveTag struct {
	smallSpec *util.CommandSpec
	sdAPI     api.API
	tag       string
}

// New generates new removeTag.
// args is expected as ["namespace/name", "tag"]
func New(api api.API, args []string) (p *RemoveTag, err error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("parameters are not enough")
	}
	tag := args[1]

	commandName := strings.Split(args[0], "/") // namespace/name
	if len(commandName) != 2 {
		return nil, fmt.Errorf("%v is invalid command name", args[0])
	}

	if !util.ValidateTagName(tag) {
		return nil, fmt.Errorf("%v is invalid tag name", tag)
	}

	smallSpec := &util.CommandSpec{
		Namespace: commandName[0],
		Name:      commandName[1],
		Version:   tag,
	}

	p = &RemoveTag{
		smallSpec: smallSpec,
		sdAPI:     api,
		tag:       tag,
	}

	return
}

// Run executes tag command API
func (p *RemoveTag) Run() (err error) {
	_, err = p.sdAPI.GetCommand(p.smallSpec)
	if err != nil {
		fmt.Printf("%v does not exist yet\n", p.tag)
		err = nil
		return
	}

	res, err := p.sdAPI.RemoveTagCommand(p.smallSpec, p.tag)
	if err != nil {
		fmt.Println("Remove tag is aborted")
		return
	}

	fmt.Printf("Removing %v from %v\n", res.Tag, res.Version)
	return
}
