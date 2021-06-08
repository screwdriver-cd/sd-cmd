// Package logger log data to Stderr and file.
package logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

// Logger has information for logging
type Logger struct {
	debugFlag int
	errorFlag int
	file      io.WriteCloser
	Debug     *log.Logger // Debug is for debug log. You can set log flag. Also you can choose log stderr or not
	Error     *log.Logger // Error is always debug file and stderr with LstdFlags flag.
}

// File return current log output file. If the file = nil, logger does not output log to file
func (l *Logger) File() io.WriteCloser {
	return l.file
}

// Close finish log file safely
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// New returns logger object
func New(file io.WriteCloser) (*Logger, error) {
	lgr := new(Logger)
	lgr.file = file
	lgr.errorFlag = log.LstdFlags
	lgr.debugFlag = log.LstdFlags

	lgr.buildLoggers()
	return lgr, nil
}

func (l *Logger) buildDebugLogger() {
	if l.file != nil {
		l.Debug = log.New(l.file, "", l.debugFlag)
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
