package config

import (
	"os"
	"testing"
)

var (
	cacheEnv map[string]string
)

const (
	dummyAPIURL   = "dummy-api"
	dummyToken    = "dummy-token"
	dummyStoreURL = "dummy-store/"
)

func setEnv(key, value string) {
	if cacheEnv == nil {
		cacheEnv = make(map[string]string)
	}
	cacheEnv[key] = os.Getenv(key)
	os.Setenv(key, value)
}

func setup() {
	setEnv("SD_API_URL", dummyAPIURL)
	setEnv("SD_TOKEN", dummyToken)
	setEnv("SD_STORE_URL", dummyStoreURL)
}

func teardown() {
	for key, val := range cacheEnv {
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
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
