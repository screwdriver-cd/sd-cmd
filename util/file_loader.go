package util

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

func loadFile(filePath string) ([]byte, error) {
	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Fail to read file:%q", err)
	}
	return dat, nil
}

// LoadYml receives path for command spec.
// It returns CommandSpec struct if yml is valid.
func LoadYml(ymlPath string) (*CommandSpec, error) {
	data, err := loadFile(ymlPath)
	if err != nil {
		return nil, fmt.Errorf("Fail to load yaml:%v", err)
	}

	cs := CommandSpec{}
	err = yaml.Unmarshal(data, &cs)
	if err != nil {
		return nil, fmt.Errorf("Yaml parse failed %v", err)
	}

	return &cs, nil
}
