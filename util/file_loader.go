package util

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func loadFile(filePath string) []byte {
	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return dat
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

func LoadYml(ymlPath string) commandSpec {
	data := loadFile(ymlPath)

	cs := commandSpec{}
	err := yaml.UnmarshalStrict(data, &cs)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return cs
}

func CommandSpecToJsonBytes(cs commandSpec) []byte {
	d, _ := json.Marshal(&cs)
	print(string(d))
	return d
}
