package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/screwdriver-cd/sd-cmd/config"
	"github.com/screwdriver-cd/sd-cmd/util"
)

var logger *Logger

// Logger has information for logging
type Logger struct {
	Path   string
	file   io.WriteCloser
	logger *log.Logger
}

func init() {
	logger = &Logger{
		Path:   "",
		file:   nil,
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

// MakeLogToFile prepare for logging
func MakeLogToFile(fullCommands []string) error {
	ns, cmd, ver, _, err := util.SplitCmdWithSearch(fullCommands)
	if err != nil {
		return err
	}
	dir := fmt.Sprintf("%s/.sd/commands/%s/%s/%s/", config.SDArtifactsDir, ns, cmd, ver)
	filename := fmt.Sprintf("%v.log", time.Now().Unix())
	logger.Path = fmt.Sprintf("%s%s", dir, filename)

	err = os.MkdirAll(dir, 0777)
	if err != nil {
		return fmt.Errorf("Failed to create logging directory: %v", err)
	}

	file, err := os.Create(logger.Path)
	if err != nil {
		return fmt.Errorf("Failed to create logging file: %v", err)
	}

	logger.file = file
	writer := io.MultiWriter(os.Stderr, logger.file)
	logger.logger = log.New(writer, "", log.LstdFlags)
	return nil
}

// Write output log data
func Write(v ...interface{}) {
	logger.logger.Output(2, fmt.Sprintln(v...))
}

// Close finish logger safely
func Close() {
	logger.file.Close()
}
