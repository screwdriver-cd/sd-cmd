package config

import (
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

// LoadConfig sets config data
func LoadConfig() {
	SDAPIURL = os.Getenv("SD_API_URL")
	SDStoreURL = os.Getenv("SD_STORE_URL")
	SDToken = os.Getenv("SD_TOKEN")
	SDArtifactsDir = os.Getenv("SD_ARTIFACTS_DIR")
	if VERSION == "" {
		VERSION = "0.0.0"
	}
	if len(os.Getenv("SD_BASE_COMMAND_PATH")) != 0 {
		BaseCommandPath = os.Getenv("SD_BASE_COMMAND_PATH")
	}
}
