package logger

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

type Level int8

const (
	LevelHTTP Level = iota
	LevelInfo
	LevelError
	LevelFatal
	LevelOff
)

func (l Level) String() string {
	if l == LevelHTTP {
		return "HTTP"
	}

	if l == LevelInfo {
		return "INFO"
	}

	if l == LevelError {
		return "ERROR"
	}

	if l == LevelFatal {
		return "FATAL"
	}

	return ""
}

/*
 * out - for log destination
 * minLevel - for provided log level
 * mutext - for coordinating the writes
 */
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// Initiate a new logger instance
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	if level < l.minLevel {
		return 0, nil
	}

	// Declare anonymous struct holding the data for the log entry
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// Include a stack trace for entries at the ERROR and FATAL fields
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// Declare a line variable for holding the actual log entry
	var line []byte

	// Marshal the anonymous struct to JSON and store it in the line variable
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	// Lock the mutex so that no two writes to the output destination can happen
	// concurrently. If this is not present it's possible that the text for two
	// or more log entries will be intermingled in the output
	// in a nutshell mutex is to prevent race condition
	l.mu.Lock()
	defer l.mu.Unlock()

	// Write the log entry followed by a newline
	return l.out.Write(append(line, '\n'))
}

func (l *Logger) PrintHTTP(request http.Request, response *http.Response) {
	responseStatus := map[string]string{"Code": strconv.Itoa(response.StatusCode)}
	l.print(LevelHTTP, request.Method, responseStatus)
}

func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1)
}

// We also implement a Write() method on our Logger type so that it satisfies
// io.Writer interface. This writes a log entry at the ERROR level with no additional
// properties
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
