package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	originalEnv map[string]string
)

const (
	dummyAPIURL         = "dummy-api"
	dummyToken          = "dummy-token"
	dummyStoreURL       = "dummy-store/"
	dummySDArtifactsDir = "dummy/sd/Artifacts/"
	dummyCustomCmdPath  = "/opt/sd/commands/"
	dummyDebug          = "true"
)

func setEnv(key, value string) {
	if originalEnv == nil {
		originalEnv = make(map[string]string)
	}
	originalEnv[key] = os.Getenv(key)
	os.Setenv(key, value)
}

func setup() {
	setEnv("SD_API_URL", dummyAPIURL)
	setEnv("SD_TOKEN", dummyToken)
	setEnv("SD_STORE_URL", dummyStoreURL)
	setEnv("SD_ARTIFACTS_DIR", dummySDArtifactsDir)
	setEnv("SD_BASE_COMMAND_PATH", dummyCustomCmdPath)
	setEnv("SD_CMD_DEBUG_LOG", dummyDebug)
}

func teardown() {
	for key, val := range originalEnv {
		os.Setenv(key, val)
	}
}

func TestLoadConfig(t *testing.T) {
	LoadConfig()
	assert.Equal(t, dummyAPIURL, SDAPIURL)
	assert.Equal(t, dummyToken, SDToken)
	assert.Equal(t, dummyStoreURL, SDStoreURL)
	assert.Equal(t, dummySDArtifactsDir, SDArtifactsDir)
	assert.Equal(t, dummyCustomCmdPath, BaseCommandPath)
	wantDebug, _ := strconv.ParseBool(dummyDebug)
	assert.Equal(t, wantDebug, DEBUG)

	// check unset env
	os.Unsetenv("SD_API_URL")
	os.Unsetenv("SD_CMD_DEBUG_LOG")
	LoadConfig()
	assert.Equal(t, "", SDAPIURL)
	assert.Equal(t, false, DEBUG)
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
