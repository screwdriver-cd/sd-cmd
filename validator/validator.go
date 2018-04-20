package validator

import (
	"flag"
	"fmt"

	"github.com/screwdriver-cd/sd-cmd/screwdriver/api"
	"github.com/screwdriver-cd/sd-cmd/util"
)

// Validator is a type to validate yaml.
// It receives strings which input by a user.
type Validator struct {
	cmdPath    string
	yamlString string
	sdAPI      api.API
}

// Run is a method to validate yaml.
func (v *Validator) Run() error {
	validateResponse, err := v.sdAPI.ValidateCommand(v.yamlString)
	if err != nil {
		return fmt.Errorf("Post failed:%v", err)
	}

	if len(validateResponse.Errors) != 0 {
		errorMessage := ""
		for _, error := range validateResponse.Errors {
			errorMessage += error.Message + "\n"
		}
		return fmt.Errorf("Command is not valid for the following reasons:\n%v", errorMessage)
	}

	fmt.Println("Validation completed successfully.")

	return nil
}

// New is a method to Generate new Validator.
// Validator variable will be returned if input command is valid.
func New(api api.API, inputCommand []string) (v *Validator, err error) {
	v = new(Validator)

	v.sdAPI = api
	v.cmdPath, err = parseValidateCommand(inputCommand)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse command:%v", err)
	}

	v.yamlString, err = util.LoadString(v.cmdPath)
	if err != nil {
		return nil, fmt.Errorf("Yaml load failed:%v", err)
	}

	return
}

func parseValidateCommand(inputCommand []string) (string, error) {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	yamlPath := fs.String("f", "sd-command.yaml", "Path of yaml to validate")

	err := fs.Parse(inputCommand)
	if err != nil {
		return "", fmt.Errorf("Failed to parse input args:%v", err)
	}

	return *yamlPath, nil
}
