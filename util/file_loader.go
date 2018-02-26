package util

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// LoadByte receives path of a file.
// It returns the file content as byte array.
func LoadByte(filePath string) ([]byte, error) {
	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Fail to read file:%q", err)
	}
	return dat, nil
}

// LoadYaml receives path for command spec.
// It returns CommandSpec struct if yaml is valid.
func LoadYaml(yamlPath string) (*CommandSpec, error) {
	data, err := LoadByte(yamlPath)
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
