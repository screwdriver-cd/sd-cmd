package executor

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/util"
)

var logFile *os.File

// StartLog make the log output to the file and stderr
func StartLog(args []string) error {
	ns, name, ver, _, err := util.SplitCmdWithSearch(args)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%v.log", time.Now().Unix())
	dirPath := filepath.Join(config.SDArtifactsDir, ".sd", "commands", ns, name, ver)

	err = os.MkdirAll(dirPath, 0777)
	if err != nil {
		return fmt.Errorf("Failed to create logging directory: %v", err)
	}

	logFilePath := filepath.Join(dirPath, filename)
	logFile, err = os.Create(logFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create logging file: %v", err)
	}

	writer := io.MultiWriter(os.Stderr, logFile)
	log.SetOutput(writer)
	return nil
}

// FinishLog finish log safely
func FinishLog() {
	if logFile != nil {
		logFile.Close()
	}
}
