package config

import (
	"os"
	"testing"
)

var (
	cacheEnv map[string]string
)

const (
	dummyEnvData = "dummy-env"
)

func setEnv(key, value string) {
	if cacheEnv == nil {
		cacheEnv = make(map[string]string)
	}
	cacheEnv[key] = os.Getenv(key)
	os.Setenv(key, value)
}

func setup() {
	setEnv("SD_API_URL", dummyEnvData)
}

func teardown() {
	for key, val := range cacheEnv {
		os.Setenv(key, val)
	}
}

func TestLoadConfig(t *testing.T) {
	LoadConfig()
	if SDAPIURL != dummyEnvData {
		t.Errorf("SDAPIURL=%q, want %q", SDAPIURL, dummyEnvData)
	}
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}
