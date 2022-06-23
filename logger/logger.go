package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// LogLevel log level
type LogLevel int

const (
	// Silent silent log level
	Silent LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
	// Debug debug log level
	Debug
)

type Writer interface {
	Printf(string, ...interface{})
	SetLevel(level log.Level)
	WithFields(fields log.Fields) *log.Entry
}

type Config struct {
	SlowThreshold time.Duration
	LogLevel      LogLevel
}

type Interface interface {
	LogMode(level LogLevel) Interface
	Info(begin time.Time, args ...interface{})
	Error(begin time.Time, args ...interface{})
	Debug(begin time.Time, args ...interface{})
	Trace(ctx context.Context, begin time.Time, fc func() (string, int), err error)
	Init(fields map[string]interface{})
	SetField(key string, value interface{})
}

var (
	Default = New(Formatter(os.Stdout), Config{
		SlowThreshold: 5 * time.Second,
		LogLevel:      Warn,
	})

	// Recorder Recorder logger records running request into a recorder instance
	Recorder = traceRecorder{Interface: Default, BeginAt: time.Now()}
)

func Formatter(out io.Writer) *log.Logger {
	logger := log.StandardLogger()
	logger.Formatter = &log.TextFormatter{
		FullTimestamp: true,
		DisableQuote:  true,
	}
	logger.SetOutput(out)
	return logger
}

type logger struct {
	Writer
	Config
	fields log.Fields
}

func New(writer Writer, config Config) Interface {
	return &logger{
		Writer: writer,
		Config: config,
		fields: make(log.Fields),
	}
}

// LogMode log mode
func (l *logger) LogMode(level LogLevel) Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func fields(l log.Fields, begin time.Time) log.Fields {
	l["duration"] = fmt.Sprintf("%.3fms", float64(time.Since(begin).Nanoseconds())/1e6)
	return l
}

func (l *logger) Init(fields map[string]interface{}) {
	for key, value := range fields {
		l.fields[key] = value
	}
}

func (l *logger) Info(begin time.Time, args ...interface{}) {
	if l.LogLevel >= Info {
		l.SetLevel(log.InfoLevel)
		l.WithFields(fields(l.fields, begin)).Info(args...)
	}
}

func (l *logger) Error(begin time.Time, args ...interface{}) {
	if l.LogLevel >= Error {
		l.SetLevel(log.ErrorLevel)
		l.WithFields(fields(l.fields, begin)).Debug(args...)
	}
}

func (l *logger) Debug(begin time.Time, args ...interface{}) {
	if l.LogLevel >= Debug {
		l.SetLevel(log.DebugLevel)
		l.WithFields(fields(l.fields, begin)).Debug(args...)
	}
}

func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int), err error) {
	if l.LogLevel <= Silent {
		return
	}
	elapsed := time.Since(begin)

	_, statusCode := fc()
	l.fields["duration"] = fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)
	l.fields["status_code"] = statusCode
	switch {
	case err != nil && l.LogLevel >= Error:
		l.WithFields(l.fields).Error(err)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		showLog := fmt.Sprintf("SLOW REQUEST >= %v", l.SlowThreshold)
		l.WithFields(l.fields).Warn(showLog)
	case l.LogLevel == Info:
		l.WithFields(l.fields).Info()
	}
}

func (k *logger) SetField(key string, value interface{}) {
	k.fields[key] = value
}

type traceRecorder struct {
	Interface
	BeginAt time.Time
	Err     error
}

// New new trace recorder
func (l traceRecorder) New() *traceRecorder {
	return &traceRecorder{Interface: l.Interface, BeginAt: time.Now()}
}

// Trace implement logger interface
func (l *traceRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int), err error) {
	l.BeginAt = begin
	l.Err = err
}
