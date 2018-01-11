package executor

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/screwdriver-cd/sd-cmd/config"
)

func TestStartLog(t *testing.T) {
	defer os.RemoveAll(config.SDArtifactsDir)

	// success
	err := StartLog([]string{"exec", "namespace/name@version"})
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// check there is a log directory
	dirPath := filepath.Join(config.SDArtifactsDir, ".sd", "commands", "namespace", "name", "version")
	_, err = os.Stat(dirPath)
	if err != nil {
		t.Errorf("err=%q, want nil", err)
	}

	// check there is a file
	fileInfos, err := ioutil.ReadDir(dirPath)
	if len(fileInfos) != 1 {
		t.Errorf("filenum=%q want 1", len(fileInfos))
	}

	// check can logging
	log.Println("Hello this is log")
	for _, fileInfo := range fileInfos {
		body, _ := ioutil.ReadFile(filepath.Join(dirPath, fileInfo.Name()))
		if !strings.Contains(string(body), "Hello this is log") {
			t.Errorf("log body is %q. It should include %q", string(body), "Hello this is log")
		}
	}

	// failure
	err = StartLog([]string{"exec", "namespace:name@version"})
	if err == nil {
		t.Errorf("err=nil, want error")
	}
}
