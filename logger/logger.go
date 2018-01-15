// Package logger log data to Stderr and file.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var loggingFiles []io.WriteCloser

// Logger has information for logging
type Logger struct {
	File  io.WriteCloser
	Debug *log.Logger // Debug is for debug log. You can set log flag. Also you can choose log stderr or not
	Error *log.Logger // Error is always debug file and stderr with LstdFlags flag.
}

// New returns logger object
func New(dirPath, filename string, flag int, debug bool) (*Logger, error) {
	lagger := new(Logger)

	file, err := CreateLogFile(dirPath, filename)
	if err != nil {
		return nil, err
	}
	lagger.SetInfos(file, flag, debug)
	return lagger, nil
}

// CreateLogFile create log file
func CreateLogFile(dirPath, filename string) (io.WriteCloser, error) {
	if filename == "" {
		filename = fmt.Sprintf("%v.log", time.Now().Unix())
	}
	err := os.MkdirAll(dirPath, 0777)
	if err != nil {
		return nil, fmt.Errorf("Failed to create logging directory: %v", err)
	}

	filePath := filepath.Join(dirPath, filename)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("Failed to create logging file: %v", err)
	}
	loggingFiles = append(loggingFiles, file)
	return file, nil
}

// SetInfos set logger information from arguments
func (l *Logger) SetInfos(file io.WriteCloser, flag int, debug bool) {
	l.File = file
	if debug {
		l.Debug = log.New(io.MultiWriter(os.Stderr, file), "", flag)
	} else {
		l.Debug = log.New(file, "", flag)
	}
	l.Error = log.New(io.MultiWriter(os.Stderr, file), "", log.LstdFlags)
}

// Close finish log file safely
func (l *Logger) Close() {
	if l.File != nil {
		l.File.Close()
	}
}

// CloseAll close every file you use.
func CloseAll() {
	for _, f := range loggingFiles {
		if f != nil {
			f.Close()
		}
	}
}
