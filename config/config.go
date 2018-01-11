package config

import (
	"fmt"
	"os"
)

var (
	// VERSION is this cli version
	VERSION = ""
	// SDAPIURL is SD_API_URL value
	SDAPIURL string
	// SDStoreURL is SD_STORE_URL value
	SDStoreURL string
	// SDToken is SD_TOKEN value
	SDToken string
	// SDArtifactsDir is SD_ARTIFACTS_DIR value
	SDArtifactsDir string
	// BaseCommandPath is path of installing binary command
	BaseCommandPath = "/opt/sd/commands/"
)

func addSlash(val string) string {
	if val == "" {
		return val
	}
	if val[len(val)-1] != '/' {
		return fmt.Sprintf("%s/", val)
	}
	return val
}

// LoadConfig sets config data
func LoadConfig() {
	SDAPIURL = addSlash(os.Getenv("SD_API_URL"))
	SDStoreURL = addSlash(os.Getenv("SD_STORE_URL"))
	SDToken = os.Getenv("SD_TOKEN")
	SDArtifactsDir = addSlash(os.Getenv("SD_ARTIFACTS_DIR"))
	if VERSION == "" {
		VERSION = "0.0.0"
	}
}
