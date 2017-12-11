package config

import "os"

var (
	// VERSION is this cli version
	VERSION = ""
	// SdAPIURL is SD_API_URL value
	SdAPIURL string
	// SdStoreURL is SD_STORE_URL value
	SdStoreURL string
	// SdAPIToken is SD_API_TOKEN value
	SdAPIToken string
)

// LoadConfig set config data
func LoadConfig() {
	SdAPIURL = os.Getenv("SD_API_URL")
	SdStoreURL = os.Getenv("SD_STORE_URL")
	SdAPIToken = os.Getenv("SD_API_TOKEN")
	if VERSION == "" {
		VERSION = "0.0.0"
	}
}
