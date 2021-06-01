// Package logger log data to Stderr and file.
package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogOption enable to customize Logger
type LogOption func(l *Logger) error

var loggingFiles []io.WriteCloser

// Logger has information for logging
type Logger struct {
	debugFlag         int
	errorFlag         int
	hasOutputDebugLog bool
	file              io.WriteCloser
	Debug             *log.Logger // Debug is for debug log. You can set log flag. Also you can choose log stderr or not
	Error             *log.Logger // Error is always debug file and stderr with LstdFlags flag.
}

// DebugFlag return current log debugFlag status
func (l *Logger) DebugFlag() int {
	return l.debugFlag
}

// ErrorFlag return current log errorFlag status
func (l *Logger) ErrorFlag() int {
	return l.errorFlag
}

// File return current log output file. If the file = nil, logger does not output log to file
func (l *Logger) File() io.WriteCloser {
	return l.file
}

// OptOutputToFileWithCreate create file for output log
func OptOutputToFileWithCreate(dir, filename string) LogOption {
	return func(l *Logger) error {
		file, err := createLogFile(dir, filename)
		if err != nil {
			return err
		}
		l.file = file
		return nil
	}
}

// OptOutputToFile output log to the file
func OptOutputToFile(file io.WriteCloser) LogOption {
	return func(l *Logger) error {
		l.file = file
		return nil
	}
}

// OptDebugFlag set Logger.Debug flag
func OptDebugFlag(flag int) LogOption {
	return func(l *Logger) error {
		l.debugFlag = flag
		return nil
	}
}

// OptErrorFlag set Logger.Error flag
func OptErrorFlag(flag int) LogOption {
	return func(l *Logger) error {
		l.errorFlag = flag
		return nil
	}
}

// OptOutputDebugLog output debug log
func OptOutputDebugLog(output bool) LogOption {
	return func(l *Logger) error {
		l.hasOutputDebugLog = output
		return nil
	}
}

// New returns logger object
func New(options ...LogOption) (*Logger, error) {
	lgr := new(Logger)
	lgr.errorFlag = log.LstdFlags
	lgr.debugFlag = log.LstdFlags

	for _, o := range options {
		err := o(lgr)
		if err != nil {
			return nil, err
		}
	}

	lgr.buildLoggers()
	return lgr, nil
}

func createLogFile(dirPath, filename string) (io.WriteCloser, error) {
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

func (l *Logger) buildDebugLogger() {
	if l.hasOutputDebugLog {
		if l.file != nil {
			l.Debug = log.New(io.MultiWriter(os.Stderr, l.file), "", l.debugFlag)
			return
		}
		l.Debug = log.New(os.Stderr, "", l.debugFlag)
		return
	}
	l.Debug = log.New(ioutil.Discard, "", l.debugFlag)
}

func (l *Logger) buildErrorLogger() {
	if l.file != nil {
		l.Error = log.New(io.MultiWriter(os.Stderr, l.file), "", l.errorFlag)
		return
	}
	l.Error = log.New(os.Stderr, "", l.errorFlag)
}

func (l *Logger) buildLoggers() {
	l.buildDebugLogger()
	l.buildErrorLogger()
}

// Close finish log file safely
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
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
