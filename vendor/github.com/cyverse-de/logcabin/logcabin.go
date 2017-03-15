// Package logcabin provides common logging functionality that can be used to
// emit messages in the JSON format that we use for logstash/kibana.
package logcabin

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var (
	// Trace is the logger for the trace log level.
	Trace *log.Logger

	// Info is the logger for the info log level.
	Info *log.Logger

	// Warning is the logger for the warning log level.
	Warning *log.Logger

	// Error is the logger for the error log level.
	Error *log.Logger

	traceLincoln   *Lincoln
	infoLincoln    *Lincoln
	warningLincoln *Lincoln
	errorLincoln   *Lincoln
)

// Log Level Constants
const (
	traceLevel = "TRACE"
	infoLevel  = "INFO"
	warnLevel  = "WARN"
	errorLevel = "ERR"
)

// Init initializes the loggers.
func Init(service, artifact string) {
	traceLincoln = &Lincoln{service, artifact, traceLevel}
	infoLincoln = &Lincoln{service, artifact, infoLevel}
	warningLincoln = &Lincoln{service, artifact, warnLevel}
	errorLincoln = &Lincoln{service, artifact, errorLevel}

	Trace = log.New(traceLincoln, "", log.Lshortfile)
	Info = log.New(infoLincoln, "", log.Lshortfile)
	Warning = log.New(warningLincoln, "", log.Lshortfile)
	Error = log.New(errorLincoln, "", log.Lshortfile)
}

// LogMessage represents a message that will be logged in JSON format.
type logMessage struct {
	Service  string `json:"service"`
	Artifact string `json:"art-id"`
	Group    string `json:"group-id"`
	Level    string `json:"level"`
	Time     int64  `json:"timeMillis"`
	Message  string `json:"message"`
}

// Lincoln is a logger for jex-events.
type Lincoln struct {
	service  string
	artifact string
	level    string
}

// NewLogMessage returns a pointer to a new instance of LogMessage.
func (l *Lincoln) newLogMessage(message string) *logMessage {
	lm := &logMessage{
		Service:  l.service,
		Artifact: l.artifact,
		Group:    "org.iplantc",
		Level:    l.level,
		Time:     time.Now().UnixNano() / int64(time.Millisecond),
		Message:  message,
	}
	return lm
}

func (l *Lincoln) Write(buf []byte) (n int, err error) {
	m := l.newLogMessage(string(buf[:]))
	j, err := json.Marshal(m)
	if err != nil {
		return 0, err
	}
	j = append(j, []byte("\n")...)
	return os.Stdout.Write(j)
}
