package util

import (
	"encoding/json"
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

type commandSpec struct {
	Namespace   string `yaml:"namespace"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Format      string `yaml:"format"`
	Binary      struct {
		File string `yaml:"file"`
	}
}

func LoadYml(ymlPath string) (*CommandSpec, error) {
	data, err := loadFile(ymlPath)
	if err != nil {
		return nil, fmt.Errorf("Fail to load yaml:%q", err)
	}

	cs := CommandSpec{}
	err = yaml.Unmarshal(data, &cs)
	if err != nil {
		return nil, fmt.Errorf("Yaml parse failed %q", err)
	}

	return &cs, nil
}

func CommandSpecToJsonBytes(cs CommandSpec) []byte {
	d, _ := json.Marshal(&cs)
	print(string(d))
	return d
}
