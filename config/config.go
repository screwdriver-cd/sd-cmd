package config

import "os"

var (
	// VERSION is this cli version
	VERSION = ""
	// SDAPIURL is SD_API_URL value
	SDAPIURL string
	// SDStoreURL is SD_STORE_URL value
	SDStoreURL string
	// SDToken is SD_TOKEN value
	SDToken string
)

// LoadConfig sets config data
func LoadConfig() {
	SDAPIURL = os.Getenv("SD_API_URL")
	SDStoreURL = os.Getenv("SD_STORE_URL")
	SDToken = os.Getenv("SD_TOKEN")
	if VERSION == "" {
		VERSION = "0.0.0"
	}
}
