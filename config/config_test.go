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
	if SDAPIURL != "dummy-api/" {
		t.Errorf("SDAPIURL=%q, want %q", SDAPIURL, "dummy-api/")
	}
	if SDToken != dummyToken {
		t.Errorf("SDAPIURL=%q, want %q", SDToken, dummyToken)
	}
	if SDStoreURL != dummyStoreURL {
		t.Errorf("SDAPIURL=%q, want %q", SDStoreURL, dummyStoreURL)
	}
	if SDArtifactsDir != dummySDArtifactsDir {
		t.Errorf("SDAPIURL=%q, want %q", SDArtifactsDir, dummySDArtifactsDir)
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
