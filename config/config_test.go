package config

import (
	"os"
	"testing"
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
}

func teardown() {
	for key, val := range originalEnv {
		os.Setenv(key, val)
	}
}

func TestLoadConfig(t *testing.T) {
	LoadConfig()
	if SDAPIURL != dummyAPIURL {
		t.Errorf("SDAPIURL=%q, want %q", SDAPIURL, dummyAPIURL)
	}
	if SDToken != dummyToken {
		t.Errorf("SDToken=%q, want %q", SDToken, dummyToken)
	}
	if SDStoreURL != dummyStoreURL {
		t.Errorf("SDStoreURL=%q, want %q", SDStoreURL, dummyStoreURL)
	}
	if SDArtifactsDir != dummySDArtifactsDir {
		t.Errorf("SDArtifactsDir=%q, want %q", SDArtifactsDir, dummySDArtifactsDir)
	}
	wantBaseCommandPath := os.Getenv("SD_BASE_COMMAND_PATH")
	if wantBaseCommandPath == "" {
		wantBaseCommandPath = dummyCustomCmdPath
	}
	if BaseCommandPath != wantBaseCommandPath {
		t.Errorf("BaseCommandPath=%q, want %s", BaseCommandPath, wantBaseCommandPath)
	}

	// check unset env
	os.Unsetenv("SD_API_URL")
	LoadConfig()
	if SDAPIURL != "" {
		t.Errorf("SDAPIURL=%q, want blank", SDAPIURL)
	}

	// set SD_BASE_COMMAND_PATH
	setEnv("SD_BASE_COMMAND_PATH", dummyCustomCmdPath)
	LoadConfig()
	if BaseCommandPath != dummyCustomCmdPath {
		t.Errorf("BaseCommandPath=%q, want %q", BaseCommandPath, dummyCustomCmdPath)
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
